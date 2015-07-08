// vi: sw=4 ts=4:

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
