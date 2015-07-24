//vi: sw=4 ts=4:
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
	Mnemonic:	ostack_dump
	Abstract:	Some generic debugging functions to execute some ostack api call and
				dump the resulting json to standard error.
					

	Date:		17 June 2014
	Author:		E. Scott Daniels
	Mod:		03 Feb 2015 - Fixed fprintf() statement -- too many %s
				24 Jun 2015 - Moved Dump_json from network.go to this module.
------------------------------------------------------------------------------------------------
*/

package ostack

import (
	"bytes"
	"fmt"
	"os"
)

// ------------------------------------------------------------------------------

/*
	Runs a command and dumps the json to stderr -- debugging mostly.
*/
func (o *Ostack) Dump_cmd_response( cmd string ) ( error ) {

	err := o.Validate_auth()				// reauthorise if needed
	if err != nil {
    	fmt.Fprintf( os.Stderr, "unable to validate creds: %s\n", err );
		return err
	}

	body := bytes.NewBufferString( "" )

	//url := fmt.Sprintf( "%sv2.0/tokens/%s", *o.host, *token )								// this returns token information if token is valid
	url := fmt.Sprintf( "%s%s", *o.host, cmd )								
    jdata, _, err := o.Send_req( "GET",  &url, body );

	if err != nil {
		return err
	}
    fmt.Fprintf( os.Stderr, "json= %s\n", jdata );
	return nil
}

func (o *Ostack) Dump_json( uurl string ) ( err error ) {
	var (
		jdata	[]byte				// raw json response data
	)

	if o == nil {
		err = fmt.Errorf( "net_subnets: openstack creds were nil" )
		return
	}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	if o.nhost == nil || *o.nhost == "" {
		err = fmt.Errorf( "no network host url to query %s", o.To_str() )
		return
	}

	jdata = nil
	body := bytes.NewBufferString( "" )

	//url := fmt.Sprintf( "%s/v2.0/subnets", *o.nhost )				// nhost is the host where the network service is running
	//url := fmt.Sprintf( "%s/%s", *o.nhost, uurl )				// nhost is the host where the network service is running
	url := fmt.Sprintf( "%s/%s", *o.chost, uurl )				// nhost is the host where the network service is running
	jdata, _, err = o.Send_req( "GET",  &url, body )

	if err != nil {
		fmt.Fprintf( os.Stderr, "error: %s\n", err )
		return
	}

	fmt.Fprintf( os.Stderr, "json= %s\n", jdata )

	return
}
