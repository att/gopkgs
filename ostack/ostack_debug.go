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

package ostack

/*
	Mnemonic:	ostack_debug
	Abstract:	Some quick debug functions.
	Author:		E. Scott Daniels
	Date:		7 August 2014
*/

import (
	"fmt"
	"os"
)

/*
	Dump the url if the counter max is not yet reached.
*/
func dump_url( id string, max int, url string ) {
	if  dbug_url_count < max {
		dbug_url_count++

		fmt.Fprintf( os.Stderr, "%s >>> url= %s\n", id, url )
	}
}

/*
	Dump the raw json array if the counter max has not yet been reached.
*/
func dump_json( id string, max int, json []byte ) {
	if  dbug_json_count < max {
		dbug_json_count++

		fmt.Fprintf( os.Stderr, "%s >>> json= %s\n", id, json )
	}
}

/*
	Dump a raw array if the counter max has not yet been reached.
*/
func dump_array( id string, max int, stuff []byte ) {
	if  dbug_json_count < max {
		dbug_json_count++

		fmt.Fprintf( os.Stderr, "%s >>> raw= %s\n", id, stuff )
	}
}

/*
	Allow debugging to be reset or switched off. Set high (20) to
	turn off, 0 to reset.
*/
func Set_debugging( count int ) {
	dbug_json_count = count
	dbug_url_count = count
}
