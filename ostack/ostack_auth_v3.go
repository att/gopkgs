// vi: sw=4 ts=4:
/*
 ---------------------------------------------------------------------------
   Copyright (c) 2013-2015 AT&T Intellectual Property

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at:

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 ---------------------------------------------------------------------------
*/


/*
------------------------------------------------------------------------------------------------
	Mnemonic:	ostack_auth_v3
	Abstract:	Interface to ostack to do authorisation and authorisation (endpoint) related
				things for version 3.
	Date:		16 August 2016
	Authors:	Pradeep Gondu, E. Scott Daniels

------------------------------------------------------------------------------------------------
*/

package ostack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)


/*
	Sends an authorisation request to OpenStack and waits for it to return a
	token that can be used on subsequent calls.  Err is set to non-nil if
	the credentials fail to authorise.   Region points to a string used to
	identify the "region" of the keystone authorisation catalogue that should
	be used to snarf URLs for things.  If it is nil, or points to "", then
	the first entry in the catalogue is used.
*/
func (o *Ostack) Authorise_region_v3( region *string ) ( err error ) {
	var (
		auth_data	generic_response
		rjson		string
	)
	if o == nil ||  o.user == nil || o.passwd == nil {
		err = fmt.Errorf( "no openstack object to work on, or missing data inside" )
		return
	}
	if o.project == nil {
		rjson = fmt.Sprintf( `{ "auth":{ "identity": { "methods": ["password"], "password": {"user": {"name": %q, "domain": { "id": "default" },"password": %q } } } } }`, *o.user, *o.passwd )
	} else {
		rjson = fmt.Sprintf( `{ "auth":{ "identity": { "methods": ["password"], "password": {"user": {"name": %q, "domain": { "id": "default" },"password": %q } } }, "scope": { "project": { "name": %q, "domain": { "id": "default" } } } } }`, *o.user, *o.passwd, *o.project )
	}
	body := bytes.NewBufferString( rjson )
	url := *o.host + "v3/auth/tokens"
	dump_url( "authorise", 10, url )
	jdata, _, err := o.Send_req( "POST",  &url, body );
	dump_json( "authorise", 10, jdata )

	if err != nil {
		return
	}

	err = json.Unmarshal( jdata, &auth_data )			// unpack the json into jif
	if err != nil {
		fmt.Fprintf( os.Stderr, "ostack: auth: unable to unpack (auth) json: %s\n", err )
		dump_json( "authorise", 20, jdata )				// dump the json the first few times this happens
		return
	}

	if auth_data.Error != nil {
		o.token = nil
		o.small_tok = nil
		err = fmt.Errorf( "auth failed: %s\n", auth_data.Error )
		return
	}

	if auth_data.Token == nil  {
		err = fmt.Errorf( "auth failed: openstack response did not contain Token data" )
		return
	}

	o.expiry, err = Unix_time( &auth_data.Token.Expires_at )			// convert openstack human time string to timestamp
	if err != nil {
		o.expiry = time.Now().Unix() + 300; 	// unable to parse the expiry date, assume it's good for 5min
	} else {
		o.expiry -= 60							// we'll chase a new token a minute before the actual expiration
	}

	if len( *o.token ) > 100 { 					// take the md5 only if it's huge, otherwise we'll use the original here too
		o.small_tok = str2md5_str( *o.token )
	} else {
		o.small_tok = o.token
	}

	if *o.token != "" &&  auth_data.Token.Project != nil   {
		o.project_id = &auth_data.Token.Project.Id
	} else {
		dup_str := ""
		o.project_id = &dup_str
	}
	o.user_id = &auth_data.Token.User.Id
	o.chost = nil

	if region == nil {
		region = o.aregion								// use what was seeded on the Mk_ostack() call
	}

	found := 0											// number we found
	for i := range auth_data.Token.Catalog {	// find the compute service stuff and pull the url to reach it
		cat := auth_data.Token.Catalog[i]
		r := 0
		if region != nil && *region != ""  {			// default to first region if none given, or empty string given
			for _ = range cat.Endpoints {				// find the region that matches the name given
				if cat.Endpoints[r].Region == *region {
					break
				}
				r++
			}
		}

		if r < len( cat.Endpoints ) {					// found region in this group
			found++

			switch cat.Type {
				case "compute":
					for i := range cat.Endpoints{
						if cat.Endpoints[i].Interface == "internal" {
							o.chost = &cat.Endpoints[i].Url
						}
						if cat.Endpoints[i].Interface == "admin" {
							o.cahost = strip_ver( cat.Endpoints[i].Url )
						}
					}

				case "network":
					for i := range cat.Endpoints{
						if cat.Endpoints[i].Interface == "internal" {
							o.nhost = &cat.Endpoints[i].Url
						}
					}

				case "identity":
					for i := range cat.Endpoints{
						if cat.Endpoints[i].Interface == "internal" {
							o.ihost = strip_ver( cat.Endpoints[i].Url )		// keystone host to list projects
						}
						if cat.Endpoints[i].Interface == "admin" {
							o.iahost = strip_ver( cat.Endpoints[i].Url )
						}
					}
				// note - if we ever need to capture ec2, it has a different admin url as well
			}
		}
	}

	if  len( auth_data.Token.Catalog ) > 0 && found == 0 && err == nil {				// if there is a catalogue error if we didn't see region at all
		err = fmt.Errorf( "unable to find region in any openstack endpoint in list: %s", *region )
	}

	if auth_data.Token.User != nil && auth_data.Token.Roles != nil  {		// don't fail if not present, but don't crash either
		o.isadmin = false
		for i := range  auth_data.Token.Roles {
			if  auth_data.Token.Roles[i].Name == "admin" {
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

/*
	Backward compatible -- authorises for what ever is first in the list from a region perspective.
*/
func (o *Ostack) Authorise_v3( ) ( err error ) {
	return o.Authorise_region_v3( nil )
}

/*
	Check to see if we think we are expired and if so, reexecute the authorisation.
	(This is probably not needed as an external interfac, but might be convenient if
	an application wants to pre-authorise while doing other initialisation in parallel
	in order to speed things up.)

	This function will use the region that was set when the Ostack struct was created
	(see Mk_ostack() and Mk_ostack_region()).
*/
func (o *Ostack) Validate_auth_v3( ) ( err error ) {
	var (
		now 	int64
	)

	err = nil

	if o == nil {
		err = fmt.Errorf( "ostack_auth: openstack creds were nil (unable to authorise)" )
		return
	}

	if o.token == nil {
		err = o.Authorise_v3()
	} else {
		now = time.Now().Unix()
		if now > o.expiry {				// our expiry should be less than openstacks so that we never attempt to use a stale token
			err = o.Authorise_v3( )
		} else {
			return
		}
	}

	if err == nil {
		if o.token == nil {
			err = fmt.Errorf( "openstack did not generate an authorisation token for %s/%s", *o.user, *o.project )
		}
	}

	return;
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
func (o *Ostack) Token_validation_v3( token *string, usr_match *string ) ( expiry int64, err error ) {
	var (
		response_data	generic_response			// unpacked response json
		rjson string							// request body
	)

	expiry = 0

	if o == nil ||  o.user == nil || o.passwd == nil {
		err = fmt.Errorf( "no openstack object to work on, or missing data inside" )
		return
	}

	if len( *token ) > 100 {
		fmt.Println("str2md5_str is working no need of separate func")						// ostack cannot handle its own large tokens, so compress before sending
		token = str2md5_str( *token )
	}
	o.Validate_auth_v3()							// ensure we're still auth to make requests

	// rjson = fmt.Sprintf( `{ "auth": { "token": { "id": %q }}}`, *token );	// auth block contains the token to authenticate
	rjson = fmt.Sprintf(`{ "auth": { "identity": { "methods": ["token"], "token": { "id": %q }}}}`, *token );
	body := bytes.NewBufferString( rjson )


	url := *o.host + "v3/auth/tokens"
	jdata, _, err := o.Send_req( "POST",  &url, body );
	//fmt.Fprintf( os.Stderr, ">>>token-valid json= %s\n", jdata );				// TESTING/verification
	dump_json( "token-valid", 10, jdata )

	if err != nil {
		return
	}

	err = json.Unmarshal( jdata, &response_data )			// unpack the json into response data
	if err != nil {
		fmt.Fprintf( os.Stderr, "ostack: auth: unable to unpack (validation) json: %s\n", err )
		dump_json( "token-valid", 20, jdata )				// dump the json the first few times this happens
		return
	}

	if response_data.Error == nil  {
		if response_data.Token != nil {
			if usr_match != nil {
				if response_data.Token.User != nil {
					if response_data.Token.User.Name != *usr_match && response_data.Token.User.Id != *usr_match {
						err = fmt.Errorf( "token is not valid: token was generated for %s/%s which is not the indicated user: %s", response_data.Token.User.Name, response_data.Token.User.Id, *usr_match )
					} else {
						expiry, err = Unix_time( &response_data.Token.Expires_at )			// convert openstack human time string to timestamp
						if err != nil {
							expiry = 0
						}
						o.tok_isadmin[*token] = false
						for i := range  response_data.Token.Roles {
							if  response_data.Token.Roles[i].Name == "admin" {
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
		err = fmt.Errorf( "token is not valid: %s\n", response_data.Error )
	}

	return
}
