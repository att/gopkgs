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

	Mnemonic:	defrock
	Abstract:	Functions to support 'refrocking' a json string where some contents
				of the json are strings of json (embedded). When the desired result
				is to have the embedded strings actually be a part of the hierarchy
				the Refrock() function can be used to convert the string into a 
				string of json. If a usable map of json is desired, then the 
				Defrock_2_jif() function can be used to completely defrock the 
				string leaving it in a map[string]interface{} which can be used by
				other tools.  

	Date:		2 June 2017
	Author:		E. Scott Daniels

	Mods:
*/

package jsontools

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// --------------------- public ----------------------------------------------------------------------
/*
	Take json in the form of string, *string or []byte and unencapsulate any embedded strings which 
	are json. The result is a string with valid json but with any embedded json strings in the original
	made into a part of the resulting 'hierarchy'. This is a two step process which simply uses the 
	Defrock_2_jif() and Frock_jmap() functions below.
*/
func Refrock( jblob interface{} ) ( string, error ) {
	jif, err := Defrock_2_jif( jblob )
	if err != nil {
		return "", err
	}

	return Frock_jmap( jif ), nil
}

/*
	Take ehter json input (as string, *string or []byte), or a map[]interface{} and completey defrock 
	it into a json inteface (jif).  This assumes that some strings may be json which should be recursively 
	defrocked. Yes, what system might actually embed json as a string in json.... Openstack; grrrr.
*/
func Defrock_2_jif( iblob interface{} ) ( jif interface{}, err error ) {
	var (
		jblob []byte
		ojif interface{}		// orignal unpacked -- may contain json strings
	)

	need_unpack := true

	switch thing := iblob.(type) {
		case string:
			jblob = []byte( thing )

		case *string:
			jblob = []byte( *thing )

		case []byte:
			jblob = thing

		default:
			need_unpack = false		// we assume it's a map[] of interface and will fail later if not
			ojif = thing
	}

	if need_unpack {
		err = json.Unmarshal( jblob, &ojif )			// unpack the json into a map[]interface{}
		if err != nil {
			return nil, fmt.Errorf( "unable to unpack json into jif: %s\n", err )
		}
	}

	if m, ok := ojif.( map[string]interface{} ); ok {			// run the result and look for strings that might be embedded json; unpack those too
		for k, v := range m {
			switch thing := v.(type) {
				case string:
					sjif, err := Defrock_2_jif( rm_esc( thing ) )		// attempt to defrock json; if successful then we save the interface
					if err == nil {
						m[k] = sjif
					}

				case map[string]interface{}:
					sjif, err := Defrock_2_jif( thing )		// attempt to defrock embedded json strings; if successful then we save the interface
					if err == nil {
						m[k] = sjif
					}

				case []interface{}:
					for e, ele := range thing {
						switch ele.(type) {
							case string:
								sjif, err := Defrock_2_jif( rm_esc( ele.(string) ) )	// if it's json, defrock and save the interface
								if err == nil {
									thing[e] = sjif
								}	

							case map[string]interface{}:			// recurse to possibly defrock what lies below
								sjif, err := Defrock_2_jif( ele )
								if err == nil {
									thing[e] = sjif
								}

							default: 
								// do nothing; just leave what we found
						}
					}

				default: 
					// do nothing; leave what we found
			}
		}	
	} else {
		return nil, fmt.Errorf( "pointer to jif map wasn't to a map[string]interface{}: %s", reflect.TypeOf( ojif ) )
	}

	return ojif, nil
}

/*
	Given an interface, put it into a json frock (format it with quotes and squiggles).
*/
func frock_if( jif interface{} ) ( string ) {

	switch thing := jif.(type) {
		case string:
			return fmt.Sprintf( "%q", add_esc( thing ) )

		case int:
			return fmt.Sprintf( "%d", thing )

		case bool:
			return fmt.Sprintf( "%v", thing )

		case float64:
			return fmt.Sprintf( "%0.3f", thing )

		case map[string]interface{}:
			return Frock_jmap( thing )

		case []interface{}:
			asep := " "
			jstr := "["
				for _, iv := range thing {
					jstr += asep + frock_if( iv )
					asep = ", "	
				}
			jstr += " ]"
			return jstr

		case nil:
			return "null"

		default: 
			//fmt.Fprintf( os.Stderr, "frock_if: unknown type: %s\n", reflect.TypeOf( thing ) )
	}	

	return ""
}
/*
	Accepts a jmap (keyed interface values) and generates the corresponding formatted (frocked) json
	in a string.
*/
func Frock_jmap( jif interface{} ) ( json string ) {
	jmap := make( map[string]interface{}, 1024 )
	
	jmap, ok := jif.( map[string]interface{} )			// right now we only support a jmap
	if ! ok {
		return ""
	}

	jstr := "{"
	sep := " "

	for k, v := range jmap {
		jstr += fmt.Sprintf( "%s%q: ", sep, k )
		jstr += frock_if( v )
		sep = ", "
	}

	jstr += "}"
	return jstr
}


// ----------------- private ---------------
/*
	Escape quotes in the string.
*/
func add_esc( unq string ) ( quoted string ) {
	
	b := make( []byte, len( unq ) * 2 )
	bi := 0
	for _, c := range( unq ) {
		if c == '"' {
			b[bi] = '\\'
			bi++
		}
		b[bi] = byte( c )
		bi++
	}

	return string( b[0:bi] )
}

/*
	Remove a layer of escapes from a string.
*/
func rm_esc( esc string ) ( string ) {
	b := make( []byte, len( esc ) )
	s := []byte( esc )

	bi := 0
	si := 0
	for ; si < len( esc )-1; {
		if s[si] == '\\' {
			si++
		}
		b[bi] = s[si]
		bi++
		si++
	}

	if si < len( esc ) {
		b[bi] = s[si]
		bi++
	}

	return string( b[0:bi] )
}
