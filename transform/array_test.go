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
        Mnemonic:       array_test.go
        Absrtract:      Test the ability to populate an array in a subobject.
		Date:			08 February 2017
		Author:			E. Scott Daniels
*/

package transform_test

import (
	"fmt"
	"os"
	//"strings"
	"testing"

	"github.com/att/gopkgs/transform"
	"github.com/att/gopkgs/clike"
)

/*
	Sub object embedded with an interface that was saved as an array
	rather than a simple value.
*/
type Inner_thing struct {
	Name	string		`Bar:"_"`
	Data	interface{} `Bar:"_"`

	// todo -- handle this too
	//Data	[]string	`Bar:"_"`
}

/*
	Array test thing; main object.
*/
type AT_thing struct {
	Bar_thing_s string		`Bar:"_"`
	Itobj* Inner_thing  	`Bar:"_"`
}

/*
	Create an object to convert to a map.
*/
func mk_thing( ) ( *AT_thing ) {
	t := &AT_thing {
		Bar_thing_s: "hello world, I am a thing",
	}

	t.Itobj = &Inner_thing{ Name: "I'm the inner thing", }

	a := make( []string, 10 )
	a[0] = "string 0"
	a[1] = "string 1"
	a[2] = "string 2"
	a[3] = "string 3"
	t.Itobj.Data = a
	// CAUTION: test pass is based on not finding anything but empty string in 4-9

	return t
}

/*
	Should generate a map.
*/
func TestIting2Map( t *testing.T ) {
	ok := true

	it := mk_thing()
	fmt.Fprintf( os.Stderr, "array_test: Inner AT_thing Testing\n" )
	m := transform.Struct_to_map( it, "Bar" )
	for k, v := range m {
		fmt.Fprintf( os.Stderr, "AT_thing map: %s (%v)\n", k, v )			// uncomment to dump the map
	}

	if( len( m ) != 14 ) {
		fmt.Fprintf( os.Stderr, "array_test: [FAIL] number of elements in the map not 14; found %d\n", len( m ) )
		t.Fail()
	} else {
		fmt.Fprintf( os.Stderr, "array_test: [OK]   number of elements in the map is correct\n" )
	}

	if( clike.Atoi( m["Itobj/Data.cap"] ) != 10 ) {
		fmt.Fprintf( os.Stderr, "array_test: [FAIL] number of inner elements in the map not 10; found %s\n", m["Itobj/Data.cap"] )
		t.Fail()
		ok = false
	} else {
		fmt.Fprintf( os.Stderr, "array_test: [OK]   number of inner elements in the map is correct\n" )
	}

	fmt.Fprintf( os.Stderr, "\n" )
	nit := &AT_thing{}										// empty struct to populate
	transform.Map_to_struct( m, nit,  "Bar" )				// fill it in 

	a, ok := nit.Itobj.Data.( []string )
	if ok {
		if( len( a )  != 10 ) {
			fmt.Fprintf( os.Stderr, "array_test: [FAIL] len of populated inner data array is not 10: %d\n", len( a ) )
			t.Fail()
			ok = false
		} else {
			fmt.Fprintf( os.Stderr, "array_test: [OK]   len of populated inner data array is correct\n" )
	
			for i := 0; i < 10; i++ {
				if i > 3 && a[i] != "" {
					ok = false
					fmt.Fprintf( os.Stderr, "array_test: [FAIL] data[4+] should be empty and isn't @i=%d\n", i )
					t.Fail()
				}
				fmt.Fprintf( os.Stderr, "array_test: [INFO] data[%d] = (%s)\n", i, a[i] )
			}
		}
	} else {
		ok = false
		fmt.Fprintf( os.Stderr, "array_test: [FAIL] data didn't produce an array of strings\n" )
		t.Fail()
	}

	if ok {
		fmt.Fprintf( os.Stderr, "PASS: array_test: done\n\n" )
	} else {
		fmt.Fprintf( os.Stderr, "FAIL: array_test: done\n\n" )
	}

}

