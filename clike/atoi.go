//	vi: sw=4 ts=4:

/*
        Mnemonic:       clike.go: atoi
        Absrtract:      a clike atoi, and varients, that don't error when it encounters
						a non-digit; returning 0 if there are no digits or if the number 
						is not valid, and stopping at the first non-valid digit.  This supports
						0x<hex> or 0<octal> values and should parse them stopping
						conversion at the first non-appropriate 'digit'.  This also allows
						a lead +/-.

						There is an extension on the C functions... if the value is 
						postfixed with M/K/G  or m/k/g the return value will be 
						'expanded' accordingly with the captialised values being
						powrs of 10 (e.g. MB) and the lower case indicating powers
						of 2 (e.g. MiB).

						Input can be either a string or a byte array
*/

package clike


/*
	Convert a string or an array of bytes into a 64 bit integer.
*/
func Atoi64( objx interface{} ) (int64) {

	v := Atoll( objx )
	return v
}

/*
	Convert a string or an array of bytes into a 32 bit integer.
*/
func Atoi32( objx interface{} ) (int32) {

	v := Atoll( objx )
	return int32( v )
}

/*
	Convert a string or an array of bytes into a 16 bit integer.
*/
func Atoi16( objx interface{} ) (int16) {

	v := Atoll( objx )
	return int16( v )
}

/*
	Convert a string or an array of bytes into a default sized integer.
*/
func Atoi( objx interface{} ) (int) {

	v := Atoll( objx )
	return int( v )
}

/*
	Convert a string or an array of bytes into an unsigned integer.
*/
func Atou( objx interface{} ) (uint) {

	v := Atoll( objx )
	return uint( v )
}
