// vi: sw=4 ts=4:
/*
 ---------------------------------------------------------------------------
   Copyright (c) 2013-2015 AT&T Intellectual Property

   Licensed under the Apache License, Version 2.0 (the "License")
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


package config_test

import (
	"testing"
	"fmt"
	"os"
	"reflect"

	"github.com/att/gopkgs/config"
)

// check type and write error msg if not desired. Return true if ok, false otherwise.
func is_type( name string, thing interface{}, desired string ) ( bool ) {
	if reflect.TypeOf( thing ).String() != desired {
		fmt.Fprintf( os.Stderr, "[FAIL] %s does NOT have the expected type: wanted=%s got=%s\n", name, desired, reflect.TypeOf( thing ))
		return false
	}

	return true
}

func TestConfig( t *testing.T ) {
	sects, err := config.Parse( nil, "test.cfg", false )

	if err != nil {
		fmt.Fprintf( os.Stderr, "parsing config file failed: %s\n", err )
		t.Fail()
		return
	}

	for sname, smap := range sects {
		fmt.Fprintf( os.Stderr, "section: (%s) has %d items\n", sname, len( smap ) )
		for key, value := range smap {
			switch value.( type ) {
				case string:
					fmt.Fprintf( os.Stderr, "\t%s = %s\n", key, ( value.( string ) ) )

				case *string:
					fmt.Fprintf( os.Stderr, "\t%s = (%s)\n", key, *( value.( *string ) ) )

				case float64:
					fmt.Fprintf( os.Stderr, "\t%s = %v\n", key, value )

				default:
			}
		}
	}

	smap := sects["default"]
	fmt.Fprintf( os.Stderr, "qfoo=== (%s)\n", *(smap["qfoo"].(*string)) )
	fmt.Fprintf( os.Stderr, "ffoo=== %8.2f\n", smap["ffoo"].(float64) )
	fmt.Fprintf( os.Stderr, "jfoo=== (%s)\n", *(smap["jfoo"].(*string)) )
}

/*
	test reading only as strings
*/
func TestStrings( t *testing.T ) {
	var (
		my_map map[string]map[string]*string
		dup	string
	)

	my_map = make( map[string]map[string]*string )
	my_map["default"] = make( map[string]*string )
	dup = "should be overlaid by config file info";				// should be overridden
	my_map["default"]["ffoo"] = &dup
	dup = "initial value, should exist after read"
	my_map["default"]["init-val"] = &dup

	sects, err := config.Parse2strs( my_map, "test.cfg" )

	if err != nil {
		fmt.Fprintf( os.Stderr, "parsing config file failed: %s\n", err )
		t.Fail()
		return
	}

	for sname, smap := range sects {
		fmt.Fprintf( os.Stderr, "section: (%s) has %d items\n", sname, len( smap ) )
		for key, value := range smap {
			fmt.Fprintf( os.Stderr, "\t%s = (%s)\n", key, *value )
		}
	}

	smap := sects["default"]									// can be referenced two different ways
	fmt.Fprintf( os.Stderr, "qfoo=== (%s)\n", *smap["qfoo"] )
	fmt.Fprintf( os.Stderr, "ffoo=== (%s)\n", *smap["ffoo"] )
	fmt.Fprintf( os.Stderr, "ffoo=== (%s)\n", *sects["default"]["ffoo"] )
}

func TestConfigStruct( t *testing.T ) {
	fmt.Fprintf( os.Stderr, "\nTesting config struct functions....\n" )

	cfg, err := config.Mk_config( "test.cfg" )
	if 	err != nil {
		t.Fail()
		fmt.Fprintf( os.Stderr, "unable to load config into a config struct: %s\n", err )
		return
	}

	f1 := cfg.Extract_float( "template laser-spec", "default_size", 99.0 )
	if ! is_type( "f1", f1, "float64" )  {
		t.Fail()
	}
	f2 := cfg.Extract_float( "laser-spec template", "default_size", 99.0 )
	if ! is_type( "f2", f2, "float64" )  {
		t.Fail()
	}
	if f1 == f2 {
		fmt.Fprintf( os.Stderr, "[FAIL] expected different values from same key in different sections, got same falue: %.2f %.2f\n", f1, f2 )
		t.Fail()
	} else {
		fmt.Fprintf( os.Stderr, "[PASS] expected different values from same key in different sections, got different: %.2f %.2f\n", f1, f2 )
	}

	f3s := cfg.Extract_str( "laser-spec template", "default_size", "99.0" )
	if ! is_type( "f3s", f3s, "string" )  {
		t.Fail()
	}

	cfg, err = config.Mk_config( "/dev/null" )
	if err == nil {
		fmt.Fprintf( os.Stderr, "[PASS] parsed a nul file\n" )
	} else {
		fmt.Fprintf( os.Stderr, "[FAIL] did not parse a nul file: %s\n", err )
		t.Fail()
		return
	}

	s1 := cfg.Extract_str( "template lower default", "not-a-key", "missing-key" )
	if ! is_type( "s1", s1, "string" )  {
		t.Fail()
	}
	if s1 == "missing-key" {
		fmt.Fprintf( os.Stderr, "[PASS] missing key did return the default\n" )
	} else {
		fmt.Fprintf( os.Stderr, "[FAIL] missing key did not return the default: %s\n", s1 )
		t.Fail()
	}

	sp1 := cfg.Extract_p2str( "template lower default", "not-a-key", nil )
	if sp1 == nil {
		fmt.Fprintf( os.Stderr, "[PASS] missing key did return the default pointer\n" )
	} else {
		fmt.Fprintf( os.Stderr, "[FAIL] missing key did not return the default pointer: %v\n", sp1 )
		t.Fail()
	}
	
	sp2 := cfg.Extract_p2str( "template lower default", "not-a-key", "foobar" )
	if sp2 == nil {
		fmt.Fprintf( os.Stderr, "[FAIL] missing key did not return the default pointer: %v\n", sp2 )
		t.Fail()
	} else {
		fmt.Fprintf( os.Stderr, "[PASS] missing key did return the default pointer given a string as default: %s\n", *sp2 )
	}

	def := "foobar"
	sp3 := cfg.Extract_p2str( "template lower default", "not-a-key", &def )
	if sp3 == nil {
		fmt.Fprintf( os.Stderr, "[FAIL] missing key did not return the default pointer: %v\n", sp3 )
		t.Fail()
	} else {
		fmt.Fprintf( os.Stderr, "[PASS] missing key did return the default pointer given a string as default: %s\n", *sp3 )
	}

}

