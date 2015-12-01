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
        Mnemonic:       map_test.go
        Absrtract:      Test map oriented functions in transform package
		Date:			25 November 2015
		Author:			E. Scott Daniels
*/

package transform_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/att/gopkgs/transform"
)

/*
	Annon structs need to be labeled too.
*/
type Adjunct struct {
	Foo_AJ1	int				`Foo:"_"`
	Foo_AJ2	int				`Foo:"_"`
	Foo_AJ3	int				`Foo:"_"`
	Bar_AJ4	int				`Bar:"_"`
	Unmarked int
}

type Thing struct {
	s string
}

type Foo_bar struct {
						Adjunct					// annon structure

	Athing				Thing `Goo:"_"`
	Foo_str string		`Foo:"FooString"`
	Foo_int int			`Foo:"FooInteger"`
	Foo_intp1 *int  	`Foo:"FooPtr1"`
	Foo_intp2 *int  	`Foo:"FooPtr2"`
	Foo_uint uint		`Foo:"_"`
	Foo_uintp1 *uint16	`Foo:"_"`
	Foo_uintp2 *uint16	`Foo:"_"`
	Foo_bool bool		`Foo:"_"`
	Foo_boolp *bool		`Foo:"_"`

	Bar_str string		`Bar:"BarString"`
	Bar_int int			`Bar:"BarInteger"`
	Bar_intp1 *int  	`Bar:"BarPtr1"`
	Bar_intp2 *int  	`Bar:"BarPtr2"`
	Bar_uint uint		`Bar:"_"`
	Bar_uintp1 *uint16	`Bar:"_"`
	Bar_uintp2 *uint16	`Bar:"_"`
	Bar_bool bool		`Bar:"_"`
	Bar_boolp *bool		`Bar:"_"`

	Baz_str string		`Baz:"_"`
	Baz_int int			`Baz:"_"`

	internal1 int
	internal2 int
}

func mk_struct1( ) ( *Foo_bar ) {
	fb := false
	fip1 := 987
	fip2 := 876
	fup1 := uint16( 432 )
	fup2 := uint16( 321 )

	bb := true
	bip1 := 1987
	bip2 := 1876
	bup1 := uint16( 1432 )
	bup2 := uint16( 1321 )
	
	fbs := &Foo_bar{
		Foo_str:	"foo-str",
		Foo_int:	123,
		Foo_intp1:	&fip1,
		Foo_intp2:	&fip2,
		Foo_uint:	345,
		Foo_uintp1: &fup1,
		Foo_uintp2: &fup2,
		Foo_bool:	true,
		Foo_boolp: 	&fb,
	
		Bar_str:	"bar-str",
		Bar_int:	1123,
		Bar_intp1:	&bip1,
		Bar_intp2:	&bip2,
		Bar_uint:	1345,
		Bar_uintp1:	&bup1,
		Bar_uintp2: &bup2,
		Bar_bool:	false,
		Bar_boolp:  &bb,
	
		Baz_str:	"baz-str",
		Baz_int:	2000,

		internal1:	1,
		internal2:	2,
	}

	fbs.Foo_AJ1 = 10001
	fbs.Foo_AJ2 = 10002
	fbs.Foo_AJ3 = 10003

	return fbs
}
		

/*
	Should generate a map with only foo tagged items
*/
func TestFooOnly( t *testing.T ) {
	fs := mk_struct1()
	count := 0

	fmt.Fprintf( os.Stderr, "testfoo only\n" )
	m := transform.Struct_to_map( fs, "Foo" )		// should only generate foo elements into map
	for k, v := range m {
		if strings.Index( k, "Foo" ) != 0 {
			fmt.Fprintf( os.Stderr, "BAD: unexpected key found in 'foo' map: %s (%v)\n", k, v )
			t.Fail()
		} else {
			//fmt.Printf( "foo: %s = %s\n", k, v )
			count++
		}
	}

	if count != 12 {
		fmt.Fprintf( os.Stderr, "didn't find enough elements in foo map, expected 12 found %d\n", count )
	}
}


