// vi: sw=4 ts=4:

/*
------------------------------------------------------------------------------------------------
	Mnemonic:	ostack_auth
	Abstract:	Interface to ostack to do authorisation and authroisation (endpoint) related
				things.
	Date:		24 Octoberr 2013
	Authors:	E. Scott Daniels, Matti Hiltnuen, Kaustubh Joshi

	Mods:		23 Apr 2014 - Added support for tenant id.
				18 Jun 2014 - Added generic token validation.
				28 Jul 2014 - Changed tenant_id to project ID. 
				16 Aug 2014 - Now pick up exact token expiry time. 
				30 Sep 2014 - Added check to determin if 'admin' privs exist for the user.
				23 Oct 2014 - Added expire function call.
				28 Oct 2014 - Added support for identity requests as admin.
							Added endpoint function.
				10 Nov 2014 : Added checks to ensure response data structs aren't nil
							Added support for md5-token
				06 Jan 2015 - Added check for nil project id in response.
				03 Feb 2015 - Correct fmt format string error.
------------------------------------------------------------------------------------------------
*/

package ostack

import (
	//"bufio"
	"bytes"
	"crypto/md5"
	//"encoding/hex"
	"encoding/json"
	"fmt"
	//"io/ioutil"
	//"net/http"
	//"net/url"
	"os"
	"time"
)


// -------------------------------------------------------------------------------------------------

/*
	Compute the md5 hash of the string passed in and return a pointer to the hash as a string.
*/
func str2md5_str( src string ) ( *string ) {

	ho := md5.New( )							// create and add the string to the mix
	ho.Write( []byte( src ) )
	hash := ho.Sum( nil )						// finalise the summation
	hs := fmt.Sprintf( "%x", hash[:] )

	return &hs
}

// -------------------------------------------------------------------------------------------------

/*
	Sends an authorisation request to OpenStack and waits for it to return a
	token that can be used on subsequent calls.  Err is set to non-nil if 
	the credentials fail to authorise.  
*/
func (o *Ostack) Authorise( ) ( err error ) {
	var (
		auth_data	generic_response
		rjson		string
	)

	if o == nil ||  o.user == nil || o.passwd == nil {
		err = fmt.Errorf( "no openstack object to work on, or missing data inside" )
		return
	}

	o.token = nil			// must set this to nil to prevent it from being put into the header
	o.small_tok = nil
	if o.project == nil {
		rjson = fmt.Sprintf( `{ "auth": { "passwordCredentials": { "username": %q, "password": %q }}}`, *o.user, *o.passwd )				// just auth on uid and passwd to gen a token
	} else {
		rjson = fmt.Sprintf( `{ "auth": { "tenantName": %q, "passwordCredentials": { "username": %q, "password": %q }}}`, *o.project, *o.user, *o.passwd )
	}
	body := bytes.NewBufferString( rjson )

	url := *o.host + "v2.0/tokens"
	dump_url( "authorise", 10, url )
	jdata, _, err := o.Send_req( "POST",  &url, body ); 
	dump_json( "authorise", 10, jdata )

	if err != nil {
		return
	}

	err = json.Unmarshal( jdata, &auth_data )			// unpack the json into jif
	if err != nil {
		fmt.Fprintf( os.Stderr, "ostack: auth: unable to unpack json: %s\n", err )
		return
	}

	if auth_data.Error != nil {
		o.token = nil
		o.small_tok = nil
		err = fmt.Errorf( "auth failed: code=%d msg=%s\n", auth_data.Error.Code, auth_data.Error.Message )
		return
	}

	if auth_data.Access == nil  {
		err = fmt.Errorf( "auth failed: openstack response did not contain access data" )
		return
	}

	if auth_data.Access.Token == nil  {
		err = fmt.Errorf( "auth failed: openstack response did not contain token data" )
		return
	}

	o.expiry, err = Unix_time( &auth_data.Access.Token.Expires )			// convert openstack human time string to timestamp
	if err != nil {
		o.expiry = time.Now().Unix() + 300; 	// unable to parse the expiry date, assume it's good for 5min
	} else {
		o.expiry -= 60							// we'll chase a new token a minute before the actual expiration
	}

	o.token = &auth_data.Access.Token.Id
	if len( *o.token ) > 100 { 					// take the md5 only if it's huge, othewise we'll use the original here too
		o.small_tok = str2md5_str( *o.token )
	} else {
		o.small_tok = o.token
	}

	if auth_data.Access.Token != nil &&  auth_data.Access.Token.Tenant != nil   {
		o.project_id = &auth_data.Access.Token.Tenant.Id
	} else {
		dup_str := ""
		o.project_id = &dup_str
	}
	o.user_id = &auth_data.Access.User.Id
	o.chost = nil

	for i := range auth_data.Access.Servicecatalog {	// find the compute service stuff and pull the url to reach it 
		cat := auth_data.Access.Servicecatalog[i]

		switch cat.Type {
			case "compute":
				o.chost = &cat.Endpoints[0].Internalurl		// we'll blindly use the first for now

			case "network":
				o.nhost = &cat.Endpoints[0].Internalurl

			case "identity":
				o.ihost = &cat.Endpoints[0].Internalurl		// keystone host to list projects
				o.iahost = &cat.Endpoints[0].Adminurl		// admin url treats requests differently

			// note - if we ever need to capture ec2, it has a different admin url as well
		} 
	}

	if auth_data.Access.User != nil && auth_data.Access.User.Roles != nil  {		// don't fail if not present, but don't crash either
		o.isadmin = false
		for i := range  auth_data.Access.User.Roles {
			if  auth_data.Access.User.Roles[i].Name == "admin" {
				o.isadmin = true
			}
		}
	}

	if o.chost == nil {				// this happens if no project, or invalid project, is given and might be ok.
		empty_str := ""
		o.chost = &empty_str				// prevent failures if user attempts to make a call that needs chost
		//err = fmt.Errorf( "openstack did not find a compute host URL for %s/%s in the %d member service-cat", *o.user, *o.project, len( auth_data.Access.Servicecatalog ) )
	}

	return
}

