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


/*

	Mnemonic:	jsoncache
	Abstract:	Builds a cache of bytes returning an array of 'complete json' when available.
				Caller can use to read packets from the wire until a complete structure is
				discovered. If the first part of a second json struct is in the last packet
				it remains in the cache after the first complete struct is returned.
	Date:		16 December 2013
	Author:		E. Scott Daniels
*/

package jsontools

import (
	"fmt"
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


