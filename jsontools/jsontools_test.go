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
	"strings"

	"github.com/gopkgs/jsontools"
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
}`;
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
	
	fmt.Fprintf( os.Stderr, "blob was returned as expected: (%s)\n]", string( blob ) )
	jc.Add_bytes( []byte(` "weight": "190lbs" } { "height", "7ft10in",` ) )
	blob = jc.Get_blob( )
	fmt.Fprintf( os.Stderr, "blob was returned as expected: (%s)\n]", string( blob ) )

	fmt.Fprintf( os.Stderr, "======================== cache testing # 2 =======================\n" )
	jc = jsontools.Mk_jsoncache( ) 			// 'clear'
	jc.Add_bytes( []byte( ` { "ctype": "action_list", "actions": [ { "atype": "setqueues", "qdata": [ "qos106/fa:de:ad:7a:3a:72,E1res619_00001,2,20000000,20000000,200", "qos102/-128,res619_00001,2,20000000,20000000,200", "qos106/-128,Rres619_00001,2,10000000,10000000,200", "qos102/fa:de:ad:cc:48:f9,E0res619_00001,2,10000000,10000000,200" ], "hosts": [ "qos101", "qos102", "qos103", "qos104", "qos105", "qos106" ] } ] }` ) )
	blob = jc.Get_blob()
	fmt.Fprintf( os.Stderr, "blob was returned as expected: (%s)\n]", string( blob ) )

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
	
	fmt.Fprintf( os.Stderr, "blob was returned as expected: (%s)", string( blob ) )
	
}




