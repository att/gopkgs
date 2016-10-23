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
        Mnemonic:       i_tostruct.go
        Absrtract:      Various functions which transform something into a struct given a map of
						interfaces rather than strings. This is slightly different than the string
						based functions as it requires the data type in the map to match the 
						data type in the struct where the string based verion will convert any 
						string 'value' to the struct type.  
		Date:			23 October 2016
		Author:			E. Scott Daniels
*/

package transform

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/att/gopkgs/clike"
)

/*
	Accepts a map of [string]interface{}, and attempts to populate the fields in the 
	user struct (pointer) from the values in the map. Only the fields in the 
	struct which are tagged with the given tag ID are are affected. If the 
	tag ID is '_' (underbar), then all fields which are external/exported are 
	potentially affected.  The struct tag takes the format:
		tag_id:"tagstr"
	where tagstr is used as the field name in the map to look up.  If tagstr
	is given as an '_' (underbar) character, then the structure field name is
	used.   Names are case sensitive.

	This function supports transferring the simple types (bool, int, float, etc.) and 
	pointers to those types from the map.  It also supports structures, anonymous
	structures and pointers to structures, slices and maps.
*/
func Imap_to_struct( m map[string]interface{}, ustructp interface{}, tag_id string ) ( ) {

	thing := reflect.ValueOf( ustructp ).Elem() // get a reference to the struct
	tmeta := thing.Type()						// meta data for the struct
	
	imap_to_struct( m, thing, tmeta, tag_id, "" )
}

