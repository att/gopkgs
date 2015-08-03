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

/*
	Mnemonic:	extcmd.go
	Abstract: 	Functions that execute 'external' commands and return a response structure,
				formatted json buffer, or write each record of the output onto a user
				channel (not implemented yet).
	Date: 		04 May 2014
	Author: 	E. Scott Daniels
	Mods:		11 Jan 2015 - Fix bug that was improperly truncating the stdout stderr
							array (failing to limit it if the command could not be run).
*/

package extcmd

import (
	"bufio";
	"encoding/json"
	"os/exec"

	"github.com/att/gopkgs/token"
)


type Response struct {
	Ctype	string
	Rtype	string
	Rstate	int
	Rdata	[]string
	Edata	[]string
}

/*
	Accept and build a command, run it writing the output into a json buffer that is
	returned.  Currently, output response is truncated at 8192 records.
	If a caller needs more, then it should redirect the output and parse the
	file rather than reading it all into core.
*/
	
func Cmd2strings( cbuf string ) ( rdata []string, edata []string, err error ) {


	ntokens, tokens := token.Tokenise_qpopulated( cbuf, " " )
	if ntokens < 1 {
		return
	}

	cmd := &exec.Cmd{}						// set the command up
	cmd.Path, _ = exec.LookPath( tokens[0] )
	cmd.Args = tokens
	stdout, err := cmd.StdoutPipe()			// create pipes for stderr/out
	srdr := bufio.NewReader( stdout )		// standard out reader

	stderr, err := cmd.StderrPipe()
	erdr := bufio.NewReader( stderr )

	err = cmd.Start()
	if err != nil {
		rdata = make( []string, 1 )
		edata = make( []string, 1 )
		return
	}

	rdata = make( []string, 8192 )
	edata = make( []string, 8192 )

	i := 0
	for {
		buf, _, err := srdr.ReadLine( )
		if err != nil {
			break
		}
		if i < len( rdata ) {					// must read it all before calling wait, but don't overrun our buffer
			if len( buf ) > 0 {
				nb := make( []byte, len( buf ) )
				copy( nb, buf )
				rdata[i] = string( nb )
				i++
			}	
		}
	}
	if i > 0 {
		rdata = rdata[0:i]					// scale back the output to just what was used
	} else {
		rdata = rdata[0:1]
	}

	i = 0
	for {
		buf, _, err := erdr.ReadLine( )
		if err != nil {
			break
		}
		if i < len( edata )  {					// must read it all before calling wait, but don't overrun our buffer
	
			if len( buf ) > 0 {
				nb := make( []byte, len( buf ) )
				copy( nb, buf )
				edata[i] = string( nb )
				i++
			}		
		}
	}
	if i > 0 {
		edata = edata[0:i]					// scale back the output to just what was used
	} else {
		edata = edata[0:1]
	}
	
	err = cmd.Wait()

	return
}

/*
	Execute the command in cbuf and return a Response structure with the results.
*/
func Cmd2resp( cbuf string, rtype string )  ( rblock *Response, err error ) {
	rblock = &Response{ Ctype: "response", Rtype: rtype }

	rblock.Rdata, rblock.Edata, err = Cmd2strings( cbuf )

	if err == nil {
		rblock.Rstate = 0
	} else {
		rblock.Rstate = 1
	}

	return
}

/*
	Executes a command and returns the response structure in json format
*/
func Cmd2json( cbuf string, rtype string )  ( jdata []byte, err error ) {
	resp, err := Cmd2resp( cbuf, rtype )
	if err != nil {
		return
	}

	jdata, err = json.Marshal( *resp )	

	return
}

