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

	Mnemonic:	prtmap
	Abstract:	This may apply outside of the json world, but plays well with json
				needs so it's here.  These functions take a bunch of unknown
				stuff (described as an interface) and prints them in a hierarchical manner.

				The public Print() function takes the the 'root' of the data and
				invokes either prettyprint() to print it in a 'tree-ish' format, o
				or dotprint() to print it in a dotted notation. The ptag supplied
				is used to name the 'root'.

	Date:		02 January 2014
	Author:		E. Scott Daniels
*/

package jsontools

import (
	"fmt"
	"os"
)

/*
	users can pass on calls to make their code more readable
*/
const (
	PRINT	bool = true
	NOPRINT	bool = false
)

var pp_need_blank bool;

/*
	take an unmarshalled blob and print it as a hierarchy. If dotted output is desired
	the blob must be converted into a map (see jsontools.Unpack) and then printed by
	the dotprint method here.
	ptag is the name of the item's parent
*/
func prettyprint( ithing interface{}, depth int, ptag string ) {
	var (
		indention	string = "                                                                                                                                "
	)

	istr := indention[0:depth*4]
	switch ithing.( type )  {
		case map[string]interface{}:					// named map of things
			pp_need_blank = false;
			if ptag != "" {
				fmt.Printf( "%s%s:\n", istr, ptag  )
			}
			ptag = ""
			for k, v := range ithing.( map[string]interface{} ) {
				prettyprint( v, depth + 1, k )
			}

		case []interface{}:
			pp_need_blank = true;
			fmt.Printf( "\n" )
			a := ithing.( []interface{} )					// unnecessary, but easier to read
			for i := range a {
				switch a[i].( type ) {
					case map[string]interface{}:
						fmt.Printf( "%s%s[%d]:\n", istr, ptag, i )
						prettyprint( a[i], depth, "" )

					case interface{}:
						prettyprint( a[i], depth, fmt.Sprintf( "%s[%d]", ptag, i ) )

					default:
						fmt.Fprintf( os.Stderr, "%sERROR: unhandled type of interface array\n", istr )
				}
			}

		case string:
			fmt.Printf( "%s%s = %s\n", istr, ptag, ithing.(string) )

		case int:
			fmt.Printf( "%s%s = %d\n", istr, ptag, ithing.(int) )

		case float64:
			fmt.Printf( "%s%s = %.2f\n", istr, ptag, ithing.(float64) )

		case bool:
			vstate := "false"
			if ithing.(bool) {
				vstate = "true"
			}
			fmt.Printf( "%s%s = %s\n", istr, ptag, vstate )

		case interface{}:	
			pp_need_blank = true;
			fmt.Printf( "%s%s:\n", istr, ptag )
			prettyprint( ithing, depth+1, ptag )
			//if pp_need_blank {
				//fmt.Printf( "\n" )
			//}
			//pp_need_blank = false;

		case nil:
			fmt.Printf( "%s%s = null/undefined\n", istr, ptag )

		default:	
				fmt.Printf( "error: unexpected type at depth %d\n", depth )
				os.Exit( 1 )
	}

	if pp_need_blank {
		fmt.Printf( "\n" )
	}	
	pp_need_blank = false;
}

/*
	Takes an unknown thing and prints it in a hierarchical form as a set of dot separated names.
	ptag is the name of the item's parent
*/
func dotprint( ithing interface{}, depth int, ptag string ) {

	switch ithing.( type )  {
		case map[string]interface{}:					// named map of things
			ptag = ""
			for k, v := range ithing.( map[string]interface{} ) {
				dotprint( v, depth + 1, k )
			}
			fmt.Printf( "\n" )

		case []interface{}:
				a := ithing.( []interface{} )					// unnecessary, but easier to read
				for i := range a {
					switch a[i].( type ) {
						case map[string]interface{}:
							fmt.Printf( "%s[%d]:\n", ptag, i )
							dotprint( a[i], depth, "" )

						case interface{}:
							dotprint( a[i], depth, fmt.Sprintf( "%s[%d]", ptag, i ) )

						default:
							fmt.Fprintf( os.Stderr, "ERROR: unhandled type of interface array\n" )
					}
				}
				fmt.Printf( "\n" )

		case string:
			fmt.Printf( "%s = %s\n", ptag, ithing.(string) )

		case int:
			fmt.Printf( "%s = %d\n", ptag, ithing.(int) )

		case float64:
			fmt.Printf( "%s = %.2f\n", ptag, ithing.(float64) )

		case bool:
			vstate := "false"
			if ithing.(bool) {
				vstate = "true"
			}
			fmt.Printf( "%s = %s\n", ptag, vstate )

		case interface{}:	
			fmt.Printf( "%s:\n", ptag )
			dotprint( ithing, depth+1, ptag )
			fmt.Printf( "\n" )

		case nil:
			fmt.Printf( "%s = null/undefined\n",  ptag )

		default:	
				fmt.Printf( "error: unexpected type at depth %d\n", depth )
				os.Exit( 1 )
	}
}
// -------------------- public----------------------------------------------------------------------------------

/*
	A simple method that allows an interface representation to easily be printed in either hierarchical
	or "dotted" format.
*/
func Print( stuff interface{}, ptag string, dotted bool ) {
	if dotted {
		dotprint( stuff, 0, ptag );
	} else {
		prettyprint( stuff, 0, ptag );
	}
}
