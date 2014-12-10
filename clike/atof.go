//	vi: sw=4 ts=4:

/*
        Mnemonic:       clike.go: atof
        Absrtract:      a clike atof that doesn't error when it encounters
						a non-digit; returning 0 if there are no digits.  The input
						(string or buffer) is expected to be base 10 with optional leading 
						zeros, an optional trailing decimal and an optional fraction 
						following the decimal.  This also allows a lead +/-.

						There is an extension on the C functions... if the value is 
						postfixed with M/K/G  or m/k/g the return value will be 
						'expanded' accordingly with the captialised values being
						powrs of 10 (e.g. MB) and the lower case indicating powers
						of 2 (e.g. MiB).

						Input can be either a string or a byte array
		Author:			E. Scott Daniels
		Date:			October 12 2013
*/

package clike

import (
        "strconv" ;
)


/*
	Atof accepts a string or an array of bytes ([]byte) and converts the characters
	into a float64 value.  The input may be postfixed with either any of the following
	characters which cause the value to be expanded:
		m,k,g - powers of two expansion. e.g. 10m expands to 10 * 1024 * 1024.
		M,K,G - powers of ten expansion. e.g. 10M expands to 10000000

	Unlike the Go string functions, this stops parsing at the first non-digit in the 
	same manner that the Clib functions do. 
*/
func Atof( objx interface{} ) (v float64) {
        var (
			i int;
			buf	[]byte;
		)

		v = 0;							// ensure all early returns have a value of 0

		if objx == nil {
			return
		}

		switch objx.( type ) {
			case []byte:	
						buf = objx.([]byte);			// type assertion seems backwards doesn't it?
			case string:	
						buf = []byte( objx.(string) );
			default:
						return;					// who knows, but it doesn't convert
		}

		if len( buf ) < 1 {
			return;
		}

        i = 0;
		if buf[i] == '-' || buf[i] == '+' {
			i++
		}
       	for ; i < len(buf)  &&  ((buf[i] >= '0'  &&  buf[i] <= '9') || buf[i] == '.'); i++ {}	// find last valid character for conversion

        if i > 0 {
                v, _ = strconv.ParseFloat( string( buf[0:i] ), 64 );
        }

		if i < len( buf ) {
			switch string( buf[i:] ) {
				case "M", "MB":
						v *= 1000000;

				case "G", "GB":
						v *= 1000000000;

				case "K", "KB":
						v *= 1000;

				case "m", "MiB":
						v *= 1048576;

				case "g", "GiB":
						v *= 1073741824;

				case "k", "KiB":
						v *= 1024;

				default: break;	
			}
		}

        return;
}

