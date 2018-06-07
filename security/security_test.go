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
	Mnemonic:	security_test
	Abstract: 	self test for security functions
	Date:		06 June 2014
	Author: 	E. Scott Daniels
*/

package security_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"github.com/att/gopkgs/security"
)

/*
	Test the creation of a cert and related key
*/
func TestSecurity_cert( t *testing.T ) {
	dns_list := make( []string, 2 )
	dns_list[0] = "localhost"

	this_host, err := os.Hostname( )
	if err == nil {
		tokens := strings.Split( this_host, "." )
		dns_list[1] = tokens[0]
	}

	
	cert_fname := "test_cert.pem"
	key_fname := "test_key.pem"
	cert_name := "foo_cert"

	err = security.Mk_cert( 1024, &cert_name, dns_list, &cert_fname, &key_fname )
	if err != nil {
		fmt.Fprintf( os.Stderr, "failed: %s", err )
		t.Fail();
	}
}
