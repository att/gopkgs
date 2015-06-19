// vi: sw=4 ts=4:

/*
------------------------------------------------------------------------------------------------
	Mnemonic:	ostack_user
	Abstract:	Interface functions providing user information such as lists of roles.
	Date:		30 October 2014
	Authors:	E. Scott Daniels, Matti Hiltnuen, Kaustubh Joshi

	Mods:
				10 Nov 2014 : Added checks to ensure response data structs aren't nil
				13 Apr 2015 : Converted to more generic error structure use.
				19 Jun 2015 : Corrected bug in URL for roll gathering.
------------------------------------------------------------------------------------------------
*/

package ostack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)


// -------------------------------------------------------------------------------------------------

/*
	Builds  a map of the role names to ID strings.
*/
func (o *Ostack) Map_roles( ) ( rmap map[string]*string, err error ) {
	var (
		role_data	generic_response
		rjson		string = ""
	)

	rmap = nil

	if o == nil ||  o.user == nil || o.passwd == nil {
		err = fmt.Errorf( "no openstack object to work on, or missing data inside" )
		return
	}

	if o.iahost == nil {
		err = fmt.Errorf( "no identity url associated with ostack data object: %s", o.To_str() )
		return
	}

	body := bytes.NewBufferString( rjson );

	url := *o.iahost + "v2.0/OS-KSADM/roles";
	dump_url( "map-roles", 10, url )
	jdata, _, err := o.Send_req( "GET",  &url, body ); 
	dump_json( "map-roles", 10, jdata )

	if err != nil {
		return;
	}

	err = json.Unmarshal( jdata, &role_data )			// unpack the json into generic struct
	if err != nil {
		fmt.Fprintf( os.Stderr, "ostack: map-roles: unable to unpack json: %s\n", err );
		return;
	}

	if role_data.Error != nil {
		o.token = nil
		o.small_tok = nil
		err = fmt.Errorf( "map-role failed: %s\n", role_data.Error )
		return
	}

	if role_data.Roles != nil {
		if l := len( role_data.Roles ); l > 0 {
			rmap = make( map[string]*string, l )
	
			for _, v := range role_data.Roles {
				dup_str := v.Id							// must duplicate to pull from receive buffer
				rmap[v.Name] = &dup_str
			}
		}
	}

	return;
}


/*
	Map the roles assigned to the user and given project (must be id). If pid is nil then
	the project ID associated with the struct (returned on auth) will be used. 
*/
func (o *Ostack) Map_user_roles( pid *string ) ( rmap map[string]*string, err error ) {
	var (
		usr_data	generic_response
		rjson		string = ""
	)

	rmap = nil

	if o == nil ||  o.user == nil || o.passwd == nil {
		err = fmt.Errorf( "no openstack object to work on, or missing data inside" )
		return
	}
	if o.iahost == nil {
		err = fmt.Errorf( "no identity url associated with ostack data object: %s", o.To_str() )
		return
	}

	body := bytes.NewBufferString( rjson );

	if pid == nil {
		pid = o.project_id			// this project if not given by caller
	} 

	url := fmt.Sprintf( "%s/v2.0/tenants/%s/users/%s/roles", *o.iahost, *pid, *o.user_id )
	dump_url( "map-uroles", 10, url )
	jdata, _, err := o.Send_req( "GET",  &url, body ); 
	dump_json( "map-uroles", 10, jdata )

	if err != nil {
		return;
	}

	err = json.Unmarshal( jdata, &usr_data )			// unpack the json into usr_data
	if err != nil {
		fmt.Fprintf( os.Stderr, "ostack: map-usr-roles: unable to unpack json: %s\n", err );
		return;
	}

	if usr_data.Error != nil {
		o.token = nil
		o.small_tok = nil
		err = fmt.Errorf( "map-usr-roles failed: %s\n", usr_data.Error )
		return
	}

	if usr_data.Roles != nil {
		if l := len( usr_data.Roles ); l > 0 {
			rmap = make( map[string]*string, l )
	
			for _, v := range usr_data.Roles {
				rmap[v.Name] = &v.Id
			}
		}
	}

	return
}

/*
	Map the global roles assigned to the user associated with the structure.
	This may not always work -- depends on flavour of openstack it seems --
	and if it fails it seems to affect subsequent calls to roles (huh?).
*/
func (o *Ostack) Map_user_groles( ) ( rmap map[string]*string, err error ) {
	var (
		usr_data	generic_response
		rjson		string = ""
	)

	rmap = nil

	if o == nil ||  o.user == nil || o.passwd == nil {
		err = fmt.Errorf( "no openstack object to work on, or missing data inside" )
		return
	}
	if o.iahost == nil {
		err = fmt.Errorf( "no identity url associated with ostack data object: %s", o.To_str() )
		return
	}

	body := bytes.NewBufferString( rjson );

	url := fmt.Sprintf( "%s/users/%s/roles", *o.iahost, *o.user_id )
	dump_url( "map-groles", 100, url )
	jdata, _, err := o.Send_req( "GET",  &url, body ); 
	dump_json( "map-groles", 10, jdata )

	if err != nil {
		return;
	}

	err = json.Unmarshal( jdata, &usr_data )			// unpack the json into usr_data
	if err != nil {
		fmt.Fprintf( os.Stderr, "ostack: map-usr-groles: unable to unpack json: %s\n", err );
		return;
	}

	if usr_data.Error != nil {
		o.token = nil
		o.small_tok = nil
		err = fmt.Errorf( "map-usr-groles failed: %s\n", usr_data.Error )
		return
	}

	if usr_data.Roles != nil {
		if l := len( usr_data.Roles ); l > 0 {
			rmap = make( map[string]*string, l )
	
			for _, v := range usr_data.Roles {
				rmap[v.Name] = &v.Id
			}
		}
	}

	return
}
