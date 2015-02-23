// vi: sw=4 ts=4:

package jsontools_test

import (
	//"bytes"
	"fmt"
	"testing"
	"os"
	"strings"

	"codecloud.web.att.com/gopkgs/jsontools"
)

func TestJsontools( t *testing.T ) {
	
	var (
		json	string;
		blob []byte
	)

	//json = `{"00:00:00:00:00:00:00:02":{"FM1382646257_18326":{"version":1,"type":"FLOW_MOD","length":88,"xid":0,"match":{"dataLayerDestination":"00:00:00:00:00:00","dataLayerSource":"00:00:00:00:00:00","dataLayerType":"0x0800","dataLayerVirtualLan":-1,"dataLayerVirtualLanPriorityCodePoint":0,"inputPort":0,"networkDestination":"0.0.0.0","networkDestinationMaskLen":0,"networkProtocol":0,"networkSource":"10.0.0.2","networkSourceMaskLen":32,"networkTypeOfService":0,"transportDestination":0,"transportSource":0,"wildcards":4178159},"cookie":45035996273704960,"command":0,"idleTimeout":0,"hardTimeout":0,"priority":32767,"bufferId":-1,"outPort":-1,"flags":0,"actions":[{"type":"SET_TP_DST","length":8,"transportPort":1,"lengthU":8},{"type":"OUTPUT","length":8,"port":-6,"maxLength":32767,"lengthU":8}],"lengthU":88}},"00:00:00:00:00:00:00:01":{"FM1382536816_1011":{"version":1,"type":"FLOW_MOD","length":88,"xid":0,"match":{"dataLayerDestination":"00:00:00:00:00:00","dataLayerSource":"00:00:00:00:00:00","dataLayerType":"0x0800","dataLayerVirtualLan":-1,"dataLayerVirtualLanPriorityCodePoint":0,"inputPort":0,"networkDestination":"0.0.0.0","networkDestinationMaskLen":0,"networkProtocol":0,"networkSource":"10.0.0.2","networkSourceMaskLen":32,"networkTypeOfService":0,"transportDestination":0,"transportSource":0,"wildcards":4178159},"cookie":45035996273704960,"command":0,"idleTimeout":0,"hardTimeout":0,"priority":32767,"bufferId":-1,"outPort":-1,"flags":0,"actions":[{"type":"SET_TP_DST","length":8,"transportPort":1,"lengthU":8},{"type":"OUTPUT","length":8,"port":-6,"maxLength":32767,"lengthU":8}],"lengthU":88}}}`;

/*
json = `{
    "access":{
        "token":{
            "id":"ab48a9efdfedb23ty3494",
            "expires":"2010-11-01T03:32:15-05:00",
            "tenant":{
                "id": "t1000",
                "name": "My Project"
            }
        },
        "user":{
            "id":"u123",
            "name":"jqsmith",
            "roles":[{
                    "id":"100",
                    "name":"compute:admin"
                },
				{
                    "id":"101",
                    "name":"object-store:admin",
                    "tenantId":"t1000"
                }                       
            ],
            "roles_links":[]
        },
        "serviceCatalog":[{
                "name":"Cloud Servers",
                "type":"compute",
                "endpoints":[{
                        "tenantId":"t1000",
                        "publicURL":"https://compute.north.host.com/v1/t1000",
                        "internalURL":"https://compute.north.internal/v1/t1000",
                        "region":"North",
                        "versionId":"1",
                        "versionInfo":"https://compute.north.host.com/v1/",
                        "versionList":"https://compute.north.host.com/"
                    },
                    {
                        "tenantId":"t1000",
                        "publicURL":"https://compute.north.host.com/v1.1/t1000",
                        "internalURL":"https://compute.north.internal/v1.1/t1000",
                        "region":"North",
                        "versionId":"1.1",
                        "versionInfo":"https://compute.north.host.com/v1.1/",
                        "versionList":"https://compute.north.host.com/"
                    }
                ],
                "endpoints_links":[]
            },
            {
                "name":"Cloud Files",
                "type":"object-store",
                "endpoints":[{
                        "tenantId":"t1000",
                        "publicURL":"https://storage.north.host.com/v1/t1000",
                        "internalURL":"https://storage.north.internal/v1/t1000",
                        "region":"North",
                        "versionId":"1",
                        "versionInfo":"https://storage.north.host.com/v1/",
                        "versionList":"https://storage.north.host.com/"
                    },
                    {
                        "tenantId":"t1000",
                        "publicURL":"https://storage.south.host.com/v1/t1000",
                        "internalURL":"https://storage.south.internal/v1/t1000",
                        "region":"South",
                        "versionId":"1",
                        "versionInfo":"https://storage.south.host.com/v1/",
                        "versionList":"https://storage.south.host.com/"
                    }
                ]
            },
            {
                "name":"DNS-as-a-Service",
                "type":"dnsextension:dns",
                "endpoints":[{
                        "tenantId":"t1000",
                        "publicURL":"https://dns.host.com/v2.0/t1000",
                        "versionId":"2.0",
                        "versionInfo":"https://dns.host.com/v2.0/",
                        "versionList":"https://dns.host.com/"
                    }
                ]
            }
        ]
    }
}`
*/

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




