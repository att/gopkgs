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
	Run like this to test things and pretty print the json output
	go test|grep State|parse_json.rb

	Where parse_json.rb is some kind of json formatter (if parse_json.rb isn't available)
*/
package extcmd_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/att/gopkgs/extcmd"
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
