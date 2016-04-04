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


package jsontools_test

import (
	"fmt"
	"testing"
	"os"
	"strings"

	"github.com/att/gopkgs/jsontools"
)

func TestJsontools( t *testing.T ) {
	
	var (
		json	string;
		blob []byte
	)


json = `{
    "servers": [
        {
            "accessIPv4": "",
            "accessIPv6": "",
            "addresses": {
                "private": [
                    {
                        "addr": "192.168.0.3",
                        "version": 4
                    }
                ]
            },
            "created": "2012-09-07T16:56:37Z",
            "flavor": {
                "id": "1",
                "links": [
                    {
                        "href": "http://openstack.example.com/openstack/flavors/1",
                        "rel": "bookmark"
                    }
                ]
            },
            "hostId": "16d193736a5cfdb60c697ca27ad071d6126fa13baeb670fc9d10645e",
            "id": "05184ba3-00ba-4fbc-b7a2-03b62b884931",
            "image": {
                "id": "70a599e0-31e7-49b7-b260-868f441e862b",
                "links": [
                    {
                        "href": "http://openstack.example.com/openstack/images/70a599e0-31e7-49b7-b260-868f441e862b",
                        "rel": "bookmark"
                    }
                ]
            },
            "links": [
                {
                    "href": "http://openstack.example.com/v2/openstack/servers/05184ba3-00ba-4fbc-b7a2-03b62b884931",
                    "rel": "self"
                },
                {
                    "href": "http://openstack.example.com/openstack/servers/05184ba3-00ba-4fbc-b7a2-03b62b884931",
                    "rel": "bookmark"
                }
            ],
            "metadata": {
                "My Server Name": "Apache1"
            },
            "name": "new-server-test",
            "progress": 0,
            "status": "ACTIVE",
            "tenant_id": "openstack",
            "updated": "2012-09-07T16:56:37Z",
            "user_id": "fake"
        }
    ]
}`

	blob = []byte( json );
	jsontools.Json2map( blob[:],  nil, true );

	fmt.Fprintf( os.Stderr, "======================== cache testing =======================\n" )
	jc := jsontools.Mk_jsoncache( )
	jc.Add_bytes( []byte(`{ "height": "5ft9in",` ) )
	blob = jc.Get_blob( )
	if blob != nil {
		fmt.Fprintf( os.Stderr, " blob wasn't nil and should have been: (%s)\n", string( blob ) )
		t.Fail( )
		os.Exit( 1 )
	}
	
	jc.Add_bytes( []byte(` "weight": "100lbs" } { "height", "6ft10in",` ) )
	blob = jc.Get_blob( )
	if blob == nil {
		fmt.Fprintf( os.Stderr, " blob was nil and should NOT have been\n" )
		t.Fail( )
		os.Exit( 1 )
	}
	
	fmt.Fprintf( os.Stderr, "blob was returned as expected: (%s)\n", string( blob ) )
	jc.Add_bytes( []byte(` "weight": "190lbs" } { "height", "7ft10in",` ) )
	blob = jc.Get_blob( )
	fmt.Fprintf( os.Stderr, "blob was returned as expected: (%s)\n", string( blob ) )

	fmt.Fprintf( os.Stderr, "======================== cache testing # 2 =======================\n" )
	jc = jsontools.Mk_jsoncache( ) 			// 'clear'
	jc.Add_bytes( []byte( ` { "ctype": "action_list", "actions": [ { "atype": "setqueues", "qdata": [ "qos106/fa:de:ad:7a:3a:72,E1res619_00001,2,20000000,20000000,200", "qos102/-128,res619_00001,2,20000000,20000000,200", "qos106/-128,Rres619_00001,2,10000000,10000000,200", "qos102/fa:de:ad:cc:48:f9,E0res619_00001,2,10000000,10000000,200" ], "hosts": [ "qos101", "qos102", "qos103", "qos104", "qos105", "qos106" ] } ] }` ) )
	blob = jc.Get_blob()
	fmt.Fprintf( os.Stderr, "blob was returned as expected: (%s)\n", string( blob ) )

	fmt.Fprintf( os.Stderr, "======================== cache testing # 3 =======================\n" )
	jc = jsontools.Mk_jsoncache( ) 			// 'clear'
	strs := strings.Split( json, "\n" )
	for i := range strs {
		jc.Add_bytes( []byte( strs[i] + "\n" ) )

		if i < 14 {
			blob = jc.Get_blob( )
			if blob != nil {
				fmt.Fprintf( os.Stderr, " blob was NOT nil and should have been: i=%d\n", i )
				t.Fail( )
				os.Exit( 1 )
			}
		}
	}

	blob = jc.Get_blob( )
	if blob == nil {
		fmt.Fprintf( os.Stderr, " blob was nil and should NOT have been\n" )
		t.Fail( )
		os.Exit( 1 )
	}
	
	fmt.Fprintf( os.Stderr, "blob was returned as expected: (%s)\n", string( blob ) )
	
}


