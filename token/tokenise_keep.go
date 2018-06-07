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
	Mnemonic:	tokenise_keep (split and keep separators as individual tokens)
	Date:		22 Apr 2012
	Author: 	E. Scott Daniels
	Mods:		01 May 2012 : Added escape character support
				03 Feb 2015 : Removed unreachable code.
				07 Jun 2018 : Remove unreachable code (keep vet happy)
---------------------------------------------------------------------------------------
*/

/*
	Tokensise_keep takes a string and slices it into tokens using the characters in sepchrs
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
}
