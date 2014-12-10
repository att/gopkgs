
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

