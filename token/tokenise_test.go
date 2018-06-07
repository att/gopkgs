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
	Test qsep which should return null tokens when multiple separators exist (e.g. foo,,,bar)
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
	n, toks := token.Tokenise_qsepu( "foo bar bar man foo goo you who foo too", " " )
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

func TestJsonParsing( t *testing.T ) {
	//s := `{"_context_domain": null, "_msg_id": "01cc4c0579e64b2798ecd028a8c206aa", "_context_quota_class": null, "_context_read_only": false, "_context_request_id": "req-896b224d-a47f-43a9-abbb-d9ef8988cb21", "_context_service_catalog": [], "args": {"objmethod": "save", "args": [], "objinst": {"nova_object.version": "1.5", "nova_object.changes": ["instance_uuid", "network_info"], "nova_object.name": "InstanceInfoCache", "nova_object.data": {"instance_uuid": "80f274fc-8854-469e-8571-abc9caf73ed1", "network_info": "[{\"profile\": {}, \"ovs_interfaceid\": \"3562e20a-4441-490e-bf67-ae1c7b0dc66d\", \"preserve_on_delete\": false, \"network\": {\"bridge\": \"br-int\", \"subnets\": [{\"ips\": [{\"meta\": {}, \"version\": 4, \"type\": \"fixed\", \"floating_ips\": [], \"address\": \"10.7.4.50\"}], \"version\": 4, \"meta\": {\"dhcp_server\": \"10.7.0.2\"}, \"dns\": [{\"meta\": {}, \"version\": 4, \"type\": \"dns\", \"address\": \"135.207.177.11\"}, {\"meta\": {}, \"version\": 4, \"type\": \"dns\", \"address\": \"135.207.179.11\"}], \"routes\": [], \"cidr\": \"10.7.0.0/16\", \"gateway\": {\"meta\": {}, \"version\": 4, \"type\": \"gateway\", \"address\": \"10.7.0.1\"}}], \"meta\": {\"injected\": false, \"tenant_id\": \"65c3e5ee5ee0428caa5e5275c58ead61\"}, \"id\": \"e174ae6a-ef11-45e4-b888-add340e98c4f\", \"label\": \"cloudqos-private\"}, \"devname\": \"tap3562e20a-44\", \"vnic_type\": \"normal\", \"qbh_params\": null, \"meta\": {}, \"details\": {\"port_filter\": true, \"ovs_hybrid_plug\": true}, \"address\": \"fa:16:3e:0a:cf:80\", \"active\": true, \"type\": \"ovs\", \"id\": \"3562e20a-4441-490e-bf67-ae1c7b0dc66d\", \"qbg_params\": null}]"}, "nova_object.namespace": "nova"}, "kwargs": {"update_cells": false}}, "_unique_id": "3083ecbd1a7042f1b663d181a6f9e9bc", "_context_resource_uuid": null, "_context_instance_lock_checked": false, "_context_user": null, "_context_user_id": null, "_context_project_name": null, "_context_read_deleted": "no", "_context_user_identity": "- - - - -", "_reply_q": "reply_8305989e7b6649d6b90007ee5e4aa14c", "_context_auth_token": null, "_context_show_deleted": false, "_context_tenant": null, "_context_roles": [], "_context_is_admin": true, "version": "2.0", "_context_project_id": null, "_context_project_domain": null, "_context_timestamp": "2017-05-25T19:53:36.737176", "_context_user_domain": null, "_context_user_name": null, "method": "object_action", "_context_remote_address": null}`

	s := `{foo,bar,"hello,world","you|me"{`
	
	_, toks := token.Tokenise_qsep( s, "{}" )
	for i, v := range toks {
		fmt.Fprintf( os.Stderr, "[%d] %s\n", i, v )
	}
}

