
package token

import (
	"strings"
)


/*
---------------------------------------------------------------------------------------
	Mnemonic:	tokenise_qsep

	Returns:	number of tokens, tokens[]
	Date:		22 Apr 2012
	Author: 	E. Scott Daniels

	Mods:		14 Jan 2014 - correected bug that would allow quotes to remain on last token
					if there was not a separator between the quotes.
				30 Nov 2014 - Allows escaped quote.
---------------------------------------------------------------------------------------
*/


/*
	Takes a string and slices it into tokens using the characters in sepchrs
	as the breaking points, but allowing double quotes to provide protection
	against separatrion.  For example, if sepchrs is ",|", then the string
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
	This is the work horse for qsep and qpopulated. If keep_empty is true, then
	empty fields (concurrent separators) are kept as empty strings.
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
			for ttidx := 0; q < len( subbuf )  && subbuf[q] != '"'; ttidx++ { 		// find end trashing escape characters
				if subbuf[q] == '\\' {
					q++
				}
				ttok[ttidx] = subbuf[q]
				q++
			}
			if q > 0 {													// could have been ,foo""
				tokens[idx] += string( ttok[0:q] )
			}
			subbuf = subbuf[q+1:]
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

	return 0, nil
}

