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
        Mnemonic:       atoll
        Absrtract:      a clike atoll that doesn't error when it encounters
						a non-digit; returning 0 if there are no digits.  This supports
						0x<hex> or 0<octal> values and should parse them stopping
						conversion at the first non-appropriate 'digit'.  This also allows
						a lead +/-.

						There is an extension on the C functions... if the value is
						postfixed with M/K/G  or m/k/g the return value will be
						'expanded' accordingly with the capitalised values being
						powrs of 10 (e.g. MB) and the lower case indicating powers
						of 2 (e.g. MiB).  This might cause unwanted side effects should
						the goal be to take a string 200grains and capture just the
						value; these functions will interpret the leading g incorrectly.

						Input can be either a string, pointer to string, or a byte array.

						If the string/array passed in does not begin with a valid digit
						or sign, or is a pointer that is nil,  a value of 0 is returned
						rather than an error.
*/

package clike

import (
        "strconv"
)


/*
	Convert a string or an array of bytes into a 64 bit integer.
*/
func Atoll( objx interface{} ) (v int64) {
        var (
			i int
			buf	[]byte
		)

		v = 0							// ensure all early returns have a value of 0

		if objx == nil {
			return
		}

		switch objx.( type ) {					// place into a container we can deal with
			case []byte:	
						buf = objx.([]byte)			
			case string:	
						buf = []byte( objx.(string) )
			case *string:
						bp := objx.( *string )
						if bp == nil {
							return 0
						}
						buf = []byte( *bp );
			default:
						return					// who knows, but it doesn't convert
		}

		if len( buf ) < 1 {
			return
		}

        i = 0
		if buf[i] == '-' || buf[i] == '+' {
			i++
		}
        if buf[i] == '0' {
			if  len( buf ) > 1 && buf[i+1] == 'x' {
				i += 2
        		for ; i < len(buf)  &&  ((buf[i] >= '0'  &&  buf[i] <= '9') || (buf[i] >= 'a' && buf[i] <= 'f') ); i++ {}	// find last digit
			} else {
				i++
        		for ; i < len(buf)  &&  (buf[i] >= '0'  &&  buf[i] <= '7'); i++ {}							// find last octal digit
			}
        }  else {
        	for ; i < len(buf)  &&  (buf[i] >= '0'  &&  buf[i] <= '9'); i++ {}	// find last digit
		}

        if i > 0 {
                v, _ = strconv.ParseInt( string( buf[0:i] ), 0, 64 )
        }

		if i < len( buf ) {
			switch string( buf[i:] ) {
				case "M", "MB":
						v *= 1000000

				case "G", "GB":
						v *= 1000000000

				case "K", "KB":
						v *= 1000

				case "m", "MiB":
						v *= 1048576

				case "g", "GiB":
						v *= 1073741824

				case "k", "KiB":
						v *= 1024

				default: break;	
			}
		}

        return
}

/*
	Convert to unsigned 64bit.
*/
func Atoull( objx interface{} ) (v uint64) {
        var (
			i int
			buf	[]byte
		)

		v = 0							// ensure all early returns have a value of 0

		if objx == nil {
			return
		}

		switch objx.( type ) {					// place into a container we can deal with
			case []byte:	
						buf = objx.([]byte)			
			case string:	
						buf = []byte( objx.(string) )
			case *string:
						bp := objx.( *string )
						if bp == nil {
							return 0
						}
						buf = []byte( *bp );
			default:
						return					// who knows, but it doesn't convert
		}

		if len( buf ) < 1 {
			return
		}

        i = 0
		if buf[i] == '-' || buf[i] == '+' {
			i++
		}
        if buf[i] == '0' {
			if  len( buf ) > 1 && buf[i+1] == 'x' {
				i += 2
        		for ; i < len(buf)  &&  ((buf[i] >= '0'  &&  buf[i] <= '9') || (buf[i] >= 'a' && buf[i] <= 'f') ); i++ {}	// find last digit
			} else {
				i++
        		for ; i < len(buf)  &&  (buf[i] >= '0'  &&  buf[i] <= '7'); i++ {}							// find last octal digit
			}
        }  else {
        	for ; i < len(buf)  &&  (buf[i] >= '0'  &&  buf[i] <= '9'); i++ {}	// find last digit
		}

        if i > 0 {
                v, _ = strconv.ParseUint( string( buf[0:i] ), 0, 64 )
        }

		if i < len( buf ) {
			switch string( buf[i:] ) {
				case "M", "MB":
						v *= 1000000

				case "G", "GB":
						v *= 1000000000

				case "K", "KB":
						v *= 1000

				case "m", "MiB":
						v *= 1048576

				case "g", "GiB":
						v *= 1073741824

				case "k", "KiB":
						v *= 1024

				default: break;	
			}
		}

        return
}
