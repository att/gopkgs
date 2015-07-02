/*
-----------------------------------------------------------------------------------------
	Mnemonic:	ostack_armour
	Abstract:	Functions that are necessary to protect against failures because of 
				dumb configurations (like proxy's that return html).
	Author:		E. Scott Daniels
	Date:		16 August 2014
-----------------------------------------------------------------------------------------
*/

package ostack

import (
	"bytes"
	"fmt"
)



/*
	Accept a buffer of json, read (we assume) from openstack, and validate it for bad things:
		leading '<' before '{' is found (suggesting html), null bytes etc.
*/
func scanj4gook( buf []byte ) ( err error ) {

	jloc := bytes.IndexAny( buf, "{" )

	if jloc < 0 {
		dump_array( "scan4gook", 30, buf )
		return fmt.Errorf( "invalid json: no opening curly brace found" )
	}

	hloc := bytes.IndexAny( buf, "<>" )
	if hloc > 0 && hloc < jloc {			// html tag seems to be found before possible json
		dump_array( "scan4gook", 30, buf )
		return fmt.Errorf( "invalid json: HTML tag, or similar, found outside of json" )
	}

	return nil
}
