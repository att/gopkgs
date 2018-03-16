
package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/att/gopkgs/config"
)

/*
	These should all be tested:
	func ( cfg *Jconfig ) Extract_section( parent string, sname string, subsections string ) ( sect *Jconfig, err error ) {
	func ( cfg *Jconfig ) Extract_string( sects string, name string, def string ) ( string ) {
	func ( cfg *Jconfig ) Extract_int64( sects string, name string, def int64 ) ( int64 ) {
	func ( cfg *Jconfig ) Extract_int( sects string, name string, def int ) ( int ) {
	func ( cfg *Jconfig ) Extract_bool( sects string, name string, def bool ) ( bool ) {
	
	These need to be added:
	func ( cfg *Jconfig ) Extract_stringptr( sects string, name string, def interface{} ) ( *string ) {
	func ( cfg *Jconfig ) Extract_posint( sects string, name string, def int ) ( int ) {
	func ( cfg *Jconfig ) Extract_int32( sects string, name string, def int32 ) ( int32 ) {

*/

func TestJconfig( t *testing.T ) {
	fmt.Fprintf( os.Stderr, "\nTesting Json config file in jtest.cfg\n" )

	jc,err := config.Mk_jconfig( "jtest.cfg", "template bool-test laser-spec" )		// order of sections should be unimportant
	if err != nil {
		fmt.Fprintf( os.Stderr, "[FAIL] unable to allocate a jconfig struct: %s\n", err )
		t.Fail( )
		return
	}

	v1 := jc.Extract_posint( "template bool-test default laser-spec", "posint", 0 ) 					// should come from default section and have good value
	if v1 < 0 {
		fmt.Fprintf( os.Stderr, "[FAIL] attempt to get positive integer from default section returned unexpected value: wanted >0 got: %d\n", v1 )
		t.Fail( )
	} else {
		fmt.Fprintf( os.Stderr, "[OK]   extract positive int from correct section\n" )
	}

	v1 = jc.Extract_posint( "template bool-test default laser-spec", "negint", 1 ) 					// should come from default section and have default (0) value
	if v1 != 1 {
		fmt.Fprintf( os.Stderr, "[FAIL] attempt to get positive integer from default section returned unexpected value: wanted 1 got: %d\n", v1 )
		t.Fail( )
	} else {
		fmt.Fprintf( os.Stderr, "[OK]   extract positive int (default value)  from correct section\n" )
	}

	v1 = jc.Extract_int( "template bool-test laser-spec", "size", 0 ) 					// should come from template
	if v1 != 45 {
		fmt.Fprintf( os.Stderr, "[FAIL] attempt to get size from template section returned unexpected value: wanted 45 got: %d\n", v1 )
		t.Fail( )
	} else {
		fmt.Fprintf( os.Stderr, "[OK]   extract int from correct section\n" )
	}

	v1 = jc.Extract_int( "bool-test laser-spec template", "size", 0 ) 					// should not come from template
	if v1 == 45 {
		fmt.Fprintf( os.Stderr, "[FAIL] attempt to get size from section other than template returned unexpected value: wanted !45 got: %d\n", v1 )
		t.Fail( )
	} else {
		fmt.Fprintf( os.Stderr, "[OK]   extract int from correct section with alternate list\n" )
	}

	s1 := jc.Extract_string( "template bool-test laser-spec",  "laser-string", "" )
	if s1 != "string only in laser" {
		fmt.Fprintf( os.Stderr, "[FAIL] attempt to get string from laser-spec returned unexpected value: wanted 'string only in laser',  got: %s\n", s1 )
		t.Fail( )
	} else {
		fmt.Fprintf( os.Stderr, "[OK]   extract string from section\n" )
	}

	b := jc.Extract_bool( "bool-test", "istrue", false )
	if !b {
		fmt.Fprintf( os.Stderr, "[FAIL] attempt to extract boolian with true value failed\n" )
		t.Fail( )
	} else {
		fmt.Fprintf( os.Stderr, "[OK]   extract true bool from section\n" )
	}

	b = jc.Extract_bool( "bool-test", "isfalse", true )
	if b {
		fmt.Fprintf( os.Stderr, "[FAIL] attempt to extract boolian with false value failed\n" )
		t.Fail( )
	} else {
		fmt.Fprintf( os.Stderr, "[OK]   extract false booliean from section\n" )
	}

	b = jc.Extract_bool( "bool-test", "badbool", false )				// ensure returns the default
	if b {
		fmt.Fprintf( os.Stderr, "[FAIL] bool extract for missing did not return the default\n" )
		t.Fail( )
	} else {
		fmt.Fprintf( os.Stderr, "[OK]   extract false integer booliean from section\n" )
	}

	lv := jc.Extract_int64( "default template", "i64", 0 )		// value should be > 1234567890
	if lv < 1234567891 {
		fmt.Fprintf( os.Stderr, "[FAIL] int 64 value was too small = %d\n", lv )
		t.Fail( )
	} else {
		fmt.Fprintf( os.Stderr, "[OK]   int 64 value was ok %d\n", lv )
	}
		

	// dig a section from the main config, and then pull values from it and it's sections
	lss, err := jc.Extract_section( "template laser-spec", "laser-ss", "" )		// get 'default' section from laser-spec subsection laser-ss (not found in template searched first)
	if err != nil {
		fmt.Fprintf( os.Stderr, "[FAIL] attempt to get subsection section laser-ss from laser-spec failed: %s\n", err )
		t.Fail( )
	} else {
		i := lss.Extract_int( "default", "lssv1", -1 )
		if i != 1 {
			fmt.Fprintf( os.Stderr, "[FAIL] lssv1 = %d -- expected 1\n", i )
			t.Fail( )
		} 
		i = lss.Extract_int( "default", "lssv2", -2 )
		if i != 2 {
			fmt.Fprintf( os.Stderr, "[FAIL] lssv2 = %d -- expected 1\n", i )
			t.Fail( )
		} 

		s := lss.Extract_string( "default", "type", "not found" )
		if s != "laser subsection" {
			fmt.Fprintf( os.Stderr, "[FAIL] expected type string to be 'laser subsection' but got '%s'\n", s )
			t.Fail( )
		}
	}
}

