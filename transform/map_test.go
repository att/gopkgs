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
	Foo_thing_s string			`Foo:"_"`
	Bar_thing_s string			`Bar:"_"`
}

type Foo_bar struct {
						Adjunct				// annon structure (tagged fields will be taken)

	Athingp				*Thing	`Foo:"_"`	// only picked up for foo and all
	Athing				Thing	`Bar:"_"`	// only picked up for bar and all	
	Foo_str string		`Foo:"FooString"`
	Foo_int int			`Foo:"FooInteger"`
	Foo_intp1 *int  	`Foo:"FooPtr1"`
	Foo_intp2 *int  	`Foo:"FooPtr2"`
	Foo_uint uint		`Foo:"_"`
	Foo_uintp1 *uint16	`Foo:"_"`
	Foo_uintp2 *uint16	`Foo:"_"`
	Foo_bool bool		`Foo:"_"`
	Foo_boolp *bool		`Foo:"_"`
										// 12 foo things at outer level (9 here and 2 in adjunct)

	Foo_array []int					`Foo:"_"`
	Foo_arrayp []*int				`Foo:"_"`
	Foo_arraym []map[string]int		`Foo:"_"`
	Foo_mapp map[string]*Thing		`Foo:"_"`
	Foo_mapi map[string]int			`Foo:"_"`

	Bar_str string		`Bar:"BarString"`
	Bar_int int			`Bar:"BarInteger"`
	Bar_intp1 *int  	`Bar:"BarPtr1"`
	Bar_intp2 *int  	`Bar:"BarPtr2"`
	Bar_uint uint		`Bar:"_"`
	Bar_uintp1 *uint16	`Bar:"_"`
	Bar_uintp2 *uint16	`Bar:"_"`
	Bar_bool bool		`Bar:"_"`
	Bar_boolp *bool		`Bar:"_"`
										// 10 bar things; 9 here and 1 in adjunct

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

	// array/map things are only tagged with foo tags, and should appear only when we do all or foo
	// based transformation.  Counts of expected elements are listed with each group

	fbs.Foo_array = make( []int, 0, 10 )			// will generate capacity and len elements so +2
	fbs.Foo_array = append( fbs.Foo_array, 1 )
	fbs.Foo_array = append( fbs.Foo_array, 3 )
	fbs.Foo_array = append( fbs.Foo_array, 5 )
													// 3 elements

	fbs.Foo_arrayp = make( []*int, 0, 17 )			// will generate capacity and len elements so +2
	iv1 := 11
	iv2 := 13
	iv3 := 15
	fbs.Foo_arrayp = append( fbs.Foo_arrayp, &iv1 )
	fbs.Foo_arrayp = append( fbs.Foo_arrayp, &iv2 )
	fbs.Foo_arrayp = append( fbs.Foo_arrayp, &iv3 )
													// 3 elements

	fbs.Foo_mapp = make( map[string]*Thing )
	fbs.Foo_mapi = make( map[string]int )
	fbs.Foo_mapi["one"] = 1
	fbs.Foo_mapi["three"] = 3
	fbs.Foo_mapi["seven"] = 7
	fbs.Foo_mapi["thirteen"] = 13					// 4 elements

	tp := &Thing{ Foo_thing_s: "Foo-thing-s-in-map1", Bar_thing_s: "bar-thing-s-in-map1"  }  // 1 element
	fbs.Foo_mapp["one"] = tp

	tp = &Thing{ Foo_thing_s: "foo-thing-s-in-map2", Bar_thing_s: "bar-thing-s-in-map2"  }	// 1 element
	fbs.Foo_mapp["ninety"] = tp

	fbs.Foo_arraym = make( []map[string]int, 0, 10 )	// array of maps, will add count and cap elements so +2
	mp := make( map[string]int )
	mp["bush_225"] = 1978
	mp["bush_209"] = 1979
	mp["biddle_318"] = 1980
	mp["biddle_409"] = 1981							// 4 elements
	fbs.Foo_arraym = append( fbs.Foo_arraym, mp )

	mp = make( map[string]int )
	mp["grapevine"] = 1982
	mp["dallas"] = 1984
	mp["dallas_frank2"] = 1985						// 3 elements
	fbs.Foo_arraym = append( fbs.Foo_arraym, mp )

	mp = make( map[string]int )
	mp["vienna"] = 1992
	mp["dallas"] = 1993
	mp["dallas_whisp"] = 1995						// 3 elements
	fbs.Foo_arraym = append( fbs.Foo_arraym, mp )
													// 28 slice/map things to check for

	fbs.Foo_AJ1 = 10001
	fbs.Foo_AJ2 = 10002
	fbs.Foo_AJ3 = 10003

	fbs.Athingp = &Thing{}
	fbs.Athingp.Foo_thing_s = "hello foo from a pointer to struct field"
	fbs.Athingp.Bar_thing_s = "hello bar from a pointer to struct field"

	fbs.Athing = Thing{}
	fbs.Athing.Foo_thing_s = "foo thing inline"
	fbs.Athing.Bar_thing_s = "bar thing inline"

	return fbs
}
		

