//vi: sw=4 ts=4:
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


package jsontools_test

import (
	"fmt"
	"testing"
	"os"

	"codecloud.web.att.com/gopkgs/jsontools"
)

func TestJsonblob( t *testing.T ) {
	var (
		jstring	string
		jbytes []byte
	)

	jstring = `{ "actions": [ { "action": "setqueues1", "qdata": [ "queue1", "queue2", "queue3" ] }, { "action": "setqueues2", "qdata": [ "2queue1", "2queue2", "2queue3" ] } ] }`
	jbytes = []byte( jstring );
	jif, err := jsontools.Json2blob( jbytes[:], nil, false )

	if err != nil {
		fmt.Fprintf( os.Stderr, "errors unpacking json: %s    [FAIL]\n", err )
		t.Fail();
		return;
	}

	root_map := jif.( map[string]interface{} )
	alist := root_map["actions"].( []interface{} );

	if alist == nil {
		fmt.Fprintf( os.Stderr, "alist in json was nil  [FAIL]\n" )
		t.Fail();
		return;
	}
	fmt.Fprintf( os.Stderr, "found alist, has %d actions   [OK]\n", len( alist ) )

	for i := range alist {
		action := alist[i].( map[string]interface{} )
		atype := action["action"].( string )
		data := action["qdata"].( []interface{} )

		fmt.Fprintf( os.Stderr, "action %d has type: %s\n", i, atype )
		for j := range data {
			fmt.Fprintf( os.Stderr, " data[%d] = %s\n", j, data[j].( string ) )
		}
	}

	

	fmt.Fprintf( os.Stderr, "===== end blob testing ======\n\n" )
}
