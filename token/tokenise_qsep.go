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
	"strings"
)


/*
---------------------------------------------------------------------------------------------
	Mnemonic:	tokenise_qsep

	Returns:	number of tokens, tokens[]
	Date:		22 Apr 2012
	Author: 	E. Scott Daniels

	Mods:		14 Jan 2014 - corrected bug that would allow quotes to remain on last token
					if there was not a separator between the quotes.
				30 Nov 2014 - Allows escaped quote.
				09 Apr 2015 - Corrected problem where index was not being checked and range
					was being busted causing a panic. Removed the 2k limit.
				11 Aug 2018 - Correct bug if string ends in single quote.
				27 Aug 2018 - Correct bug when string has escaped characters.
---------------------------------------------------------------------------------------------
*/


/*
	Takes a string and slices it into tokens using the characters in sepchrs
	as the breaking points, but allowing double quotes to provide protection
	against separation.  For example, if sepchrs is ",|", then the string
		foo,bar,"hello,world","you|me"

	would break into 4 tokens:
		foo
		bar
		hello,world
		you|me

	If there are empty fields, they are returned as empty tokens.

	The return values are the number of tokens and the list of tokens.
*/
func Tokenise_qsep(  buf string, sepchrs string ) (int, []string) {
	return tokenise_all( buf, sepchrs, true )
}

/*
	Tokenises a string, but returns only an array of unique tokens.
	Empty tokens are discarded.
*/
func Tokenise_qsepu( buf string, sepchrs string ) ( int, []string ) {

	seen := make( map [string]bool, 1024 )
	n, toks := tokenise_all( buf, sepchrs, false )
	rtoks := make( []string, n )
	j := 0
	for _, v := range toks {
		if ! seen[v] 	{
			seen[v] = true
			rtoks[j] = v
			j++
		}
	}

	return j,  rtoks[0:j]
}

/*
	This is the work horse for qsep and qpopulated. If keep_empty is true, then
	empty fields (adjacent separators) are kept as empty strings.
*/
func tokenise_all( buf string, sepchrs string, keep_empty bool ) (int,  []string) {
	var tokens []string
	var	idx int
	var	i int
	var	q int

	idx = 0
	tokens = make( []string, 2048 )

	subbuf := buf
	for {
		i = strings.IndexAny( subbuf, sepchrs ) 		// index of the next sep character
		q = strings.IndexAny( subbuf, "\"" ) 			// index of the next quote

		if idx >= len( tokens ) {						// more than we had room for; alloc new
			tnew := make( []string, len( tokens ) * 2 )
			copy( tnew[:], tokens )
			tokens = tnew
		}

		if q < 0 || (q >= i && i >= 0) {				// sep before quote, or no quotes
			if i > 0 {
				tokens[idx] = subbuf[0:i]; 				// snarf up to the sep
				idx++
			} else {
				if i == 0 {
					if keep_empty {
						tokens[idx] = ""					// empty token when sep character @ 0
						idx++
					}
				} else {								// no more sep chrs and no quotes; capture last and bail
					if q < 1 {
						tokens[idx] = subbuf
						return idx+1, tokens[0:idx+1]
					}
				}
			}

			subbuf = subbuf[i+1:];			// skip what was added as token, and the sep
		} else {

			if q > 0 {						// stuf before quote -- capture it now
				tokens[idx] = subbuf[0:q]
			} else {
				tokens[idx] = ""
			}

			subbuf = subbuf[q+1:];										// skip what we just snarfed, and opening quote
			ttok := make( []byte, len( subbuf ) )						// work space to strip escape characters in
			q = 0
			ttidx := 0
			for ttidx = 0; q < len( subbuf )  && subbuf[q] != '"'; ttidx++ { 		// find end trashing escape characters
				if subbuf[q] == '\\' {
					q++
				}
				ttok[ttidx] = subbuf[q]
				q++
			}
			if q > 0 {													// could have been ,foo""
				tokens[idx] += string( ttok[0:ttidx] )
			}
			if q < len( subbuf ) - 1 {									// prevent bounds exception if single trailing quote
				subbuf = subbuf[q+1:]
			} else {
				subbuf = ""
			}
			i = strings.IndexAny( subbuf, sepchrs )						// next sep, if there, past quoted stuff
			q = 0
			if q < i {
				tokens[idx] += subbuf[0:i-q]							// snarf anything after quote and before sep
				if len( subbuf ) > 0 {
					subbuf = subbuf[(i-q)+1:]							// finally skip past sep
				}
			} else {
				if len( subbuf ) > 0 {
					subbuf = subbuf[q+1:]
				}
			}

			idx++
		}

		if  len( subbuf ) < 1 {
			return idx, tokens[0:idx]
		}

	}	

	return 0, nil				// unreacable and vet will complain, but without this older compilers refuse to compile!
}

