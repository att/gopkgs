// vi: sw=4 ts=4:

/*
------------------------------------------------------------------------------------------------
	Mnemonic:	ostack_crack
	Abstract:	Functions that involve cracking open a token and returning an Ostack_tstuff
				structure to the user.   There aren't any specific functions to maniuplate 
				the tstuff stuct, but there might be so it's off in its own module.

				CAUTION: 	this uses the v3 interface since the same call in v2 returned
							useless information. Also, the V3 doesn't return any useful
							information in some environments either.

	Date:		03 April 2015 
	Authors:	E. Scott Daniels

	Mods:
				13 Apr 2015 - Converted to more generic error structure use.
				10 Jul 2015 - Added specific v3/v2 calls since the v3 call doesn't seem to 
						provide useful role information in all cases.
------------------------------------------------------------------------------------------------
*/

package ostack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)


/*
	Token stuff from a cracked token.
*/
type Ostack_tstuff struct  {
	User	string
	Id		string
	Roles	map[string]bool
	Expiry	int64
}

// ----- externally visiable wrappers to the generic crack function -----------------------------------

/*
	Accepts a token and sends a query to openstack to crack it open. Returns a structure with exposed
	fields for the caller to digest. (There are no functions associated with the cracked info structure.)

	This is the generic crack funtion and as such the following defaults apply:
		- openstack indentity version 2 interface is used
		- the project associated with the ostack struct used on the call is provided to 
		  scope the request.

	See Crack_ptoken for ways to change the defaults.
*/
func (o *Ostack) Crack_token( token *string ) ( stuff *Ostack_tstuff, err error ) {
	return o.crack_token( token, o.project, false )
}

/*
	Accepts a token and sends a query to openstack to crack it open. Returns a structure with exposed
	fields for the caller to digest. (There are no functions associated with the cracked info structure.)

	The use_v3 parameter causes the openstack indentity version 3 interface to be used in place of the 
	version 2 interface.
	
*/
func (o *Ostack) Crack_ptoken( token *string, project *string, use_v3 bool ) ( stuff *Ostack_tstuff, err error ) {
	return o.crack_token( token, project, use_v3 )
}

/*
	Accepts a token and project and has a go at cracking open the token. If successful, a struct is
	returned containing useful information from the token.  The struct's fields are all visible to 
	the user; there are no other functions associated with a stuff structure.
*/
func (o *Ostack) crack_token( token *string, project *string, use_v3 bool ) ( stuff *Ostack_tstuff, err error ) {
	var (
		rjson string						// request body to send 
	)

	if o == nil ||  o.user == nil || o.passwd == nil {
		err = fmt.Errorf( "no openstack object to work with, or missing data inside" )
		return
	}

	if len( *token ) > 100 {						// ostack cannot handle its own large tokens, so compress before sending
		token = str2md5_str( *token )
	}
	o.Validate_auth()							// ensure we're still auth to make requests

	if use_v3 {
		rjson = fmt.Sprintf( `{ "auth": { "identity": { "tenantName": %q, "methods": [ "token" ], "token": { "id": %q }}}}`, *project, *token );	// auth block contains the token to crack
	} else {
		rjson = fmt.Sprintf( `{ "auth": { "tenantName": %q, "token": { "id": %q } } }`, *project, *token )
	}

	body := bytes.NewBufferString( rjson )

	url := ""
	if use_v3 {
		url = *o.host + "v3/auth/tokens"
	} else {
		url = *o.host + "v2.0/tokens"
	}

	dump_url( "token-crack", 10, url + " " +  body.String() )
	jdata, _, err := o.Send_req( "POST",  &url, body ); 
	dump_json( "token-crack", 10, jdata )

	if err != nil {	
		return
	}

	if use_v3 {
		var response_data	osv3_generic					// version 3 format not compatable with v2

		err = json.Unmarshal( jdata, &response_data )			// unpack the json into response data
		if err != nil {
			fmt.Fprintf( os.Stderr, "ostack: crack-v3: unable to unpack json: %s\n", err )
			dump_json( "token-crack", 30, jdata )
			return
		}

		if response_data.Error == nil  {
			if response_data.Token.User != nil {
				stuff = &Ostack_tstuff{ }
	
				stuff.User = response_data.Token.User.Name
				stuff.Id = response_data.Token.User.Id
				stuff.Expiry, err = Unix_time( &response_data.Token.Expires_at )			// convert openstack human time string to timestamp
				if err != nil {
					stuff.Expiry = 0
				}
	
				if len( response_data.Token.Roles ) > 0 {
					stuff.Roles = make( map[string]bool, len( response_data.Token.Roles ) )
					for i := range  response_data.Token.Roles {
						stuff.Roles[response_data.Token.Roles[i].Name] = true
					}
				} 
	
			} else {
				err = fmt.Errorf( "token is not valid: response from openstack did not contain valid data: missing user information" )
			}
		} else {
			err = fmt.Errorf( "token is not valid: %s\n", response_data.Error )
		}
	} else {
		var response_data	generic_response			// version 2 format not compatable with v3

		err = json.Unmarshal( jdata, &response_data )			// unpack the json into response data
		if err != nil {
			fmt.Fprintf( os.Stderr, "ostack: crack-v2: unable to unpack json: %s\n", err )
			dump_json( "token-crack", 30, jdata )
			return
		}

		if response_data.Error == nil  {
			if response_data.Access.User != nil {				// blody v3 adds a layer; ditch for v2
				stuff = &Ostack_tstuff{ }
	
				stuff.User = response_data.Access.User.Name
				stuff.Id = response_data.Access.User.Id
				stuff.Expiry, err = Unix_time( &response_data.Access.Token.Expires )			// convert openstack human time string to timestamp
				if err != nil {
					stuff.Expiry = 0
				}
	
				if len( response_data.Access.User.Roles ) > 0 {
					stuff.Roles = make( map[string]bool, len( response_data.Roles ) )
					for i := range  response_data.Access.User.Roles {
						stuff.Roles[response_data.Access.User.Roles[i].Name] = true
					}
				} 
	
			} else {
				err = fmt.Errorf( "token is not valid: v2 response from openstack did not contain valid data: missing user information" )
			}
		} else {
			err = fmt.Errorf( "token is not valid (v2): %s\n", response_data.Error )
		}
	}

	return
}

/*
	Convenience function for printing. Not efficient, but simple
*/
func (s *Ostack_tstuff) String() ( string ) {
	sep := ""
	roles := ""
	for r := range s.Roles {
		roles = roles + sep + r	
		sep = " "
	}

	return fmt.Sprintf( "User: %s  Id: %s Expiry: %d Roles: %s", s.User, s.Id, s.Expiry, roles )
}