/*
	Should generate a map with only foo tagged items
*/
func TestFooOnly( t *testing.T ) {
	fs := mk_struct1()
	count := 0
	acount := 0
	apcount := 0

	fmt.Fprintf( os.Stderr, "test: capture only foo fields\n" )
	m := transform.Struct_to_map( fs, "Foo" )								// should only generate foo elements into map
	for k, v := range m {
		fmt.Fprintf( os.Stderr, "checking foo: %s (%v)\n", k, v )			// uncomment to dump the map

		if strings.Index( k, "Foo" ) != 0 {									// count things; see struct and build function for expected counts
			if strings.Index( k, "Athingp/Foo_" ) != 0 {
				if strings.Index( k, "Athing/Foo_" ) != 0 {
					fmt.Fprintf( os.Stderr, "BAD: unexpected key found in 'foo' map: %s (%v)\n", k, v )		// all fields should start with Foo_ or Athingp/Foo_
					t.Fail()
				} else {
					acount++
				}
			} else {
				apcount++
			}
		} else {
			count++
		}
	}

	expect_count := 12 + 28				// top level + array and map elements (don't forget cap and len elements for each array)
	expect_apcount := 1					// pointer to thing struct elements (athing and athingp)
	expect_acount := 0					// direct struct thing elements
	if count != expect_count || apcount != expect_apcount || acount != expect_acount {			// should only find the athing pointer referenced thing
		fmt.Fprintf( os.Stderr, "didn't find right number of elements in foo map, expected %d Foo_, %d apointer things, and %d athings; found foo=%d apthing=%d athing=%d\n", 
			expect_count, expect_apcount, expect_acount, count, apcount, acount )
		for k, v := range m {
			fmt.Fprintf( os.Stderr, "m[%s] = (%s)\n", k, v )
		}
	}
}


