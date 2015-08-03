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


package token_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/att/gopkgs/token"
)

func TestToken_count( t *testing.T ) {

	str := "now is the time for all good boys to come to the aid of their country and to see if boys like dogs in the country"
	cmap := token.Tokenise_count( str, " " )

	if cmap == nil {
		fmt.Fprintf( os.Stderr, "tokenise count failed to return a map\n" );
		t.Fail()
	}

	for k, v := range( cmap ) {
		fmt.Fprintf( os.Stderr, "%s == %d\n", k, v )
	}


	emap := make( map[string]int )
	emap["aid"] = 1
	emap["the"] = 3
	emap["for"] = 1
	emap["country"] = 2
	emap["boys"] = 2
	emap["to"] = 3

	fcount := 0
	for k,v := range emap {
		if v == cmap[k] {
			fmt.Fprintf( os.Stderr, "      spot check: %s is good\n", k )
		} else {
			fmt.Fprintf( os.Stderr, "FAIL: spot check: %s  fails: %d != %d\n", k, v, cmap[k] )
			fcount++
		}
	}

	if fcount > 0 {
		fmt.Fprintf( os.Stderr, "spot check failed\n" )
		t.Fail()
	}

	return
}

/*
	Test a case where we have more than 2k tokens in the buffer.
*/
func TestLarge_token_count( t *testing.T ) {

	fmt.Fprintf( os.Stderr, "\n" )

	str := ""
	sep := ""
	for i := 0; i < 4360; i++ {								// build a long string of space separated tokens
		str += fmt.Sprintf( "%s%d", sep, i % 100 )			// 100 unique tokens
		sep = " "
	}

	cmap := token.Tokenise_count( str, " " )
	if len( cmap ) != 100 {
		fmt.Fprintf( os.Stderr, "FAIL: large_token_count: expected 100 unique tokens, found %d\n", len( cmap ) )
		for k := range cmap {								// key is the token
			fmt.Fprintf( os.Stderr, "%s\n", k )
		}
		t.Fail()
	} else {
		fmt.Fprintf( os.Stderr, "OK:   large_token_count: expected 100 unique tokens, found 100\n" )
	}
}

/*
	Test count with an empty string.
*/
func TestToken_count_empty( t *testing.T ) {
	cmap := token.Tokenise_count( " ", " " )
	fmt.Fprintf( os.Stderr, "count of empty string resulted in cmap of %d elements\n\n", len( cmap ) )

	for k, v := range( cmap ) {
		fmt.Fprintf( os.Stderr, "(%s) == %d\n", k, v )
	}
	fmt.Fprintf( os.Stderr, "\n" )
}

func TestToken_qsep_pop( t *testing.T ) {
	str := `hello "world this is" a test where token 2 was quoted`
	expect := "world this is"
	fmt.Fprintf( os.Stderr, "testing: (%s)\n", str )

	ntokens, tokens := token.Tokenise_qpopulated( str, " " )
	if ntokens != 9 {
		fmt.Fprintf( os.Stderr, "FAIL: bad number of tokens, expected 9 got %d from (%s)\n", ntokens, str )
		t.Fail()
	}

	if strings.Index( tokens[1], `"` ) >= 0 {
		fmt.Fprintf( os.Stderr, "FAIL: token 2 contained quotes (%s) in %s\n", tokens[1], str )
		t.Fail()
	}

	if tokens[1] != expect {
		fmt.Fprintf( os.Stderr, "FAIL: token 2 expected to be 'world this is' but was '%s'\n", tokens[1] )
		t.Fail()
	}
	fmt.Fprintf( os.Stderr, "expected: '%s' found: '%s'   [OK]\n", expect, tokens[1] )

	//----------------------------------------------

	str = `hello "world"`									// specific test on 2014/01/13 bug fix
	expect = "world"
	fmt.Fprintf( os.Stderr, "testing: (%s)\n", str )
	ntokens, tokens = token.Tokenise_qpopulated( str, " " )
	if strings.Index( tokens[1], `"` ) >= 0 {
		fmt.Fprintf( os.Stderr, "FAIL: token 2 contained quotes (%s) in (%s)\n", tokens[1], str )
		t.Fail()
	}
	if tokens[1] != expect {
		fmt.Fprintf( os.Stderr, "FAIL: token 2 expected to be 'world this is' but was '%s'\n", tokens[1] )
		t.Fail()
	}
	fmt.Fprintf( os.Stderr, "expected: '%s' found: '%s got %d tokens'   [OK]\n", expect, tokens[1], ntokens )

	//----------------------------------------------
	str = `hello      "world"`									// lots of spaces -- result should be same as previous; 2 tokens
	expect = "world"
	fmt.Fprintf( os.Stderr, "testing: (%s)\n", str )
	ntokens, tokens = token.Tokenise_qpopulated( str, " " )
	if ntokens != 2 {
		fmt.Fprintf( os.Stderr, "FAIL: expected 2 tokens bug %d came back\n", ntokens )
		t.Fail()
	}
	if strings.Index( tokens[1], `"` ) >= 0 {
		fmt.Fprintf( os.Stderr, "FAIL: token 2 contained quotes (%s) in (%s)\n", tokens[1], str )
		t.Fail()
	}
	if tokens[1] != expect {
		fmt.Fprintf( os.Stderr, "FAIL: token 2 expected to be 'world this is' but was '%s'\n", tokens[1] )
		t.Fail()
	}
	fmt.Fprintf( os.Stderr, "expected: '%s' found: '%s got %d tokens'   [OK]\n", expect, tokens[1], ntokens )

	//----------------------------------------------
}


