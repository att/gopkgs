// vi: ts=4 sw=4:

/*
	Mnemonic:	security_test
	Abstract: 	self test for security funcitons
	Date:		06 June 2014
	Author: 	E. Scott Daniels
*/

package security_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"forge.research.att.com/gopkgs/security"
)


//func Mk_cert( key_bits int, string *cert_name, dns_list []string, fname *string ) ( err error ) {
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
