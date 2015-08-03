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
