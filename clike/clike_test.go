/*
	simple testing of the clike things
*/

package clike_test

import (
	"testing"
	"codecloud.web.att.com/gopkgs/clike"
)

func TestAtoll ( t *testing.T ) {           // must use bloody camel case to be recognised by go testing
	var (
		iv	interface{};					// use interface to ensure that the type coming back is 64 bit
		v 	int64;
	)

	iv = clike.Atoll( "123456" );
	switch iv.( type ) {
		case int64:		
			break;

		default:	
			t.Errorf( "atoll() did not return int64, no other atoll() tests executed" );
			return;
	}

	if iv.(int64) != 123456 {
		t.Errorf( "atoll( '123456' did not return 123456, got: %d", iv.(int64) );
	}

	v = clike.Atoll( "0x8000" );
	if v != 0x8000 {
		t.Errorf( "atoll( '0x8000' ) did not return 0x8000" );
	}

	v = clike.Atoll( "foo" );
	if v != 0 {
		t.Errorf( "atoll( \"foo\" )  did not return 0 as expected." );
	}

	v = clike.Atoll( "092" );
	if v != 0 {
		t.Errorf( "atoll( \"092\" )  did not return 0 as expected." );
	}

	v = clike.Atoll( "029" );
	if v != 2 {
		t.Errorf( "atoll( \"029\" )  did not return 2 as expected." );
	}
}

func TestAtof( t *testing.T ) {
	var (
		iv	interface{};
	)
	iv = clike.Atof( "123.456" );			// pick up result as an interface so we can test type as well as value
	switch iv.( type ) {
		case float64:		
			break;

		default:	
			t.Errorf( "atof() did not return float64, no other atof() tests executed" );
			return;
	}

	if iv.(float64) != 123.456 {
		t.Errorf( "atoll( '123.456' ) returned %.3f; did not return 123.456", iv.(float64)  );
	}
}

func TestAtoi32 ( t *testing.T ) {           // must use bloody camel case to be recognised by go testing
	var (
		iv	interface{};					// use interface to ensure var type returned is int32
		v 	int32;
	)

	iv = clike.Atoi32( "1256" );

	switch iv.( type ) {
		case int32:		
			break;

		default:	
			t.Errorf( "atoi32() did not return int32, no other ato32() tests executed" );
			return;
	}

	if iv.(int32) != 1256 {
		t.Fail();
	}

	v = clike.Atoi32( "0x8000" );
	if v != 0x8000 {
		t.Fail();
	}
}

func TestAtoi16 ( t *testing.T ) {           // must use bloody camel case to be recognised by go testing
	var (
		iv	interface{};					// use interface to ensure var type returned is int32
		v 	int16;
	)

	iv = clike.Atoi16( "256" );

	switch iv.( type ) {
		case int16:		
			break;

		default:	
			t.Errorf( "atoi16() did not return int16, no other atol16() tests executed" );
			return;
	}

	if iv.(int16) != 256 {
		t.Fail();
	}

	v = clike.Atoi16( "0x7def" );
	if v != 0x7def {
		t.Fail();
	}
}
