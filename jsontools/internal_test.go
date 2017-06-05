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
	Mnemonic:	internal_test.go
	Abstract:	Tests which drive internal funcitons and thus the package must
				be jsontools and not jsontools_test
	Author:		E. Scott Daniels
	Date:		5 June 2017
*/

package jsontools

import (
	"fmt"
	"testing"
	"os"

//	"github.com/att/gopkgs/jsontools"
)


func TestEscString( t *testing.T ) {
	s := `now is the "time" for all good "boys" to get real`
	es := `now is the \"time\" for all good \"boys\" to get real`		// expected

	fmt.Fprintf( os.Stderr, "------ testing escape addition starts ----- \n" )
	gs := add_esc( s )
	if gs != es {
		fmt.Fprintf( os.Stderr, "[FAIL] escaped string not what we expected\n   got (%s)\nexpect (%s)\n", gs, es )
		t.Fail()
	} else { 
		fmt.Fprintf( os.Stderr, "[OK]   escaped string is what we expected\n   got (%s)\nexpect (%s)\n", gs, es )
	}

	fmt.Fprintf( os.Stderr, "------ testing escape addition complete -----\n" )
}
