// vi: sw=4 ts=4:

/*
	Run like this to test things and pretty print the json output
	go test|grep State|parse_json.rb

	Where parse_json.rb is some kind of json formatter (if parse_json.rb isn't available)
*/
package extcmd_test

import (
	"fmt"
	"os"
	"testing"
	//"time"

	"codecloud.web.att.com/gopkgs/extcmd"
)

func TestExec( t *testing.T ) {
	fmt.Fprintf( os.Stderr, "test started\n" );
	jdata, err := extcmd.Cmd2json( `test_script.ksh foo bar you "this is one" last`, "test_cmd" ) 

	if err != nil {
		fmt.Fprintf( os.Stderr, "command error: %s\n", err )
		t.Fail()
	}

	if jdata != nil {
		fmt.Fprintf( os.Stderr, "%s\n", jdata );
	} else {
		fmt.Fprintf( os.Stderr, "{jdata was nil}\n" )
	}
}

func TestLong( t *testing.T ) {
	fmt.Fprintf( os.Stderr, "long test started\n" );
	jdata, err := extcmd.Cmd2json( `test_script.ksh long foo bar you "this is one" last`, "test_cmd" ) 

	if err != nil {
		fmt.Fprintf( os.Stderr, "command error: %s\n", err )
		t.Fail()
	}

	if jdata != nil {
		fmt.Fprintf( os.Stderr, "%s\n", jdata );
	} else {
		fmt.Fprintf( os.Stderr, "{jdata was nil}\n" )
	}
}