/*
	Test the jtree functions
*/
func TestJtree( t *testing.T ) {
	
	var (
		blob []byte
		fv	float64
		iv  int64
	)


	json := `{
				"accessIPv4": "",
				"accessIPv6": "",
				"addresses": {
					"private": [
						{
							"addr": "192.168.0.1",
							"version": 4
						},
						{
							"addr": "192.168.0.2",
							"version": 4
						}
					]
				},
				"created": "2012-09-07T16:56:37Z",
				"flavor": {
					"id": "1",
					"links": [
						{
							"href": "http://openstack.example.com/openstack/flavors/1",
							"rel": "bookmark"
						}
					]
				},
				"hostId": "16d193736a5cfdb60c697ca27ad071d6126fa13baeb670fc9d10645e",
				"id": "05184ba3-00ba-4fbc-b7a2-03b62b884931",
				"image": {
					"id": "70a599e0-31e7-49b7-b260-868f441e862b",
					"links": [
						{
							"href": "http://openstack.example.com/openstack/images/70a599e0-31e7-49b7-b260-868f441e862b",
							"rel": "bookmark"
						}
					]
				},
				"links": [
					{
						"href": "http://openstack.example.com/v2/openstack/servers/05184ba3-00ba-4fbc-b7a2-03b62b884931",
						"rel": "self"
					},
					{
						"href": "http://openstack.example.com/openstack/servers/05184ba3-00ba-4fbc-b7a2-03b62b884931",
						"rel": "bookmark"
					}
				],
				"metadata": {
					"My Server Name": "Apache1"
				},
				"name": "new-server-test",
				"progress": 0,
				"status": "ACTIVE",
				"tenant_id": "openstack",
				"updated": "2012-09-07T16:56:37Z",
				"user_id": "fake",
				"some_value": 42,
				"some_fvalue": 42.34,
				"array_of_string": [ "hello", "world", "how", "is", "it", "spinning?" ],
				"array_of_int": [ 1, 2, 3, 4, 5, 6 ,7 ], 
				"array_of_float": [ 1.1, 2.2, 3.3, 4.4, 5.5 ]
		}`

	fmt.Fprintf( os.Stderr, "\n------ jtree testing ---------------------\n" )
	blob = []byte( json );
	jtree, err := jsontools.Json2tree( blob )
	if err != nil {
		t.Fail()
		fmt.Fprintf( os.Stderr, "failed to convert json blob to tree: %s\n", err )
		return
	}

	fmt.Fprintf( os.Stderr, "[OK]    Json blob converted to tree\n" )
	s := jtree.Get_string( "hostId" )
	if s != nil {
		fmt.Fprintf( os.Stderr, "[OK]   string fetch: %s\n", *s )
	} else {
		fmt.Fprintf( os.Stderr, "[FAIL] did not find hostId in the json\n" )
		t.Fail()
	}
	// -----------------------------------------------------------------------------

	iv, ok := jtree.Get_int( "some_value" )
	if  ok {
		fmt.Fprintf( os.Stderr, "[OK]   value fetch: %d\n", iv )
		fv, ok := jtree.Get_float( "some_value" )
		if  ok {
			fmt.Fprintf( os.Stderr, "[OK]   value fetch int as float: %.3f\n", fv )
		}
	} else {
		fmt.Fprintf( os.Stderr, "[FAIL] did not find value 'some_value' in the json\n" )
		t.Fail()
	}


	// -----------------------------------------------------------------------------

	fv, ok = jtree.Get_float( "some_fvalue" )
	if ok {
		fmt.Fprintf( os.Stderr, "[OK]   float value fetch: %.3f\n", fv )
		iv, ok = jtree.Get_int( "some_fvalue" )
		if ok {
			fmt.Fprintf( os.Stderr, "[OK]   value fetch as int: %d\n", iv )
		}
	} else {
		fmt.Fprintf( os.Stderr, "[FAIL] did not find value 'some_fvalue' in the json\n" )
		t.Fail()
	}

	sub_tree, ok := jtree.Get_subtree( "addresses" )			// get a subtree whch has an array of objects; run the array printing field from each object
	if ok {
		sub_tree.Dump()
		fmt.Fprintf( os.Stderr, "[OK]   found  subtree 'addresses'\n" )
		l := sub_tree.Get_ele_count( "private" )
		if l > 0 {
			for i := 0; i < l; i++ {
				ao, ok := sub_tree.Get_ele_subtree( "private", i )		// get the object at element i
				if ok {

					st := ao.Get_string( "addr" )			
					if st != nil {
						fmt.Fprintf( os.Stderr, "[OK]   addresses subtree element [%d] (%s)\n", i, *st )
					} else {
						t.Fail()
						fmt.Fprintf( os.Stderr, "[FAIL] addresses subtree element %d didn't have an addr string\n", i )
					}
				} else {
					t.Fail()
					fmt.Fprintf( os.Stderr, "[FAIL] addresses subtree element %d didn't return an object\n", i )
				}
			}
		} else {
			t.Fail()
			fmt.Fprintf( os.Stderr, "[FAIL] addresses subtree didn't have private array, or array size was <= 0\n" )
		}
	} else {
		fmt.Fprintf( os.Stderr, "[FAIL] did not find subtree 'addresses'\n" )
		t.Fail()
	}

	// -----------------------------------------------------------------------------
	loop_state := true
	l := jtree.Get_ele_count( "array_of_string" )
	state := "[FAIL]"
	if l > 0 {
		state = "[OK]"
	}
	fmt.Fprintf( os.Stderr, "%-7s array_of_string reported to have %d elements\n", state, l )
	for i := 0; i < l; i++ {
		st := jtree.Get_ele_string( "array_of_string", i )
		if st != nil {
			fmt.Fprintf( os.Stderr, "\t[%d] %s\n", i, *st )
		} else {
			loop_state = false
		}
	}
	if ! loop_state {
		fmt.Fprintf( os.Stderr, "[FAIL] some elements were not strings\n" )
		t.Fail()
	}


	loop_state = true
	l = jtree.Get_ele_count( "array_of_int" )
	state = "[FAIL]"
	if l > 0 {
		state = "[OK]"
	} else {
		t.Fail()
	}
	fmt.Fprintf( os.Stderr, "%-7s array_of_int reported to have %d elements\n", state, l )
	for i := 0; i < l; i++ {
		v, ok := jtree.Get_ele_int( "array_of_int", i )
		if ok {
			fmt.Fprintf( os.Stderr, "\t[%d] %d\n", i, v)
		} else {
			loop_state = false
			fmt.Fprintf( os.Stderr, "\t[%d] no value returned\n", i)
		}
	}
	if ! loop_state {
		fmt.Fprintf( os.Stderr, "[FAIL] some elements were not integers or could not be converted to integer\n" )
		t.Fail()
	}


	l = jtree.Get_ele_count( "array_of_float" )
	loop_state = true
	state = "[FAIL]"
	if l > 0 {
		state = "[OK]"
	}
	fmt.Fprintf( os.Stderr, "%-7s array_of_float reported to have %d elements\n", state, l )
	for i := 0; i < l; i++ {
		fv, ok := jtree.Get_ele_float( "array_of_float", i )
		if ok {
			fmt.Fprintf( os.Stderr, "\t[%d] %.3f\n", i, fv)
		} else {
			fmt.Fprintf( os.Stderr, "\t[%d] no value returned\n", i)
			loop_state = false
		}
	}
	if ! loop_state {
		fmt.Fprintf( os.Stderr, "[FAIL] some elements were not float or could not be converted to float\n" )
		t.Fail()
	}

}



