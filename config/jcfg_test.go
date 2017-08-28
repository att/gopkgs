
package config_test

import (
	"fmt"
	"os"
	"testing"

	"forge.research.att.com/switchboard/src/golib/cfg"
)

func TestJconfig( t *testing.T ) {
	jc,err := cfg.Mk_jconfig( "test.cfg", "sect1 sect3 sect2" )		// order of sections should be unimportant
	if err != nil {
		fmt.Fprintf( os.Stderr, "unable to allocate a jconfig struct: %s\n", err )
		t.Fail( )
		return
	}

	v1 := jc.Get_int( "sect3 sect2 sect1", "verbose", 0 ) 
	if v1 != 3 {
		fmt.Fprintf( os.Stderr, "attempt to get verbose from section 3 returned unexpected value: wanted 3 got: %d\n", v1 )
		t.Fail( )
	}	

	s1 := jc.Get_string( "sect3 sect2 sect1", "two-only", "" )
	if s1 != "only in two" {
		fmt.Fprintf( os.Stderr, "attempt to get string from section 2 returned unexpected value: wanted 'two-only',  got: %s\n", s1 )
		t.Fail( )
	}


	// dig a section from the main config, and then pull values from it and it's sections
	ss1, err := jc.Get_section( "sect2 sect3", "s3ss", "s3ss_deep1 s3ss_deep2" )
	if err != nil {
		fmt.Fprintf( os.Stderr, "attempt to get section s3ss failed: %s\n", err )
		t.Fail( )
	} else {
		fmt.Fprintf( os.Stderr, "ss3v1 = %d\n", ss1.Get_int( 	"s3ss_deep1 default", "s3ssv1", -1 ) )
		fmt.Fprintf( os.Stderr, "ss3v4 = %d\n", ss1.Get_int( 	"s3ss_deep4 default", "s3ssv4", -1 ) )
		fmt.Fprintf( os.Stderr, "ss3v4 = %d\n", ss1.Get_int( 	"s3ss_deep4 s3sss_deep3", "s3ssv4", -1 ) )
	}
}

