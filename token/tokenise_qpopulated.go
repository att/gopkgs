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
						of this disappears.
---------------------------------------------------------------------------------------------
*/

/*
	Takes a string and slices it into tokens using the characters in sepchrs
	as the breaking points, but allowing double quotes to provide protection
	against separation.  For example, if sepchrs is ",|", then the string
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

