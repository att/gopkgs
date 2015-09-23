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
------------------------------------------------------------------------------------------------
	Mnemonic:	ostack_endpoint
	Abstract:	Manage an endpoint (port/interface/attachment) struct.

	Date:		23 September 2015
	Author:		E. Scott Daniels

	Related:

	Mods:
------------------------------------------------------------------------------------------------
*/

package ostack

import (
	"bytes"
	"fmt"
)

/*
	Create an endpoint struct.
*/
func Mk_endpt( id string, mac string, netid string, proj *string, phost *string ) (*End_pt) {
	ep := &End_pt {
		id: 		&id,
		project: 	proj,
		phost:		phost,
		mac:		&mac,
		network:	&netid,	
	}

	return ep
}

/*
 	Generate endpoint (port/interface) information for a VM. Vmid is either the VM name or the 
	UUID as openstack seems to accept either. Returns a map, indexed by the endpoint UUID
	of each port/interface that is associated with the named VM.
*/
func (o *Ostack) Get_endpoints( vmid *string, phost *string ) ( epmap map[string]*End_pt, err error ) {
	var (
		resp 	generic_response		// unpacked json from response
	)

	if o == nil {
		err = fmt.Errorf( "mk_snlist: openstack creds were nil" )
		return
	}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	if o.chost == nil || *o.chost == "" {		// borderline network/compute, ostack handels via compute interface
		err = fmt.Errorf( "no compute host url to query %s", o.To_str() )
		return
	}

	body := bytes.NewBufferString( "" )
	url := fmt.Sprintf( "%s/servers/%s/os-interface", *o.chost, *vmid )				// chost has project id built-in
	err = o.get_unpacked( url, body, &resp, "vm_ports:" )
	if err != nil {
		return
	}

	epmap = make( map[string]*End_pt )
	for _, a := range resp.Interfaceattachments {
		epmap[a.Port_id] = Mk_endpt( a.Port_id, a.Mac_addr, a.Net_id, o.project, phost )
	}

	return
}

/*
	Implement stringer.
*/
func (ep *End_pt) String( ) (string) {
	return fmt.Sprintf( "uuid=%s phost=%s mac=%s proj=%s netid=%s", *ep.id, *ep.phost, *ep.mac, *ep.project, *ep.network )
}
