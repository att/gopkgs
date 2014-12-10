// vi: sw=4 ts=4:

/*
------------------------------------------------------------------------------------------------
	Mnemonic:	ostack_dump
	Abstract:	Run a command and dump the resulting json to stderr. 
					

	Date:		17 June 2014
	Author:		E. Scott Daniels
	Mod:		
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
    	fmt.Fprintf( os.Stderr, "cmmand not executed (%s) unable to validate creds: %s\n", err );               
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
