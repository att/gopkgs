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
        Mnemonic:       tomap.go
        Absrtract:      Various functions which transform something into a map.
		Date:			25 November 2015
		Author:			E. Scott Daniels
*/

package transform

import (
	"fmt"
	"os"
	"reflect"
)

var (
	depth int = 0
)

/*
	Accept a structure and build a map from its values. The map
	is [string]string, and the keys are taken from fields tagged with
	tags that match the tag_id string passed in. 
	
	If the tag id is map, then a tag might be `map:"xyz"` where xyz is used
	as the name in the map, or `map:"_"` where the structure field name is 
	used in the map. If the tag id "_" is passed in, then all fields in the 
	structure are captured. 

	This function will capture all simple fields (int, bool, float, etc.) and 
	structures, anonynmous structures and pointers to structures. It will _NOT_
	capture arrays or maps.
*/
func Struct_to_map( ustruct interface{}, tag_id string ) ( m map[string]string ) {
	var imeta reflect.Type
	var thing reflect.Value

	thing = reflect.ValueOf( ustruct )			// thing is the actual usr struct (in reflect container)
	if thing.Kind() == reflect.Ptr {
		thing = thing.Elem()					// deref the pointer to get the real container
		imeta = thing.Type() 					// snag the type allowing for extraction of meta data
	} else {
		imeta = reflect.TypeOf( thing )			// convert input to a Type allowing for extraction of meta data
	}

	m = make( map[string]string )	
	return struct_to_map( thing, imeta, tag_id, m, "" )
}

func dec_depth() {
	if depth > 0 {
		depth--
	}
}

/*
	Insert a value into the map using its 'tag'.  If the value is a struct, map or array (slice)
	then we recurse to insert each component (array/map) and invoke the struct function to 
	dive into the struct extracting tagged fields.
*/
func insert_value( thing reflect.Value, tkind reflect.Kind, anon bool, tag string, tag_id string, m map[string]string, pfx string ) {
	depth++
	if depth > 50 {
		fmt.Fprintf( os.Stderr, "recursion to deep in transform insert value: %s\n%s\n", pfx, tkind ) 
		os.Exit( 1 )
	}
	defer dec_depth()

	tag = pfx + tag
	switch tkind {
		case reflect.String:
			m[tag] = fmt.Sprintf( "%s", thing )

		case reflect.Ptr:
			p := thing.Elem()
			switch p.Kind() {
				case reflect.String:
					m[tag] = fmt.Sprintf( "%s", p )

				case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
					m[tag] = fmt.Sprintf( "%d", p.Int() )

				case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
					m[tag] = fmt.Sprintf( "%d", p.Uint() )

				case reflect.Float64, reflect.Float32:
					m[tag] = fmt.Sprintf( "%f", p.Float() )

				case reflect.Bool:
					m[tag] = fmt.Sprintf( "%v", p.Bool() )

				case reflect.Struct:
					struct_to_map( p, p.Type(), tag_id, m, pfx + tag + "/"  )	// recurse to process with a prefix which matches the field

				default:
					fmt.Fprintf( os.Stderr, "transform: ptr of this kind is not handled: %s\n", p.Kind() )
			}
			
		case reflect.Uintptr:
			p := thing.Elem()
			m[tag] = fmt.Sprintf( "%d", p.Uint() )

		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			m[tag] = fmt.Sprintf( "%d", thing.Int() )

		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			m[tag] = fmt.Sprintf( "%d", thing.Uint() )

		case reflect.Float64, reflect.Float32:
			m[tag] = fmt.Sprintf( "%f", thing.Float() )

		case reflect.Bool:
			m[tag] = fmt.Sprintf( "%v", thing.Bool() )

		case reflect.Struct:
			if anon {
				struct_to_map( thing, thing.Type(), tag_id, m, pfx )				// recurse to process; anonymous fields share this level namespace
			} else {
				struct_to_map( thing, thing.Type(), tag_id, m, tag + "/" )			// recurse to process; tag becomes the prefix
			}

		case reflect.Slice:
			length := thing.Len()
			capacity := thing.Cap()
			for j := 0; j <  length; j++ {
				v := thing.Slice( j, j+1 )								// value struct for the jth element (which will be a slice too :(
				vk := v.Type().Elem().Kind()							// must drill down for the kind
				mtag := fmt.Sprintf( "%s%s.cap", pfx, tag )
				m[mtag] = fmt.Sprintf( "%d", capacity )
				mtag = fmt.Sprintf( "%s%s.len", pfx, tag )
				m[mtag] = fmt.Sprintf( "%d", length )
				new_tag := fmt.Sprintf( "%s/%d", tag, j )
				insert_value(  thing.Index( j ), vk, false, new_tag, tag_id, m, pfx  )		// prefix remains the same, the 'index' changes as we go through the slice
			}

		case reflect.Map:
			keys := thing.MapKeys()						// list of all of the keys (values)
			for _, key := range keys {
				vk := thing.Type().Elem().Kind()			// must drill down for the kind
				new_tag := fmt.Sprintf( "%s/%s", tag, key )
				insert_value(  thing.MapIndex( key ), vk, false, new_tag, tag_id, m, pfx  )					// prefix stays the same, just gets a new tag
			}

		case reflect.Interface:
			p := thing.Elem()
			insert_value( p, p.Kind(), anon, tag, tag_id, m, pfx )

		default:
			//fmt.Fprintf( os.Stderr, "transform: stm: field cannot be captured in a map: tag=%s type=%s val=%s\n", tag, thing.Kind(), reflect.ValueOf( thing ) )
	}	
}


/*
	We require the initial 'thing' passed to be a struct, this then is the real entry point, but 
	is broken so that it can be recursively called as we encounter structs in the insert code.
*/
func struct_to_map( thing reflect.Value, imeta reflect.Type, tag_id string, m map[string]string, pfx string ) ( map[string]string ) {

	if thing.Kind() != reflect.Struct {
		return m
	}
	
	if m == nil {
		m = make( map[string]string )	
	}

	for i := 0; i < thing.NumField(); i++ {
		f := thing.Field( i )					// get the _value_ of the ith field
		fmeta := imeta.Field( i )				// get the ith field's metadata from Type (a struct_type)
		ftag := fmeta.Tag.Get( tag_id ) 		// get the field's datacache tag
		if ftag == "_" || tag_id == "_" {
			ftag = fmeta.Name
		}

		if ftag != "" || fmeta.Anonymous {		// process all structs regardless of tag
			insert_value( f, f.Kind(), fmeta.Anonymous,  ftag, tag_id, m, pfx )
		}
	}

	return m
}