/*
	Should generate a map with only bar tagged items
*/
func TestBarOnly( t *testing.T ) {
	fmt.Fprintf( os.Stderr, "test: capture only bar fields\n" )
	count := 0
	acount := 0
	apcount := 0

	fs := mk_struct1()
	m := transform.Struct_to_map( fs, "Bar" )		// should only generate bar elements into map

	for k, v := range m {
		//fmt.Fprintf( os.Stderr, "checking bar: %s = (%s)\n", k, v )			//uncomment to dump as we check things
		if strings.Index( k, "Bar" ) != 0 {
			if strings.Index( k, "Athingp/Bar_" ) != 0 {
				if strings.Index( k, "Athing/Bar_" ) != 0 {
					fmt.Fprintf( os.Stderr, "BAD: unexpected key found in 'bar' map: %s (%v)\n", k, v )
					t.Fail()
				} else {
					acount++
				}
			} else {
				apcount++
			}
		} else {
			//fmt.Printf( "bar: %s = %s\n", k, v )
			count++
		}
	}

	expect_count := 10 + 0				// top level + no array things for bar
	expect_apcount := 0					// pointer to thing struct elements (athing and athingp)
	expect_acount := 1					// direct struct thing elements
	if count != expect_count || apcount != expect_apcount || acount != expect_acount {			// should only find the athing pointer referenced thing
		fmt.Fprintf( os.Stderr, "didn't find right number of elements in bar map, expected %d Foo_, %d apointer things, and %d athings; found bar=%d apthing=%d athing=%d\n", 
			expect_count, expect_apcount, expect_acount, count, apcount, acount )
		for k, v := range m {
			fmt.Fprintf( os.Stderr, "m[%s] = (%s)\n", k, v )
		}
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
	athingp := 0
	athing := 0

	fmt.Fprintf( os.Stderr, "test: capture all fields in map\n" )
	m := transform.Struct_to_map( fs, "_" )		// should capture all fields
	for k, _ := range m {
		//fmt.Fprintf( os.Stderr, "k=(%s) v=(%s)\n", k, v )
		if strings.Index( k, "Foo" ) == 0 {
			fcount++
		} else {
			if strings.Index( k, "Bar" ) == 0 {
				bcount++
			} else {
				if strings.Index( k, "Baz" ) == 0 {
					zcount++
				} else {
					if strings.Index( k, "Athingp" ) == 0 {
						athingp++
					} else {
						if strings.Index( k, "Athing" ) == 0 {
							athing++
						}
					}
				}
			}
		}
	}

	expect_fcount := 12 + 30	// two additional fields captured in foo space when doing all
	expect_bcount := 10
	expect_zcount := 2
	expect_acount := 2
	expect_apcount := 2
	if fcount != expect_fcount || bcount != expect_bcount || zcount != expect_zcount || athing != expect_acount || athingp != expect_apcount {
		fmt.Fprintf( os.Stderr, "all test: unexpected count, expected %d/%d/%d/%d/%d got %d/%d/%d/%d/%d\n", 
			expect_fcount, expect_bcount, expect_zcount, expect_acount, expect_apcount,
			fcount, bcount, zcount, athingp, athing )
		for k, v := range m {
			fmt.Fprintf( os.Stderr, "m[%s] = (%s)\n", k, v )
		}
		t.Fail()
	}
}


/*
	Run two maps which we expect to be equal
*/
func map_ck( m1 map[string]string, m2 map[string]string ) ( pass bool ) {
	pass = true

	for k, v := range m1 {
		if m2[k] == "" {
			fmt.Fprintf( os.Stderr, "key %s not in m2\n", k )
			pass = false
		} else {
			if v != m2[k] {
				fmt.Fprintf( os.Stderr, "values for key %s do not match", k )
				pass = false
			}
		}
	}

	
	for k, _ := range m2 {		// just need to check for missing from m1
		if m1[k] == "" {
			fmt.Fprintf( os.Stderr, "key %s not in m1\n", k )
			pass = false
		}
	}

	return pass
}

/*
	Run two maps of integers which we expect to be the same
*/
func imap_ck( m1 map[string]int, m2 map[string]int ) ( pass bool ) {
	pass = true

	for k, v := range m1 {
		if m2[k] == 0 {
			fmt.Fprintf( os.Stderr, "key %s not in m2\n", k )
			pass = false
		} else {
			if v != m2[k] {
				fmt.Fprintf( os.Stderr, "values for key %s do not match", k )
				pass = false
			}
		}
	}

	
	for k, _ := range m2 {		// just need to check for missing from m1
		if m1[k] == 0 {
			fmt.Fprintf( os.Stderr, "key %s not in m1\n", k )
			pass = false
		}
	}

	return pass
}

/*
	Run two maps which are pointers to structs.
	This supports the populate test which only populates based on foo
	tagged things, so we need to ensure that bar things are nil and foo things
	are equal.
*/
func tmap_ck( m1 map[string]*Thing, m2 map[string]*Thing ) ( pass bool ) {
	pass = true

	for k, v := range m1 {
		if m2[k] == nil {
			fmt.Fprintf( os.Stderr, "key %s not in m2\n", k )
			pass = false
		} else {
			if v.Foo_thing_s != m2[k].Foo_thing_s {
				fmt.Fprintf( os.Stderr, "values for key %s do not match", k )
				pass = false
			}
			if m2[k].Bar_thing_s != "" {			// bar strings should be empty
				fmt.Fprintf( os.Stderr, "bar string for key %s was not empty: (%s)\n", k, m2[k].Bar_thing_s )
				pass = false
			}
		}
	}

	for k, _ := range m2 {		// just need to check for missing from m1
		if m1[k] == nil {
			fmt.Fprintf( os.Stderr, "key %s not in m1\n", k )
			pass = false
		}
	}

	return pass
}

func array_ck( t1 interface{}, t2 interface{}, name string ) ( pass bool ) {
	pass = true

	switch t1a := t1.( type ) {
		case []*int:
			if t2a, ok := t2.( []*int ); ok {
				ol := len( t1a )
				oc := cap( t1a )
				nl := len( t2a )
				nc := cap( t2a )
				if ol != nl  ||  oc != nc {
					fmt.Fprintf( os.Stderr, "len and caps don't match for %s:  old: %d/%d  new: %d/%d\n", name, ol, oc, nl, nc )
					return false
				} else {
					for j := 0; j < ol; j++ {
						if t1a[j] == t2a[j] {				// pointers should not be the same
							fmt.Fprintf( os.Stderr, "pointers at %d  match and they shouldn't for %s\n", j, name )
							pass = false
						} else {
							if *(t1a[j]) != *(t2a[j]) {
								fmt.Fprintf( os.Stderr, "element %d does not match: %d != %d\n", j, *t1a[j], *t2a[j] )
								pass = false
							}
						}
					}
				}	
			} else {
				fmt.Fprintf( os.Stderr, "%s: array types don't match\n", name )
				pass = false
			}
		case []int:
			if t2a, ok := t2.( []int ); ok {
				ol := len( t1a )
				oc := cap( t1a )
				nl := len( t2a )
				nc := cap( t2a )
				if ol != nl  ||  oc != nc {
					fmt.Fprintf( os.Stderr, "len and caps don't match for %s:  old: %d/%d  new: %d/%d\n", name, ol, oc, nl, nc )
					return false
				} else {
					for j := 0; j < ol; j++ {
						if t1a[j] != t2a[j] {
							fmt.Fprintf( os.Stderr, "element %d does not match: %d != %d\n", j, t1a[j], t2a[j] )
							return false
						}
					}
				}	
			} else {
				fmt.Fprintf( os.Stderr, "%s: array types don't match\n", name )
				pass = false
			}

		case []map[string]string:				// we'll only test this one specific case assuming a map is a map :)
			if t2a, ok := t2.( []map[string]string ); !ok {
				fmt.Fprintf( os.Stderr, "%s: array types don't match expected map[string]string\n", name )
				pass = false
			} else {
				ol := len( t1a )
				oc := cap( t1a )
				nl := len( t2a )
				nc := cap( t2a )
				if ol != nl  ||  oc != nc {
					fmt.Fprintf( os.Stderr, "len and caps don't match for %s:  old: %d/%d  new: %d/%d\n", name, ol, oc, nl, nc )
					return false
				}

				for j := 0; j < ol; j++ {
					if ! map_ck( t1a[j], t2a[j] )  {
						fmt.Fprintf( os.Stderr, "maps at element %d don't validate, see earlier message(s) name=%s\n", j, name )
						pass = false
					}
				}
			}

	}

	return pass
}

/*
	Generate a map, and then use it to populate an empty struct.
*/
func TestPopulate( t *testing.T ) {
	fmt.Fprintf( os.Stderr, "test: populate struct from map\n" )
	ofs := mk_struct1( )									// make original struct
	fm := transform.Struct_to_map( ofs, "Foo" )				// generate foo map
	nfs := &Foo_bar{}										// empty strut to populate

	if ofs.Foo_AJ1 == nfs.Foo_AJ1 {						// spot check to ensure that the struct to fill is really 'empty'
		fmt.Fprintf( os.Stderr, "FAIL: old foo_aj1 matches new before transform: (%d)  != (%d)\n", ofs.Foo_AJ1, nfs.Foo_AJ1 )
		t.Fail()
	} 

	if nfs.Athingp != nil {
		fmt.Fprintf( os.Stderr, "FAIL: athing in new object is not nil before transform\n" )
		t.Fail()
	}

	transform.Map_to_struct( fm, nfs,  "Foo" )				// fill it in just basedon foo tags

	if nfs.Athingp == nil {
		fmt.Fprintf( os.Stderr, "FAIL: athing in new object is sitll nil after transform\n" )
		t.Fail()
	}

	if ofs.Athingp.Foo_thing_s != nfs.Athingp.Foo_thing_s {
		fmt.Fprintf( os.Stderr, "FAIL: old athingp.foo_thing_s did not match new: (%s) != (%s)\n", ofs.Athingp.Foo_thing_s, nfs.Athingp.Foo_thing_s )
		t.Fail()
	} 

	if ofs.Athing.Foo_thing_s == nfs.Athing.Foo_thing_s {			// athing isn't foo tagged, so these should NOT match
		fmt.Fprintf( os.Stderr, "FAIL: old athing.foo_thing_s  DID match new and shouldn't: (%s) != (%s)\n", ofs.Athing.Foo_thing_s, nfs.Athing.Foo_thing_s )
		t.Fail()
	} 

	if ofs.Foo_AJ1 != nfs.Foo_AJ1 {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_aj1 did not match new: (%d) != (%d)\n", ofs.Foo_AJ1, nfs.Foo_AJ1 )
		t.Fail()
	} 

	if ofs.Foo_AJ2 != nfs.Foo_AJ2 {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_aj1 did not match new: (%d) != (%d)\n", ofs.Foo_AJ2, nfs.Foo_AJ2 )
		t.Fail()
	} 

	if ofs.Foo_AJ3 != nfs.Foo_AJ3 {
		fmt.Fprintf( os.Stderr, "FAIL: old foo_aj1 did not match new: (%d) != (%d)\n", ofs.Foo_AJ3, nfs.Foo_AJ3 )
		t.Fail()
	} 

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

	// ---------- validate contents of arrays -----------------------------
	if ! array_ck( ofs.Foo_array, nfs.Foo_array, "foo_array" ) {
		t.Fail()
	}

	if ! array_ck( ofs.Foo_arrayp, nfs.Foo_arrayp, "foo_array" ) {
		t.Fail()
	}

	// ---------- validate contents of maps --------------------------------
	if ! imap_ck( ofs.Foo_mapi, nfs.Foo_mapi ) {
		fmt.Fprintf( os.Stderr, "map foo_mapi doesn't validate; see above messages\n" )
		t.Fail()
	}

	if ! tmap_ck( ofs.Foo_mapp, nfs.Foo_mapp ) {
		fmt.Fprintf( os.Stderr, "map foo_mapp doesn't validate; see above messages\n" )
		t.Fail()
	}
}
