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
	"reflect"
)

/*
	Accept a structure and build a map from it's values. The map
	is [string]string, and the keys are taken from fields tagged with
	datacahce: tags.   Only 'simple' fields are captured; structs, arrays,
	and other 'recursive' things are not.
*/
func Struct_to_map( ustruct interface{}, tag_id string ) ( m map[string]string ) {
	var imeta reflect.Type

	thing := reflect.ValueOf( ustruct )		// thing is the interface 'value'
	if thing.Kind() == reflect.Ptr {
		thing = thing.Elem()
		imeta = thing.Type() //reflect.TypeOf( thing )			// convert input to a Type allowing for extraction of meta data
	} else {
		imeta = reflect.TypeOf( thing )			// convert input to a Type allowing for extraction of meta data
	}

	if thing.Kind() != reflect.Struct {
		return nil
	}
	
	m = make( map[string]string )	
	for i := 0; i < thing.NumField(); i++ {
		f := thing.Field( i )					// get the value of the ith field
		fmeta := imeta.Field( i )				// get the ith field's metadata from Type
		ftag := fmeta.Tag.Get( tag_id ) 		// get the field's datacache tag
		//fmt.Printf( ">>>> k=%s field=(%s) tag=(%s) id=%s\n", thing.Kind(), fmeta.Name, ftag, tag_id )
		if ftag == "_" || tag_id == "_" {
			ftag = fmeta.Name
		}

		if ftag != "" {
			switch f.Kind() {
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

				default:
					fmt.Printf( "unkonw at %d\n", i )
			}	
		}
	}

	return m
}
