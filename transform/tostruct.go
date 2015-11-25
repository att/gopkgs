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

	Currently, only 'simple' types (bool, int, float, string, and pointers to them) 
	are supported (nested structs, arrays, maps, etc. are not supported). For
	boolean, the value is set to true if the map value is one of: true, True, or TRUE,
	and false otherwise.
*/
func Map_to_struct( m map[string]string, ustructp interface{}, tag_id string ) ( ) {
	
	thing := reflect.ValueOf( ustructp ).Elem() // get a reference to the struct
	tmeta := thing.Type()						// meta data for the struct
	
	for i := 0; i <  thing.NumField(); i++ {	// try all fields (they must be external!)
		f := thing.Field( i )					// get the value of the ith field
		fmeta := tmeta.Field( i )				// get the meta data for field i
		ftag := fmeta.Tag.Get( tag_id ) 		// get the field's datacache tag
		if ftag == "_" || tag_id == "_" {
			ftag = fmeta.Name
		}

		if ftag != "" && f.CanAddr()  {			// if there was a datacache tag, then attempt to pull the field from the map
			switch f.Kind() {
				default:
					fmt.Fprintf( os.Stderr, "tagged sturct member cannot be converted from map: tag=%s kind=%v", ftag, f.Kind() )

				case reflect.String:
						f.SetString( m[ftag] )

				case reflect.Ptr:
					p := f.Elem()								// get the pointer value; allows us to suss the type
					if ! p.IsValid() {							// ptr is nill in the struct so we must allocate a pointer to 0 so it can be changed below
						f.Set( reflect.New(f.Type().Elem()) )
						p = f.Elem()
					}
					switch p.Kind() {
						case reflect.String:
							s := m[ftag]						// copy it and then point to the copy
							f.Set( reflect.ValueOf(  &s ) )

						case reflect.Int:
							i := clike.Atoi( m[ftag] )				// convert to integer and then point at the value
							f.Set( reflect.ValueOf(  &i ) )

						case  reflect.Int64:
							i := clike.Atoi64( m[ftag] )
							f.Set( reflect.ValueOf(  &i ) )

						case  reflect.Int32:
							i := clike.Atoi32( m[ftag] )
							f.Set( reflect.ValueOf(  &i ) )

						case  reflect.Int16:
							i := clike.Atoi16( m[ftag] )
							f.Set( reflect.ValueOf(  &i ) )

						case  reflect.Int8:
							i := int8( clike.Atoi16( m[ftag] ) )
							f.Set( reflect.ValueOf(  &i ) )

						case reflect.Uint:
							ui := clike.Atou( m[ftag] )
							f.Set( reflect.ValueOf(  &ui ) )

						case reflect.Uint64:
							ui := clike.Atou64( m[ftag] )
							f.Set( reflect.ValueOf(  &ui ) )

						case reflect.Uint32:
							ui := clike.Atou32( m[ftag] )
							f.Set( reflect.ValueOf(  &ui ) )

						case reflect.Uint16:
							ui := clike.Atou16( m[ftag] )
							f.Set( reflect.ValueOf(  &ui ) )

						case reflect.Uint8:
							ui := uint8( clike.Atou16( m[ftag] ) )
							f.Set( reflect.ValueOf(  &ui ) )

						case reflect.Float64:
							fv := clike.Atof( m[ftag] )
							f.Set( reflect.ValueOf(  &fv ) )

						case  reflect.Float32:
							fv := float32( clike.Atof( m[ftag] ) )
							f.Set( reflect.ValueOf(  &fv ) )

						case reflect.Bool:
							b := m[ftag] == "true" || m[ftag] == "True" || m[ftag] == "TRUE"
							f.Set( reflect.ValueOf(  &b ) )
					}
					
				case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
					f.SetInt( clike.Atoi64( m[ftag] ) )

				case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
					f.SetUint( uint64( clike.Atoi64( m[ftag] ) ) )

				case reflect.Float64, reflect.Float32:
					f.SetFloat( clike.Atof( m[ftag] ) )

				case reflect.Bool:
					f.SetBool(  m[ftag] == "true" )
			}	
		}
	}

	return
}
