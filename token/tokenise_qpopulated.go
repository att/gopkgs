
package token

import (
)


/*
---------------------------------------------------------------------------------------------
	Mnemonic:	tokenise_qpopulated

	Returns:	number of tokens, tokens[]
	Date:		22 Apr 2012
	Author: 	E. Scott Daniels

	Mods:		13 Jan 2014 - corrected bug causing quotes to be left on last token if 
						there isn't a sep inside of the tokens.
				02 May 2014 - corrected documentation, removed unneeded commented out code.
				30 Nov 2014 - Now allows escaped quotes in quoted portion.
				01 Dec 2014 - Now uses the work horse tokenise_all() function, so the majority
						of this disapears.
---------------------------------------------------------------------------------------------
*/

/*
	Takes a string and slices it into tokens using the characters in sepchrs
	as the breaking points, but allowing double quotes to provide protection
	against separatrion.  For example, if sepchrs is ",|", then the string
		foo,bar,"hello,world",,"you|me"

	would break into 4 tokens:
		foo
		bar
		hello,world
		you|me

	Similar to tokenise_qsep, but this method removes empty tokens from the 
	final result.


	The return value is the number of tokens and the list of tokens. 
*/
func Tokenise_qpopulated(  buf string, sepchrs string ) (int, []string) {
	return tokenise_all( buf, sepchrs, false )
}

