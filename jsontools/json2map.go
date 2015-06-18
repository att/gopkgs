// vi: sw=4 ts=4:

/*

	Mnemonic:	json2map
	Abstract:	Take a bunch of json and construct a map.
	Date:		16 December 2013
	Author:		E. Scott Daniels
*/

package jsontools

import (
	"encoding/json"
	"fmt"
	"os"
)

// --------------------- public ----------------------------------------------------------------------

/*
	Builds a map from the interface 'structure', aka jif, (ithing) passed in putting into hashtab. The map names
	generated for hashtab are dot separated.  For example, the json {foo:"foo-val", bar:["bar1", "bar2"] }
	would generate the following names in the map:
	root.foo, root.bar[0], root.bar[1] where 'root' is the initial tag passed in.
*/
func Jif2map( hashtab map[string]interface{}, ithing interface{}, depth int, ptag string, printit bool ) {

	switch ithing.( type )  {
		case map[string]interface{}:					// named map of things
			for k, v := range ithing.( map[string]interface{} ) {
				Jif2map( hashtab, v, depth + 1, fmt.Sprintf( "%s.%s", ptag, k ), printit )
			}

		case []interface{}:
				a := ithing.( []interface{} )					// unnecessary, but easier to read
				alen := fmt.Sprintf( "%s[len]", ptag )	
				hashtab[alen] = len( a );
				if printit { fmt.Printf( "%s = %d\n", alen, len(a) ) }
				for i := range a {
					switch a[i].( type ) {
						case map[string]interface{}:
							Jif2map( hashtab, a[i], depth, fmt.Sprintf( "%s[%d]", ptag, i ), printit )

						case interface{}:
							Jif2map( hashtab, a[i], depth, fmt.Sprintf( "%s[%d]", ptag, i ), printit )

						default:
							fmt.Fprintf( os.Stderr, "ERR: Jif2map: unhandled type of interface array; not added to the hashtable.\n" )
					}
				}

		case string:
			if printit { fmt.Printf( "%s = %s\n", ptag, ithing.(string) ) }
			hashtab[ptag] = ithing

		case int:
			if printit { fmt.Printf( "%s = %d\n", ptag, ithing.(int) ) }
			hashtab[ptag] = ithing

		case float64:
			if printit { fmt.Printf( "%s = %.2f\n", ptag, ithing.(float64) ) }
			hashtab[ptag] = ithing

		case bool:
			vstate := "false"
			if ithing.(bool) {
				vstate = "true"
			}
			if printit { fmt.Printf( "%s = %s\n", ptag, vstate ) }
			hashtab[ptag] = ithing

		case interface{}:	
			Jif2map( hashtab, ithing, depth+1, ptag, printit )
			hashtab[ptag] = ithing

		case nil:
			if printit { fmt.Printf( "%s = null\n", ptag ) }
			hashtab[ptag] = nil

		default:	
				fmt.Fprintf( os.Stderr, "ERROR: unpack: unexpected type at depth %d\n", depth )
				// TODO:  return error rather than aborting
				os.Exit( 1 )
	}

	return;
}


/*
	Accepts the json blob as a byte string, probably just read off of the wire, and builds a symbol table.
	If the printit parm is true, then the representation of the json is written to the standard output
	in addition to the generation of the map.
*/
func Json2map( json_blob []byte, root_tag *string, printit bool ) ( symtab map[string]interface{}, err error ) {
	var (
		jif	interface{};				// were go's json will unpack the blob into interface form
		def_root_tag string = "root";	
	)

	
	symtab = nil;
	err = json.Unmarshal( json_blob, &jif )			// unpack the json into jif
	if err != nil {
		return;
	}

	if root_tag == nil {
		root_tag = &def_root_tag;
	}

	symtab = make( map[string]interface{} );
	Jif2map( symtab, jif, 0, *root_tag, printit )			// unpack the jif into the symtab

	return;
}

