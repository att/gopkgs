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
        Mnemonic:       tostruct.go
        Absrtract:      Various functions which transform something into a struct.
		Date:			25 November 2015
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
	Accepts a map of [string]string and attempts to populate the fields in the 
	user struct (pointer) from the values in the map. Only the fields in the 
	struct which are tagged with the given tag ID are given are affected. If the 
	tag ID is '_' (underbar), then all fields which are external/exported are 
	potentially affected.  The struct tag takes the format:
		tag_id:"tagstr"
	where tagstr is used as the field name in the map to look up.  If tagstr
	is given as an '_' (underbar) character, then the structure field name is
	used.   Names are case sensitive.

	This function supports transferring the simple types (bool, int, float, etc.) and 
	pointers to those types from the map.  It also supports structures, anonymous
	structures and pointers to structures. It does not support maps or arrays.
*/
func Map_to_struct( m map[string]string, ustructp interface{}, tag_id string ) ( ) {
	
	thing := reflect.ValueOf( ustructp ).Elem() // get a reference to the struct
	tmeta := thing.Type()						// meta data for the struct
	
	map_to_struct( m, thing, tmeta, tag_id, "" )
}

/*
	Given a thing (field value in a struct), set the thing with the element in the map (key)
	the map value to the proper type.
*/
func set_value( thing reflect.Value, kind reflect.Kind, key string, tag_id string, pfx string, annon bool, m map[string]string ) {

	if ! thing.CanAddr( ) {			// prevent stack dump
		return
	}

	switch kind {
		default:
			fmt.Fprintf( os.Stderr, "transform.mts: tagged sturct member cannot be converted from map: tag=%s kind=%v\n", key, thing.Kind() )

		case reflect.String:
				thing.SetString( m[key] )

		case reflect.Ptr:
			p := thing.Elem()								// get the pointer value; allows us to suss the type
			if ! p.IsValid() {							// ptr is nill in the struct so we must allocate a pointer to 0 so it can be changed below
				thing.Set( reflect.New(thing.Type().Elem()) )
				p = thing.Elem()
			}
			switch p.Kind() {
				case reflect.String:
					s := m[key]						// copy it and then point to the copy
					thing.Set( reflect.ValueOf(  &s ) )

				case reflect.Int:
					i := clike.Atoi( m[key] )				// convert to integer and then point at the value
					thing.Set( reflect.ValueOf(  &i ) )

				case  reflect.Int64:
					i := clike.Atoi64( m[key] )
					thing.Set( reflect.ValueOf(  &i ) )

				case  reflect.Int32:
					i := clike.Atoi32( m[key] )
					thing.Set( reflect.ValueOf(  &i ) )

				case  reflect.Int16:
					i := clike.Atoi16( m[key] )
					thing.Set( reflect.ValueOf(  &i ) )

				case  reflect.Int8:
					i := int8( clike.Atoi16( m[key] ) )
					thing.Set( reflect.ValueOf(  &i ) )

				case reflect.Uint:
					ui := clike.Atou( m[key] )
					thing.Set( reflect.ValueOf(  &ui ) )

				case reflect.Uint64:
					ui := clike.Atou64( m[key] )
					thing.Set( reflect.ValueOf(  &ui ) )

				case reflect.Uint32:
					ui := clike.Atou32( m[key] )
					thing.Set( reflect.ValueOf(  &ui ) )

				case reflect.Uint16:
					ui := clike.Atou16( m[key] )
					thing.Set( reflect.ValueOf(  &ui ) )

				case reflect.Uint8:
					ui := uint8( clike.Atou16( m[key] ) )
					thing.Set( reflect.ValueOf(  &ui ) )

				case reflect.Float64:
					fv := clike.Atof( m[key] )
					thing.Set( reflect.ValueOf(  &fv ) )

				case  reflect.Float32:
					fv := float32( clike.Atof( m[key] ) )
					thing.Set( reflect.ValueOf(  &fv ) )

				case reflect.Bool:
					b := m[key] == "true" || m[key] == "True" || m[key] == "TRUE"
					thing.Set( reflect.ValueOf(  &b ) )

				case reflect.Struct:
					map_to_struct( m, p, p.Type(), tag_id, pfx  )			// recurse to process
			}
			
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			thing.SetInt( clike.Atoi64( m[key] ) )

		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			thing.SetUint( uint64( clike.Atoi64( m[key] ) ) )

		case reflect.Float64, reflect.Float32:
			thing.SetFloat( clike.Atof( m[key] ) )

		case reflect.Bool:
			thing.SetBool(  m[key] == "true" )

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

					set_value( mthing, mthing.Kind(), idx + tokens[0], tag_id, idx + tokens[0] + "/", false, m  )	// put the value into the map thing
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
				set_value(  thing.Index( j ), thing.Type().Elem().Kind(), idx, tag_id, idx + "/", false, m  )	// populate each value of the slice up to len
			}

		case reflect.Struct:
			if annon {
				map_to_struct( m, thing, thing.Type(), tag_id, key )			// anon structs share namespace, so prefix is the same
			} else {
				map_to_struct( m, thing, thing.Type(), tag_id, pfx )			// dive to get the substruct adding a level to the prefix
			}

	}

}


/*
	Real work horse which can recurse down to process anon structs.
	Prefix (pfx) allows us to manage nested structs.
*/
func map_to_struct( m map[string]string, thing reflect.Value, tmeta reflect.Type, tag_id string, pfx string ) ( ) {

	for i := 0; i <  thing.NumField(); i++ {	// try all fields (they must be external!)
		f := thing.Field( i )					// get the value of the ith field
		fmeta := tmeta.Field( i )				// get the meta data for field i
		ftag := fmeta.Tag.Get( tag_id ) 		// get the field's datacache tag
		if ftag == "_" || tag_id == "_" {
			ftag = fmeta.Name
		}

		fkind := f.Kind()

		//if (fkind == reflect.Struct || ftag != "" ) && f.CanAddr()  {			// if there was a datacache tag, then attempt to pull the field from the map
		if (fmeta.Anonymous || ftag != "" ) && f.CanAddr()  {			// if there was a datacache tag, then attempt to pull the field from the map
			ftag = pfx + ftag

			set_value( f, fkind, ftag, tag_id, pfx +  fmeta.Name + "/", fmeta.Anonymous, m )

		}
	}

	return
}
