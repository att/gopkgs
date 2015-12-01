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

/*
	Accept a structure and build a map from it's values. The map
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
	return value_to_map( thing, imeta, tag_id, m, "" )
}

/*
	This is the work horse which can call itself to process nested structs.
*/
func value_to_map( thing reflect.Value, imeta reflect.Type, tag_id string, m map[string]string, pfx string ) ( map[string]string ) {

	if thing.Kind() != reflect.Struct {
		return m
	}
	
	if m == nil {
		m = make( map[string]string )	
	}

	for i := 0; i < thing.NumField(); i++ {
		f := thing.Field( i )					// get the value of the ith field
		fmeta := imeta.Field( i )				// get the ith field's metadata from Type
		ftag := fmeta.Tag.Get( tag_id ) 		// get the field's datacache tag
		if ftag == "_" || tag_id == "_" {
			ftag = fmeta.Name
		}

		fkind := f.Kind()
		if ftag != "" || fmeta.Anonymous {		// process all structs regardless of tag
			ftag = pfx + ftag

			switch fkind {
				case reflect.String:
					m[ftag] = fmt.Sprintf( "%s", f )

				case reflect.Ptr:
					p := f.Elem()
					switch p.Kind() {
						case reflect.String:
							m[ftag] = fmt.Sprintf( "%s", p )

						case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
							m[ftag] = fmt.Sprintf( "%d", p )

						case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
							m[ftag] = fmt.Sprintf( "%d", p )

						case reflect.Float64, reflect.Float32:
							m[ftag] = fmt.Sprintf( "%f", p )

						case reflect.Bool:
							m[ftag] = fmt.Sprintf( "%v", p)

						case reflect.Struct:
							value_to_map( p, p.Type(), tag_id, m, pfx + fmeta.Name + "/" )	// recurse to process with a prefix which matches the field
					}
					
				case reflect.Uintptr:
					m[ftag] = fmt.Sprintf( "%d", f )

				case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
					m[ftag] = fmt.Sprintf( "%d", f )

				case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
					m[ftag] = fmt.Sprintf( "%d", f )

				case reflect.Float64, reflect.Float32:
					m[ftag] = fmt.Sprintf( "%f", f )

				case reflect.Bool:
					m[ftag] = fmt.Sprintf( "%v", f )

				case reflect.Struct:
					if fmeta.Anonymous {
						value_to_map( f, f.Type(), tag_id, m, pfx )			// recurse to process; only anonymous fields as they share this level namespace
					} else {
						value_to_map( f, f.Type(), tag_id, m, pfx + fmeta.Name + "/" )	// recurse to process; only anonymous fields as they share this level namespace
					}

				default:
					fmt.Fprintf( os.Stderr, "transform:stm: field %d cannot be captured in map: tag=%s type=%s val=%s\n", i, ftag, f.Kind(), reflect.ValueOf( f ) )
			}	
		}
	}

	return m
}