// version 3 authorisation such a lovely example of backwards compatibility -- NOT!
//`{ "auth": { "identity": { "methods": [ "password" ], "password": { "user": { "id": %q, "password": %q, } } } } }`, 

/*
	Check to see if we think we are expired and if so, reexecute the authorisation. 
	(This is probably not needed as an external interfac, but might be convenient if
	an application wants to pre-authorise while doing other initialisation in parallel
	in order to speed things up.)
*/
func (o *Ostack) Validate_auth( ) ( err error ) {
	var (
		now 	int64
	)

	err = nil

	if o == nil {
		err = fmt.Errorf( "ostack_auth: openstack creds were nil (unable to authorise)" )
		return
	}

	if o.token == nil {
		err = o.Authorise()
	} else {
		now = time.Now().Unix()
		if now > o.expiry {				// our expiry should be less than openstacks so that we never attempt to use a stale token
			err = o.Authorise( )
			return
		}
	}

	if err == nil {
		if o.token == nil {				// parninoia
			err = fmt.Errorf( "openstack did not generate an authorisation token for %s/%s", *o.user, *o.project )
		}
	}

	return;			// shouldn't get here, but keeps compiler happy
}

/*
	Allow the user to force the authorisation to be expired which will force a new authorisation
	on the next request.
*/
func (o *Ostack) Expire( ) {
	o.token = nil
	o.small_tok = nil
	o.expiry = 0
}

/*
	Accept a token and put it in.
*/
func (o *Ostack) Insert_token( tok *string ) {
	if o != nil {
		o.token = tok
		o.small_tok = str2md5_str( *o.token )
	}
}

/*
	Returns true if the username and password that were authorised seemed to have admin privs too.
*/
func ( o *Ostack ) Isadmin( ) ( bool ) {
	return o.isadmin
}

/*
	Validate a token that is NOT associated with the credential block and optionally 
	checks  to see if it was issued for a specific user. Returns an error struct if 
	it is not valid, otherwise it will return nil. 

	Usr_match is a pointer to either the user name or the Openstack ID and if supplied
	will be matched against the data returned for the token.  If either the user name
	or ID returned matches, then the result is valid and the error return will be nil. 
	If usr_match is not given (nil), then the reult is good if there is no error 
	generated by openstack.

	This does NOT validate against a specific project.
*/
func (o *Ostack) Token_validation( token *string, usr_match *string ) ( expiry int64, err error ) {
	var (
		response_data	generic_response			// unpacked response json
		rjson string							// request body
	)

	expiry = 0

	if o == nil ||  o.user == nil || o.passwd == nil {
		err = fmt.Errorf( "no openstack object to work on, or missing data inside" )
		return
	}

	o.Validate_auth()							// ensure we're still auth to make requests

	rjson = fmt.Sprintf( `{ "auth": { "token": { "id": %q }}}`, *token );	// auth block contains the token to authenticate
	body := bytes.NewBufferString( rjson )

	url := *o.host + "v2.0/tokens"
	jdata, _, err := o.Send_req( "POST",  &url, body ); 
	//fmt.Fprintf( os.Stderr, ">>>token-valid json= %s\n", jdata );				// TESTING/verification
	dump_json( "token-valid", 10, jdata )

	if err != nil {	
		return
	}

	err = json.Unmarshal( jdata, &response_data )			// unpack the json into response data
	if err != nil {
		fmt.Fprintf( os.Stderr, "ostack: auth: unable to unpack json: %s\n", err )
		return
	}

	if response_data.Error == nil  {
		if response_data.Access != nil {
			if usr_match != nil {
				if response_data.Access.User != nil {
					if response_data.Access.User.Username != *usr_match && response_data.Access.User.Id != *usr_match {
						err = fmt.Errorf( "token is not valid: token was generated for %s/%s which is not the indiated user: %s", response_data.Access.User.Username, response_data.Access.User.Id, *usr_match )
					} else {
						expiry, err = Unix_time( &response_data.Access.Token.Expires )			// convert openstack human time string to timestamp
						if err != nil {
							expiry = 0
						}
						o.tok_isadmin[*token] = false
						for i := range  response_data.Access.User.Roles {
							if  response_data.Access.User.Roles[i].Name == "admin" {
								o.tok_isadmin[*token] = true
							}
						}
					}
				} else {
					err = fmt.Errorf( "token is not valid: response from openstack did not contain valid data: missing user information" )
				}
			}
		} else {
			err = fmt.Errorf( "token is not valid: response from openstack did not contain valid data: missing access information" )
		}
	} else {
		err = fmt.Errorf( "token is not valid: code=%d msg=%s", response_data.Error.Code, response_data.Error.Message )
	}

	return
}

/*
	Return the url for the desired service as defined by the EP_ constants.
	Returns a pointer to the string, or nil if none or bad constant.
*/
func (o *Ostack) Get_service_url( svc int ) ( *string ) {
	switch svc {
		case EP_COMPUTE:	return o.chost
		case EP_IDENTITY:	return o.ihost
		case EP_NETWORK:	return o.nhost
	}

	return nil
}