/*
	Should generate a map with only bar tagged items
*/
func TestBarOnly( t *testing.T ) {
	fmt.Fprintf( os.Stderr, "testbar only\n" )
	count := 0
	fs := mk_struct1()
	m := transform.Struct_to_map( fs, "Bar" )		// should only generate foo elements into map
	for k, v := range m {
		if strings.Index( k, "Bar" ) != 0 {
			fmt.Fprintf( os.Stderr, "BAd: unexpected key found in 'bar' map: %s (%v)\n", k, v )
			t.Fail()
		} else {
			count++
		}
	}

	if count != 10 {
		fmt.Fprintf( os.Stderr, "didn't find enough elements in bar map, expected 10 found %d\n", count )
	}
}

/*
	Should generate a map with all exported items
*/
func TestAll( t *testing.T ) {
	fs := mk_struct1()
	bcount := 0
	fcount := 0
	zcount := 0

	fmt.Fprintf( os.Stderr, "testall\n" )
	m := transform.Struct_to_map( fs, "_" )		// should capture all fields
	for k, _ := range m {
		if strings.Index( k, "Foo" ) == 0 {
			fcount++
		} else {
			if strings.Index( k, "Bar" ) == 0 {
				bcount++
			} else {
				if strings.Index( k, "Baz" ) == 0 {
					zcount++
				}
			}
		}
	}

	if fcount != 12 || bcount != 10 || zcount != 2 {
		fmt.Fprintf( os.Stderr, "all test: unexpected count, expected 12/10/2 got %d/%d/%d\n", fcount, bcount, zcount )
		t.Fail()
	}
}

/*
	Generate a map, and then use it to populate an empty struct.
*/
func TestPopulate( t *testing.T ) {
	fmt.Fprintf( os.Stderr, "test populate\n" )
	ofs := mk_struct1( )									// make original struct
	fm := transform.Struct_to_map( ofs, "Foo" )				// generate foo map
	nfs := &Foo_bar{}										// empty strut to populate
	transform.Map_to_struct( fm, nfs,  "Foo" )				// fill it in

	if ofs.Foo_str != nfs.Foo_str {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_str did not match new: (%s) != (%s)\n", ofs.Foo_str, nfs.Foo_str )
		t.Fail()
	}
	if ofs.Foo_int != nfs.Foo_int {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_str did not match new: (%d) != (%d)\n", ofs.Foo_int, nfs.Foo_int )
		t.Fail()
	}

	if *ofs.Foo_intp1 != *nfs.Foo_intp1 {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_str did not match new: (%d) != (%d)\n", *ofs.Foo_intp1, *nfs.Foo_intp1 )
		t.Fail()
	}
	if ofs.Foo_intp1 == nfs.Foo_intp1 {			// pointers should NOT be the same
		fmt.Fprintf( os.Stderr, "FAIL: old and new intp1 pointers are the same and shouldn't be!\n" )
		t.Fail()
	}

	if *ofs.Foo_intp2 != *nfs.Foo_intp2 {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_str did not match new: (%d) != (%d)\n", *ofs.Foo_intp2, *nfs.Foo_intp2 )
		t.Fail()
	}
	if ofs.Foo_uint != nfs.Foo_uint {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_str did not match new: (%d) != (%d)\n", ofs.Foo_uint, nfs.Foo_uint )
		t.Fail()
	}
	if *ofs.Foo_uintp1 != *nfs.Foo_uintp1 {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_str did not match new: (%d) != (%d)\n", *ofs.Foo_uintp1, *nfs.Foo_uintp1 )
		t.Fail()
	}
	if *ofs.Foo_uintp2 != *nfs.Foo_uintp2 {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_str did not match new: (%d) != (%d)\n", *ofs.Foo_uintp2, *nfs.Foo_uintp2 )
		t.Fail()
	}
	if ofs.Foo_bool != nfs.Foo_bool {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_str did not match new: (%v) != (%v)\n", ofs.Foo_bool, nfs.Foo_bool )
		t.Fail()
	}
	if *ofs.Foo_boolp != *nfs.Foo_boolp {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_str did not match new: (%v) != (%v)\n", *ofs.Foo_boolp, *nfs.Foo_boolp )
		t.Fail()
	}
}
