
package token

import (
	"strings";
)


/*
---------------------------------------------------------------------------------------
	Mnemonic:	tokenise_keep (split and keep separaters as individual tokens)
	Date:		22 Apr 2012
	Author: 	E. Scott Daniels
	Mods:		01 May 2012 : Added escape character support
				03 Feb 2015 : Removed unreachable code.
---------------------------------------------------------------------------------------
*/

/*
	Takes a string and slices it into tokens using the characters in sepchrs
	as the breaking points.  The separation characters are returned as 
	individual tokens in the list.  Separater characters escaped with a backslant
	are NOT treated like separators. 

	The return values are ntokens (int) and the list of tokens and separators.
*/
func Tokenise_keep(  buf string, sepchrs string ) (int, []string) {
	var (
		tokens []string;
		idx int;
	)

	idx = 0;
	tokens = make( []string, 2048 );	

	sepchrs += "\\";				// escape character is a separator

	subbuf := buf;
	tokens[idx] = "";				// prime the first one
	for {
		if i := strings.IndexAny( subbuf, sepchrs ); i >= 0 {
			if subbuf[i:i+1] == "\\" {				// next character is escaped
				tokens[idx] += subbuf[0:i];			// add everything before slant
				tokens[idx] += subbuf[i+1:i+2];		// add escaped char
				subbuf = subbuf[i+2:];				// advance past escaped char
			} else {
				if i > 0 {							// characters before the sep, capture them
					tokens[idx] += subbuf[0:i]; 	// add everything before separator
					idx++;
				} else {
					if tokens[idx] != "" {				// finish previous token
						idx++;
					}
				}
	
				tokens[idx] = subbuf[i:i+1];		// keep the sep as a separate token
				idx++;
				subbuf = subbuf[i+1:];				// advance past

				tokens[idx] = "";					// initialise next when advancing idx
			}
		} else {									// no more seps, just complete the list with remaining stuff and return

			tokens[idx] = subbuf[:len( subbuf )];	// capture last bit in tokens and go
			return idx+1, tokens[0:idx+1];
		}

		if len( subbuf ) < 1 {
			return idx, tokens[0:idx];
		}
	
	}	

	return
}
