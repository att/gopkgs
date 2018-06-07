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

/*
	Mnemonic:	arista
	Abstract: 	Support for interfacing with the arista eAPI.

	Date: 		17 January 2014
	Author: 	E. Scott Daniels

	Mods:		07 Jun 2018 - Fix return bug (line 240)

*/

/*
The arista package contains several methods which allow an application to easily
send requests to an Arista switch which is configured with its HTTPs API enabled.
*/
package arista

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/att/gopkgs/jsontools"
)


type Arista_api struct {
	url		*string
}

/*
	Single switch interface data.  This struct contains a very limited amount of
	information that is returned by the switch for a single interface (port). All
	fields are available directly to the user.
*/
type Swif struct {						// a subset of stuff from a single interface section of output from switch
	Name		string					// all fields are exposed since caller will likely be using them directly
	State		string
	Addr		string
	Mtu			float64
	Bandwidth	float64
}

// ---------------------------------------------------------------------------------------------


// -- a small bit of (private) specialised http support ----------------------------------------------------
/*
	Build a client struct that is set to discard the certificate that is received when
	accessing https sites.  The arista switch doesn't have a valid cert, and this keeps
	us from having to load and manage the cert on any machine where we may need to access
	the switch.
*/
func build_client( ) ( *http.Client ) {
 	tr := &http.Transport {
        TLSClientConfig: &tls.Config {
			InsecureSkipVerify: true,					// prevents cert verification and allows us to hit the site with a bad cert
		},
    }

    return  &http.Client { Transport: tr }
}

/*
	Send a request off, wait for the response  and return the raw data.
*/
func send_req( method string, url *string, data *bytes.Buffer ) (jdata []byte, err error) {
	var (
		req 	*http.Request;
		rsrc	*http.Client;		// request source	
	)
	
	jdata = nil;
	req, err = http.NewRequest( method, *url, data );
	if err != nil {
		fmt.Fprintf( os.Stderr, "error making request for %s to %s\n", method, *url );
		return
	}
	req.Header.Set( "Content-Type", "application/json" )

	rsrc = build_client( )
	resp, err := rsrc.Do( req )
	if err == nil {
		jdata, err = ioutil.ReadAll( resp.Body )
		resp.Body.Close( )
	} else {
		fmt.Fprintf( os.Stderr, "send_req: received err response %s\n", err );
	}

	return;
}


/*
	From an arista json mess, pull out the idx-th result with the named string.
*/
func getresult( jif interface{}, idx int, rname string ) (map[string]interface{}, error) {
	var (
		err error
	)

	err = fmt.Errorf( "result data for %s was not the expected type", rname )

	thing, ok := jif.(map[string]interface{}) 				// outside of if to keep thing round after
	if !ok {
		return nil, err
	}

	result, ok := thing["result"].( []interface{} )
	if !ok  {
		return nil, err
	}

	robj, ok := result[idx].( map[string]interface{} )
	if !ok {
		return nil, err
	}

	user_thing, ok := robj[rname].( map[string]interface{} )
	if ! ok {
		return nil, err
	}

	return user_thing, nil
}


// ------------------ public interface ----------------------------------------------

/*
	Mk_aristaif creates a small object which is then used for all requests to a specific
	switch.
*/
func Mk_aristaif( usr *string, pw *string, host *string, port *string ) (aif * Arista_api) {

	target_url := fmt.Sprintf( "https://%s:%s@%s:%s/command-api", *usr, *pw, *host, *port )

	aif = &Arista_api {
		url: &target_url,
	}
		
	return
}


/*
	Submit_req sends a command set to the switch and returns the raw result.
	If text request is true, we change the format to text in the request. (There are some
	Arista commands that do not support returning json output and thus the text_req
	parameter must be set to true in order for the command to work.  Err will be non-nil
	if any error was detected.

	The cmds parameter is a chain of commands to be executed by the switch API.  Each command
	must be supplied as a quoted substring, and if multiple substrings are contained in the
	string they must be comma separated.  For example, the following command string makes
	a query to enable openflow:

		`"configure", "openstack", "no shutdown"`

	Which has the effect of entering all three of those commands on the Arista command line.
*/
func (aif *Arista_api) Submit_req( cmds *string, text_req bool ) ( raw_json []byte, err error ) {
	var (
		json_req_suffix string			// last part of the json sent; depends on text or json output
	)

	raw_json = nil
	if aif == nil || aif.url == nil {
		err = fmt.Errorf( "nil interface object pointer supplied" )
	}

	json_req_prefix := `{ "jsonrpc": "2.0", "method": "runCmds", "params": { "version": 1, "cmds": [ `	// comma separated list of commands in braces
	if text_req {
		json_req_suffix = ` ], "format": "text" }, "id": 1 }`
	} else {
		json_req_suffix = ` ], "format": "json" }, "id": 1 }`
	}

	cmd_str := fmt.Sprintf( "%s %s %s", json_req_prefix, *cmds, json_req_suffix )
	request_body := bytes.NewBufferString( cmd_str )

	raw_json, err = send_req( "POST", aif.url, request_body )
	return
}

/*
	Get_interfaces sends a request to the switch and generate a map of single interface structs that is
	indexed by the interface name.  The single interface struct contains very limited information compared
	to what is returned by the switch.
*/
func (aif *Arista_api) Get_interfaces( ifstate string ) ( ifmap map[string]*Swif, err error ) {

	var (
		tag	string = "root"
	)

	ifmap = make( map[string]*Swif )

	query := `"show interfaces"`
	raw_json, err := aif.Submit_req( &query, false )
	if err != nil {
		return
	}

	jif, err := jsontools.Json2blob( raw_json, &tag, jsontools.NOPRINT ); 	// parse the json into an interface hierarchy
	if err != nil {
		return
	}

	ifaces, err := getresult( jif, 0, "interfaces" )			// pull the list of switch interfaces into an array of interface type, one for each interface
	if err != nil {
		return
	}

	for ik, _ := range ifaces {										// ifaces should be an array of struct
		sif, ok := ifaces[ik].( map[string]interface{} )			// single interface data
		if !ok {
			err = fmt.Errorf( "single interface %s was not expected type", ik )
			return
		}

		if ifstate == "all"  || sif["interfaceStatus"].(string) == ifstate {
			ifmap[ik] = &Swif{
				Name: sif["name"].(string),
				Mtu: sif["mtu"].(float64),
				State: sif["interfaceStatus"].(string),
				Bandwidth: sif["bandwidth"].(float64),
				Addr: sif["physicalAddress"].(string),
			}
		}
	}

	return
}
