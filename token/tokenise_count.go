//vi: sw=4 ts=4:
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
---------------------------------------------------------------------------------------
	Mnemonic:	tokenise_count

	Returns:	map[string]int
	Date:		22 Apr 2012
	Author: 	E. Scott Daniels
---------------------------------------------------------------------------------------
*/


/*
	Tokenises the string using tokenise_qsep and then builds a map which
	counts each token (map[string]int). Can be used by the caller as an
	easy de-dup process.
*/
func Tokenise_count(  buf string, sepchrs string ) ( cmap map[string]int ) {
	var (
		tokens 	[]string
		ntokens int
	)
	
	cmap = make( map[string]int )

	ntokens, tokens = Tokenise_qsep( buf, sepchrs )
	for i := 0; i < ntokens; i++ {
		cmap[tokens[i]]++
	}

	return
}

