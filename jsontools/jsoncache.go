// vi: sw=4 ts=4:

/*

	Mnemonic:	jsoncache
	Abstract:	Builds a cache of bytes returning an array of 'complete json' when available.
	Date:		16 December 2013
	Author:		E. Scott Daniels
*/

package jsontools

import (
	//"bufio"
	//"bytes"
	//"encoding/json"
	//"flag"
	"fmt"
	//"io/ioutil"
	//"net/http"
	"os"
)

// --------------------- public ----------------------------------------------------------------------

type Jsoncache struct {
	buf	[]byte
	open	int
	nxt		int		// starting point for next get check
	len		int
	insrt	int		// insertion point
}

func Mk_jsoncache( ) ( jc *Jsoncache ) {
	jc = &Jsoncache { 
		open: 0,
		nxt: 0,
		insrt: 0,
	 }

	jc.buf = make( []byte, 8192 )
	
	return
}

/*
	Adds a buffer of bytes to the current cache.
*/
func ( jc *Jsoncache ) Add_bytes( newb []byte ) {
	var (
	)

	if jc == nil || len( jc.buf ) == 0 {
		fmt.Fprintf( os.Stderr, "ERR: jsoncache: object not allocated correctly\n" )
		return
	}

	for ; len( newb ) + jc.insrt > len( jc.buf ) ; {
		bigger := make( []byte, len( jc.buf ) * 2 )
		copy( bigger, jc.buf )
		jc.buf = bigger
	}

	for i := range newb {				// append new bytes
		jc.buf[jc.insrt] = newb[i]
		jc.insrt++
	}

	return
}

/*
	Returns a blob of complete json if it exists in the cache; nil otherwise
*/
func (jc *Jsoncache)  Get_blob( ) ( blob []byte ) {
	var (
		encountered_brace bool = false
	)

	blob = nil

	for ; jc.nxt < jc.insrt; jc.nxt++ {
		switch( jc.buf[jc.nxt] ) {
			case '{':	
					jc.open++
					encountered_brace = true

			case '}':	
					jc.open--
					encountered_brace = true
		}

		if encountered_brace && jc.open == 0 {
			blob = jc.buf[0:jc.nxt+1]
			endbuf := jc.buf[jc.nxt+1:]

			jc.buf = make( []byte, len( jc.buf ) )
			copy( jc.buf, endbuf )
			jc.insrt = jc.insrt - (jc.nxt+1)
			jc.nxt = 0
			return
		}
	}

	return;
}


