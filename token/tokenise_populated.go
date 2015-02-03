
package token

import (
	"strings";
)


/*
---------------------------------------------------------------------------------------
	Mnemonic:	tokenise_fields
	Returns:	ntokens, tokens[]
	Date:		21 November 2013
	Author: 	E. Scott Daniels
	Mod:		03 Feb 2015 - Removed unreachable code.
---------------------------------------------------------------------------------------
*/

/*
	Takes a string and slices it into tokens using the characters in sepchrs
	as the breaking points keeping only populated fields.  The separation characters 
	are discarded.  Null (empty) tokens are dropped allowing a space separated record to 
	treat multiple spaces as a single separator.

	The return values are the number of tokens and the list of token strings.
*/
func Tokenise_populated(  buf string, sepchrs string ) (int, []string) {
	var tokens []string;
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
					idx++;							// capture only non-empty tokens
					tokens[idx] = "";
				} 

				subbuf = subbuf[i+1:];
			}
		} else {
			tokens[idx] += subbuf[:len( subbuf )];
			return idx+1, tokens[0:idx+1];
		}

		if len( subbuf ) < 1 {
			return idx, tokens[0:idx+1];
		}
	}	
}
