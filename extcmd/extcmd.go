/* 
	Mnemonic:	extcmd.go
	Abstract: 	Functions that execute 'external' commands and return a response structure, 
				formatted json buffer, or write each record of the output onto a user
				channel (not implemented yet).
	Date: 		04 May 2014
	Author: 	E. Scott Daniels
*/

package extcmd

import (
	"bufio";
	"encoding/json"
	"os/exec"

	"forge.research.att.com/gopkgs/token"
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

	rdata = make( []string, 8192 )
	edata = make( []string, 8192 )

	ntokens, tokens := token.Tokenise_qpopulated( cbuf, " " )
	if ntokens < 1 {
		return
	}

	cmd := &exec.Cmd{}				// set the command up
	cmd.Path, _ = exec.LookPath( tokens[0] )
	cmd.Args = tokens
	stdout, err := cmd.StdoutPipe()			// create pipes for stderr/out 
	srdr := bufio.NewReader( stdout )		// standard out reader

	stderr, err := cmd.StderrPipe()
	erdr := bufio.NewReader( stderr )

	err = cmd.Start()
	if err != nil {
		return
	}

	i := 0
	for {
		buf, _, err := srdr.ReadLine( )
		if err != nil {
			break
		}
		if i < 4095 {					// must read it all before calling wait, but only snarf 4095
			if len( buf ) > 0 {
				nb := make( []byte, len( buf ) )
				copy( nb, buf )
				rdata[i] = string( nb )
				i++
			}	
		}
	}
	rdata = rdata[0:i]					// scale back the output to just what was used 

	i = 0
	for {
		buf, _, err := erdr.ReadLine( )
		if err != nil {
			break
		}
		if i < 4095  {					// must read it all before calling wait, but only snarf 4095
	
			if len( buf ) > 0 {
				nb := make( []byte, len( buf ) )
				copy( nb, buf )
				edata[i] = string( nb )
				i++
			}		
		}
	}
	edata = edata[0:i]					// scale back the output to just what was used 
	
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

