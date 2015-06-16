// vi: sw=4 ts=4:

/*
------------------------------------------------------------------------------------------------
	Mnemonic:	ostack_crack
	Abstract:	Functions that involve cracking open a token and returning an Ostack_tstuff
				structure to the user.   There aren't any specific functions to maniuplate 
				the tstuff stuct, but there might be so it's off in its own module.

				CAUTION: 	this uses the v3 interface since the same call in v2 returned
							useless information.

	Date:		03 April 2015 
	Authors:	E. Scott Daniels

	Mods:
				13 Apr 2015 - Converted to more generic error structure use.
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

/*
	Accepts a token and sends a query to openstack to crack it open. Returns a structure with exposed
	fields for the caller to digest. (There are no functions associated with the cracked info structure.)
*/
func (o *Ostack) Crack_token( token *string ) ( stuff *Ostack_tstuff, err error ) {
	var (
		response_data	osv3_generic			// returned stuff after unmashed from json
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

	rjson = fmt.Sprintf( `{ "auth": { "identity": { "methods": [ "token" ], "token": { "id": %q }}}}`, *token );	// auth block contains the token to crack
	body := bytes.NewBufferString( rjson )

	url := *o.host + "v3/auth/tokens"
	jdata, _, err := o.Send_req( "POST",  &url, body ); 
	dump_json( "token-crack", 10, jdata )

	if err != nil {	
		return
	}

	err = json.Unmarshal( jdata, &response_data )			// unpack the json into response data
	if err != nil {
		fmt.Fprintf( os.Stderr, "ostack: crack: unable to unpack json: %s\n", err )
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
