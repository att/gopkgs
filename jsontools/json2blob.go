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

	Mnemonic:	json2blob
	Abstract:	Takes json input and generates a 'raw' interface built from it. The output
				can be used as input to the pretty printer.
	Date:		16 December 2013
	Author:		E. Scott Daniels
*/

package jsontools

import (
	"encoding/json"
)

// --------------------- public ----------------------------------------------------------------------

/*
	Accepts the json "blob" as a byte string, probably just read off of the wire. A generic interface
	is returned and if the printit parameter is true, then a representation of the json is printed
	to the standard output device.
*/
func Json2blob( json_blob []byte, root_tag *string, printit bool ) ( jif interface{}, err error ) {
	var (
		def_root_tag string = "root";	
	)

	err = json.Unmarshal( json_blob, &jif )			// unpack the json into jif
	if err != nil {
		jif = nil;
		return;
	}

	if root_tag == nil {
		root_tag = &def_root_tag;
	}

	if printit {
		prettyprint( jif, 0, *root_tag );
	}

	return;
}