/*
	Given a thing (field value in a struct), set the thing with the element in the map (key)
	the map value to the proper type.
*/
func set_ivalue( thing reflect.Value, kind reflect.Kind, key string, tag_id string, pfx string, annon bool, m map[string]interface{} ) {

	if ! thing.CanAddr( ) {			// prevent stack dump
		return
	}

	switch kind {
		default:
			fmt.Fprintf( os.Stderr, "transform.mts: tagged sturct member cannot be converted from map: tag=%s kind=%v\n", key, thing.Kind() )

		case reflect.Interface:
				thing.Set( reflect.ValueOf( m[key] ) )

		case reflect.String:
				s, ok := m[key].(string)
				if ok {
					thing.SetString( s )
				}

		case reflect.Ptr:
			p := thing.Elem()								// get the pointer value; allows us to suss the type
			if ! p.IsValid() {							// ptr is nill in the struct so we must allocate a pointer to 0 so it can be changed below
				thing.Set( reflect.New(thing.Type().Elem()) )
				p = thing.Elem()
			}
			switch p.Kind() {
				case reflect.String:
					s, ok := m[key].(*string)		
					if ok {
						thing.Set( reflect.ValueOf(  s ) )
					}

				case reflect.Int:
					i, ok := m[key].(*int)		
					if ok {
						thing.Set( reflect.ValueOf(  i ) )
					}

				case  reflect.Int64:
					i, ok := m[key].(*int64)		
					if ok {
						thing.Set( reflect.ValueOf(  i ) )
					}

				case  reflect.Int32:
					i, ok := m[key].(*int32)		
					if ok {
						thing.Set( reflect.ValueOf(  i ) )
					}

				case  reflect.Int16:
					i, ok := m[key].(*int16)		
					if ok {
						thing.Set( reflect.ValueOf(  i ) )
					}

				case  reflect.Int8:
					i, ok := m[key].(*int8)		
					if ok {
						thing.Set( reflect.ValueOf(  i ) )
					}

				case reflect.Uint:
					ui, ok := m[key].(*uint)
					if ok {
						thing.Set( reflect.ValueOf(  ui ) )
					}

				case reflect.Uint64:
					ui, ok := m[key].(*uint64)		
					if ok {
						thing.Set( reflect.ValueOf(  ui ) )
					}

				case reflect.Uint32:
					ui, ok := m[key].(*uint32)		
					if ok {
						thing.Set( reflect.ValueOf(  ui ) )
					}

				case reflect.Uint16:
					ui, ok := m[key].(*uint16)		
					if ok {
						thing.Set( reflect.ValueOf(  ui ) )
					}

				case reflect.Uint8:
					ui, ok := m[key].(*uint8)		
					if ok {
						thing.Set( reflect.ValueOf(  ui ) )
					}

				case reflect.Float64:
					fv, ok := m[key].(*float64)		
					if ok {
						thing.Set( reflect.ValueOf(  fv ) )
					}

				case  reflect.Float32:
					fv, ok := m[key].(*float32)		
					if ok {
						thing.Set( reflect.ValueOf(  fv ) )
					}

				case reflect.Bool:
					b, ok := m[key].(*bool)		
					if ok {
						thing.Set( reflect.ValueOf(  b ) )
					}

				case reflect.Struct:
					imap_to_struct( m, p, p.Type(), tag_id, pfx  )			// recurse to process
			}

		case reflect.Int:
			i, ok := m[key].(int)
			if ok {
				thing.SetInt( int64( i ) )
			}
		case  reflect.Int64:
			i, ok := m[key].(int64)
			if ok {
				thing.SetInt( int64( i ) )
			}
		case  reflect.Int32:
			i, ok := m[key].(int32)
			if ok {
				thing.SetInt( int64( i ) )
			}
		case  reflect.Int16:
			i, ok := m[key].(int16)
			if ok {
				thing.SetInt( int64( i ) )
			}
		case  reflect.Int8:
			i, ok := m[key].(int8)
			if ok {
				thing.SetInt( int64( i ) )
			}
			
		case reflect.Uint:
			i, ok := m[key].(uint)
			if ok {
				thing.SetUint( uint64( i ) )
			}
		case  reflect.Uint64:
			i, ok := m[key].(uint64)
			if ok {
				thing.SetUint( uint64( i ) )
			}
		case  reflect.Uint32:
			i, ok := m[key].(uint32)
			if ok {
				thing.SetUint( uint64( i ) )
			}
		case  reflect.Uint16:
			i, ok := m[key].(uint16)
			if ok {
				thing.SetUint( uint64( i ) )
			}
		case  reflect.Uint8:
			i, ok := m[key].(uint8)
			if ok {
				thing.SetUint( uint64( i ) )
			}

		case reflect.Float64, reflect.Float32:
			i, ok := m[key].(float64)
			if ok {
				thing.SetFloat( i )
			}

		case reflect.Bool:
			i, ok := m[key].(bool)
			if ok {
				thing.SetBool( i )
			}

		case reflect.Map:
			new_map := reflect.MakeMap( thing.Type() ) 					// create the map
			thing.Set( new_map )								// put it in the struct

			idx := key + "/"									// now populate the map
			ilen := len( idx )
			for k, _ := range m {								// we could keep a separate list of keys, but for now this should do
				if strings.HasPrefix( k, key ) {
					tokens := strings.Split( k[ilen:], "/" )
					map_key := reflect.ValueOf( tokens[0] )				// map key is everything past the tag to the next slant
					map_ele_type := new_map.Type().Elem()				// the type of the element that the map references

					mthing := reflect.New( map_ele_type ).Elem() 		// new returns pointer, so dereference with Elem() (value not type!)

					set_ivalue( mthing, mthing.Kind(), idx + tokens[0], tag_id, idx + tokens[0] + "/", false, m  )	// put the value into the map thing
					new_map.SetMapIndex( map_key,  mthing ) 				// put it into the map (order IS important; add to map after recursion)
					//fmt.Fprintf( os.Stderr, "saving: %s  thing-type=%s mthing=%s\n", map_key, thing.Type(), mthing )
				}
			}

		case reflect.Slice:
			c := clike.Atoi( m[key + ".cap"] )
			l := clike.Atoi( m[key + ".len"] )

			thing.Set( reflect.MakeSlice( thing.Type(), l, c ) )		// create a new slice with the same len/cap that it had
			for j := 0; j < l; j++ {
				idx := fmt.Sprintf( "%s/%d", key, j )
				set_ivalue(  thing.Index( j ), thing.Type().Elem().Kind(), idx, tag_id, idx + "/", false, m  )	// populate each value of the slice up to len
			}

		case reflect.Struct:
			if annon {
				imap_to_struct( m, thing, thing.Type(), tag_id, key )			// anon structs share namespace, so prefix is the same
			} else {
				imap_to_struct( m, thing, thing.Type(), tag_id, pfx )			// dive to get the substruct adding a level to the prefix
			}

	}

}


/*
	Real work horse which can recurse down to process anon structs.
	Prefix (pfx) allows us to manage nested structs.
*/
func imap_to_struct( m map[string]interface{}, thing reflect.Value, tmeta reflect.Type, tag_id string, pfx string ) ( ) {

	for i := 0; i <  thing.NumField(); i++ {	// try all fields (they must be external!)
		f := thing.Field( i )					// get the value of the ith field
		fmeta := tmeta.Field( i )				// get the meta data for field i
		ftag := fmeta.Tag.Get( tag_id ) 		// get the field's datacache tag
		if ftag == "_" || tag_id == "_" {
			ftag = fmeta.Name
		}

		fkind := f.Kind()

		if (fmeta.Anonymous || ftag != "" ) && f.CanAddr()  {			// if there was a datacache tag, then attempt to pull the field from the map
			ftag = pfx + ftag

			set_ivalue( f, fkind, ftag, tag_id, pfx +  fmeta.Name + "/", fmeta.Anonymous, m )

		}
	}

	return
}
