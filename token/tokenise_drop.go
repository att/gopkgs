
package token

import (
	"strings";
)


/*
---------------------------------------------------------------------------------------
	Mnemonic:	tokenise_drop
	Returns:	ntokens, tokens[]
	Date:		22 Apr 2012
	Author: 	E. Scott Daniels
---------------------------------------------------------------------------------------
*/

/*
	Takes a string and slices it into tokens using the characters in sepchrs
	as the breaking points.  The separation characters are discarded.
*/
func Tokenise_drop(  buf string, sepchrs string ) ( ntokens int, tokens []string) {
	var	idx int;

	idx = 0;
	tokens = make( []string, 2048 );	

	sepchrs += "\\"; 
	subbuf := buf;
	tokens[idx] = "";
	for {
		if i := strings.IndexAny( subbuf, sepchrs ); i >= 0 {
			if subbuf[i:i+1] == "\\" {				// next character is escaped
				tokens[idx] += subbuf[0:i];			// add everything before slant
				tokens[idx] += subbuf[i+1:i+2];		// add escaped char
				subbuf = subbuf[i+2:];				// advance past escaped char
			} else {
				if i > 0 {							// characters before the sep, capture them
					tokens[idx] += subbuf[0:i]; 	// add everything before separator
				} 

				idx++;
				tokens[idx] = "";
				subbuf = subbuf[i+1:];
			}
		} else {
			tokens[idx] += subbuf[:len( subbuf )];
			tokens = tokens[0:idx+1]
			ntokens = idx + 1
			return 
		}

		if len( subbuf ) < 1 {
			ntokens = idx 
			tokens = tokens[0:idx+1];
			return 
		}
	}	

	tokens = nil
	ntokens = 0
	return
}
