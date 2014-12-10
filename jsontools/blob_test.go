// vi: sw=4 ts=4:

package jsontools_test

import (
	//"bytes"
	"fmt"
	"testing"
	"os"

	"forge.research.att.com/gopkgs/jsontools"
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
