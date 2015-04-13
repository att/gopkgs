// vi: sw=4 ts=4:

/*
------------------------------------------------------------------------------------------------
	Mnemonic:	ostack_tenant
	Abstract:	Functions to support getting tenant information assuming our user name is
				an admin name.
				Doc: http://developer.openstack.org/api-ref-identity-v2.html
					

	Date:		06 June 2014
	Author:		E. Scott Daniels
	Mod:		07 Jun 2014 - Added cache to prevent trashing openstack on validation requests. 
				16 Jun 2014 - Added support for verifying admin privs on the token
				19 Jun 2014 - Corrected bug in the token validation that wasn't checking that the
					project name on the token matches the project in the struct.
				19 Aug 2014 - Now uses the date from the token as the token's expiration.
				28 Oct 2014 - Added support for identity requests as admin.
				09 Oct 2014 - Corrected bug causing core dump when access or token was missing
					from the openstack response.  (Valid_for_project)
				10 Nov 2014 - Now using md5 hash of token when attempting to validate the token.
				17 Nov 2014 - Added token to project function.
				25 Nov 2014 - Added better error checking for nil admin url.
				13 Apr 2015 - Converted to more generic error structure use.
------------------------------------------------------------------------------------------------
*/

package ostack

import (
	"bytes"
	"encoding/json"
	"fmt"
	//"os"
)

// ------------- structs that are used to unbundle the json auth response data -------------------


// ------------------------------------------------------------------------------

/*
	Run a list of roles and return true if the named role is in the list. 
	Caution: this seems to be dodgy as openstack never seems to flip the admin 
			flag on even when the ID is an admin. 
*/
func find_role( list []*ost_role, r string ) ( bool ) {
	for i := range list {
		if list[i].Name == r {
			return true
		}
	}

	return false
}

// ------------------------------------------------------------------------------

/*
	Uses keystone to map all of the known tenants, not just the one that the token
	has been assigned to.  If the identity admin host (iahost) isn't specificed when
	the ostack struct was created and authorised, then this function is just a passthrough
	to the map tenants function and will return just a list of tenants that the 
	user has the ability to see rather than a list of all tenants.
*/
func (o *Ostack) Map_all_tenants( ) ( name2id map[string]*string, id2name map[string]*string, err error ) {
	var (
		tenant_data	generic_response
	)

	name2id = nil
	id2name = nil
	err = nil;

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	if o.iahost == nil {						// may be available only in newer (icehouse+) versions
		name2id, id2name, err = o.Map_tenants( )		// just fish out what we can and send that back
		return
	}

	body := bytes.NewBufferString( "" )

	url := *o.iahost + "/tenants"					// version is built into the ihost string (ugg); must use admin for all
    dump_url( "all-tenants", 10, url )
    jdata, _, err := o.Send_req( "GET",  &url, body );
	dump_json( "tenants", 10, jdata )

	if err != nil {
		return
	}
	dump_json( "map-tenants", 10, jdata )

	err = json.Unmarshal( jdata, &tenant_data )			// unpack the json into  our structures
	if err != nil {
		dump_json(  fmt.Sprintf( "map-tenants: unpack err: %s\n", err ), 30, jdata )
		return
	}

	name2id = make( map[string]*string, len( tenant_data.Tenants ) )
	id2name = make( map[string]*string, len( tenant_data.Tenants ) )
	for k := range tenant_data.Tenants {
		t := tenant_data.Tenants[k]
		dup_id := t.Id							// copy out of object for reference
		dup_nm := t.Name
		name2id[t.Name] = &dup_id	
		id2name[t.Id] = &dup_nm
	}

	return
}