/*
	Test qsep which should return null tokens when multiple separaters exist (e.g. foo,,,bar)
*/
func TestToken_qsep( t *testing.T ) {
	str := `hello "world this is" a test where token 2 was quoted`
	expect := "world this is"

	ntokens, tokens := token.Tokenise_qsep( str, " " )
	if ntokens != 9 {
		fmt.Fprintf( os.Stderr, "FAIL: bad number of tokens, expected 9 got %d from (%s)\n", ntokens, str )
		t.Fail()
	}

	if strings.Index( tokens[1], `"` ) >= 0 {
		fmt.Fprintf( os.Stderr, "FAIL: token 2 contained quotes (%s) in %s\n", tokens[1], str )
		t.Fail()
	}

	if tokens[1] != expect {
		fmt.Fprintf( os.Stderr, "FAIL: token 2 expected to be 'world this is' but was '%s'\n", tokens[1] )
		t.Fail()
	}
	fmt.Fprintf( os.Stderr, "expected: '%s' found: '%s'   [OK]\n", expect, tokens[1] )

	// ------------------------------------------------------------------------------------------------------
	str = `hello "world"`									// specific test on 2014/01/13 bug fix
	_, tokens = token.Tokenise_qsep( str, " " )
	if strings.Index( tokens[1], `"` ) >= 0 {
		fmt.Fprintf( os.Stderr, "FAIL: qsep test token 2 contained quotes (%s) in (%s)\n", tokens[1], str )
		t.Fail()
	} else {
		fmt.Fprintf( os.Stderr, "expected 'world' found '%s'   [OK]\n", tokens[1] )
	}
	

}

// ------------------------------------------------------------------------------------------------------
/*
	Helper function for next real test
*/
func qsep_t2( t *testing.T, str string, expect string, sep string, check_quotes bool ) {
	fmt.Fprintf( os.Stderr, "\nqsep testing with: (%s)\n", str )
	ntokens, tokens := token.Tokenise_qsep( str, "," )

	if ntokens != 4 {
		fmt.Fprintf( os.Stderr, "FAIL: qsep test expected 4 tokens, two nil, received %d tokens\n", ntokens )
		t.Fail()
	} else {
		fmt.Fprintf( os.Stderr, "expected 4 tokens, received %d   [OK]\n", ntokens )
	}
	for i := range tokens {
		fmt.Fprintf( os.Stderr, "\ttoken[%d] (%s)\n", i, tokens[i] )
	}

	if ntokens >= 4 {
		if tokens[1] != "" {
			fmt.Fprintf( os.Stderr, "FAIL: qsep test expected token 2 to be empty: (%s)\n", tokens[1] )
			t.Fail()
		}
		fmt.Fprintf( os.Stderr, "token 2 was nil/empty as expected [OK]\n" )

		if tokens[2] != "" {
			fmt.Fprintf( os.Stderr, "FAIL: qsep test expected token 3 to be empty: (%s)\n", tokens[2] )
			t.Fail()
		}
		fmt.Fprintf( os.Stderr, "token 3 was nil/empty as expected [OK]\n" )
	
		if check_quotes {
			if strings.Index( tokens[3], `"` ) >= 0 {
				fmt.Fprintf( os.Stderr, "FAIL: qsep test token 4 contained quotes (%s) in (%s)\n", tokens[3], str )
				t.Fail()
			} else {
				if tokens[3] == expect {
					fmt.Fprintf( os.Stderr, "expected '%s' found '%s'   [OK]\n", expect, tokens[3] )
				} else {
					fmt.Fprintf( os.Stderr, "FAIL: expected '%s' found '%s'   [OK]\n", expect, tokens[3] )
				}
			}
		}
	}
}

func TestToken_qsep2( t *testing.T ) {
	qsep_t2( t, `hello,,,"world"`, "world", ",", true )			// 4 tokens, two middle ones nil, different quoted token
	qsep_t2( t, `"\"hello\" world",,,"world"`, "world", ",", true )
	qsep_t2( t, `"hello world" more stuff,,,"world"`, "world", ",", true )
	qsep_t2( t, `before stuff "hello world" more stuff,,,"world"`, "world", ",", true )
	qsep_t2( t, `hello,,,"world \"world\""`, `world "world"`, ",", false )
	qsep_t2( t, `"hello,world",,,"world \"world\""`, `world "world"`, ",", false )
	qsep_t2( t, `"hello,world" stuff after,,,"world \"world\""`, `world "world"`, ",", false )
	qsep_t2( t, `hello","world test,,,last token`, `last token`, ",", true )
	qsep_t2( t, `"this is a tken",,,"last token"`, `last token`, " ,", true )
	qsep_t2( t, `foo"",,,"last token"`, `last token`, " ,", true )
}

func TestUnique( t *testing.T ) {
	n, toks := token.Tokeise_qsepu( "foo bar bar man foo goo you who foo too", " " )
	fmt.Fprintf( os.Stderr, "\n" );
	if n != 7 {
		fmt.Fprintf( os.Stderr, "FAIL:  expected 7 unique tokens, got %d", n )
		t.Fail()
	} else {
		fmt.Fprintf( os.Stderr, "OK:    expected 7 unique tokens, got %d", n )
	}
	fmt.Fprintf( os.Stderr, "        unique tokens: " )
	for i := 0; i < n; i++ {
		fmt.Fprintf( os.Stderr, "%s ", toks[i] )
	}
	fmt.Fprintf( os.Stderr, "\n" )
}

