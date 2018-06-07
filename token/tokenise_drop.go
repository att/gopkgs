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
	"strings";
)


/*
---------------------------------------------------------------------------------------
	Mnemonic:	tokenise_drop
	Returns:	ntokens, tokens[]
	Date:		22 Apr 2012
	Author: 	E. Scott Daniels
	Mod:		03 Feb 2015 - Removed unreachable code
				07 Jun 2018 : Remove unreachable code (keep vet happy)
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
}