/*
	Requests information from openstack and build a list of tenant names that we have access to.
	The return is two maps: names to IDs and IDs to names.
*/
func (o *Ostack) Map_tenants( ) ( name2id map[string]*string, id2name map[string]*string, err error ) {
	var (
		tenant_data	generic_response
	)

	name2id = nil
	id2name = nil
	err = nil;

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	body := bytes.NewBufferString( "" )

	url := *o.host + "v2.0/tenants"					// be warned: the openstack doc is not consistant or clear and suggests another url for this
    dump_url( "map-tenants", 10, url )
    jdata, _, err := o.Send_req( "GET",  &url, body );
	dump_json( "tenants", 10, jdata )

	if err != nil {
		return
	}
	dump_json( "map-tenants", 10, jdata )

	err = json.Unmarshal( jdata, &tenant_data )			// unpack the json into  our structures
	if err != nil {
		dump_json(  fmt.Sprintf( "map-tenants: unpack err: %s\n", err ), 30, jdata )
		return
	}

	if tenant_data.Tenants != nil {
		name2id = make( map[string]*string, len( tenant_data.Tenants ) )
		id2name = make( map[string]*string, len( tenant_data.Tenants ) )
		for k := range tenant_data.Tenants {
			t := tenant_data.Tenants[k]
			dup_id := t.Id							// copy out of object for reference
			dup_nm := t.Name
			name2id[t.Name] = &dup_id	
			id2name[t.Id] = &dup_nm
		}
	}

	return
}


/*
	Given a token, return the name of the project that is associated with it, or 
	nil if the toekn isn't valid.  Returns error if there are issues getting 
	info from openstack. 
*/
func (o *Ostack) Token2project( token *string ) ( project *string, id *string, err error ) {
	var (
		response_data generic_response
		small_tok  *string = nil
	)

	if o == nil  || token == nil {
		return nil, nil, fmt.Errorf( "no openstack creds (nil)" )
	}

	if len( *token ) > 100 {					// probalby one of those absurdly huge tokens; take the md5
		small_tok = str2md5_str( *token )
	} else {
		small_tok = token
	}

	err = o.Validate_auth()				// reauthorise if needed
	if err != nil {
		return nil, nil, err
	}

	body := bytes.NewBufferString( "" )

	url := fmt.Sprintf( "%sv2.0/tokens/%s", *o.host, *small_tok )								// this returns token information if token is valid
	dump_url( "valid4project", 10, url )
    jdata, _, err := o.Send_req( "GET",  &url, body );

	if err != nil {
		return nil, nil, err
	}
	dump_json( "valid4project", 10, jdata )

	err = json.Unmarshal( jdata, &response_data )
	o.tok_isadmin[*small_tok] = false							// default to not an admin
	if err == nil {
		if response_data.Error == nil  {
			if response_data.Access == nil || response_data.Access.Token == nil  {
				if response_data.Error != nil {
					err = fmt.Errorf( "tok2project: missing ostack response: token=%s: %s\n", *token, response_data.Error )
				} else {
					err = fmt.Errorf( "tok2project: missing ostack response (tok=%s): code=unk msg=not-given\n", *token )
				}

				return nil, nil, err
			}

			if response_data.Access != nil && response_data.Access.Token.Tenant != nil { 		// verify that the project for the given token matches our project
				dup_nm := response_data.Access.Token.Tenant.Name								// snag the project name
				dup_id := response_data.Access.Token.Tenant.Id
					return &dup_nm, &dup_id, nil
			}

			err = fmt.Errorf( "project information not returned by openstack" )
		} else {
			err = fmt.Errorf( "unable to fetch project from token: %s:  %s\n", *token, response_data.Error )
		}
	}

	return nil, nil, err
}

/*
	Given a project and token, return true if the token is valid for the project.
	Returns with error set if there were issues gathering information from openstack.
*/
func (o *Ostack) Valid_for_project( token *string, project *string ) ( bool, error ) {
	p, _, err := o.Token2project( token )
	if p != nil  &&  *p == *project {
		return true, err
	}

	if err == nil {
		err = fmt.Errorf( "project did not match: token=%s, expected=%s", *p, *project )
	}

	return false, err
}

/*
	Given a project id and token, return true if the token is valid for the project.
	Returns with error set if there were issues gathering information from openstack.
*/
func (o *Ostack) Valid_for_projectid( token *string, projectid *string ) ( bool, error ) {
	_, id, err := o.Token2project( token )
	if id != nil  &&  *id == *projectid {
		return true, err
	}

	if err == nil {
		err = fmt.Errorf( "project did not match: token=%s, expected=%s", *id, *projectid )
	}

	return false, err
}

