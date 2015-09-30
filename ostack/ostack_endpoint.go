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
func Mk_endpt( id string, mac string, ip []*string, netid string, proj *string, phost *string ) (*End_pt) {
	ep := &End_pt {
		id: 		&id,
		project: 	proj,
		phost:		phost,
		mac:		&mac,
		ip:			ip,
		network:	&netid,	
		router:		false,
	}

	return ep
}

/*
 	Generate endpoint (port/interface) information for one VM. Vmid is either the VM name or the
	UUID as openstack seems to accept either. Returns a map, indexed by the endpoint UUID
	of each port/interface that is associated with the named VM.
*/
func (o *Ostack) Get_endpoints( vmid *string, phost *string ) ( epmap map[string]*End_pt, err error ) {
	var (
		resp 	generic_response		// unpacked json from response
		ip []*string
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
	err = o.get_unpacked( url, body, &resp, "get_endpts:" )
	if err != nil {
		return
	}

	epmap = make( map[string]*End_pt )
	for _, a := range resp.Interfaceattachments {
		ln := len( a.Fixed_ips )
		if ln > 0 {
			ip = make( []*string, ln )
			for i, v := range a.Fixed_ips {
				ip[i] = &v.Ip_address
			}
		} else {
			ip = make( []*string, 0, 1 )
		}
		epmap[a.Port_id] = Mk_endpt( a.Port_id, a.Mac_addr, ip, a.Net_id, o.project_id, phost )
	}

	return
}

/*
	Creates a map of endpoints indexed by the endpoint ID for every VM in the project referenced
	by the ostack struct. If umap is passed in, the new endpoints are added to that map otherwise
	a map is created and returned.
*/
func (o *Ostack) Map_endpoints( umap map[string]*End_pt ) ( epmap map[string]*End_pt, err error ) {
	var (
		vm_data	generic_response	// "root" of the response goo after pulling out of json format
	)

	if umap != nil {
		epmap = umap
	} else {
		epmap = make( map[string]*End_pt, 1024 )			// 1024 is a hint, not a hard limit
	}

	if o == nil {
		err = fmt.Errorf( "ostact struct was nil" )
		return
	}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	body := bytes.NewBufferString( "" )

	url := *o.chost + "/servers/detail"
	err = o.get_unpacked( url, body, &vm_data, "get_interfaces:" )
	if err != nil {
		return
	}

	for i := range vm_data.Servers {							// for each vm
		vm := vm_data.Servers[i]
		id := vm.Id

		endpoints, _ := o.Get_endpoints( &id, &vm.Host_name )
		for k, v := range endpoints {	
			epmap[k] = v
		}
	}

	return
}

/*
	Requests router (gateway in openstack lingo) information for the project associated
	with the creds, an builds a list of endpoints for each.  If umap is not nil, then
	the map is added to and returned, otherwise a new map is created.
	Relies on the ostack_net.go functions to make the api call.
*/
func (o *Ostack) Map_gw_endpoints(  umap map[string]*End_pt ) ( epmap map[string]*End_pt, err error ) {
	var (
		ports 	*generic_response	// unpacked json from response
		ip []*string
	)

	epmap  = umap					// ensure something goes back
	if o == nil {
		err = fmt.Errorf( "map_gw_ep: openstack creds were nil" )
		return
	}

	if o.nhost == nil || *o.nhost == "" {
		err = fmt.Errorf( "no network host url to query %s", o.To_str()  )
		return
	}

	ports, err =  o.fetch_gwmac_data( )				// fetch the openstack crap
	if err != nil {
		return
	}

	if epmap == nil {								// create maps if user didn't supply one/both
		epmap = make( map[string]*End_pt )
	}

	for j := range ports.Ports {

		id := ports.Ports[j].Device_id				// MUST duplicate them
		mac := ports.Ports[j].Mac_address
		netid := ports.Ports[j].Network_id
		phost := ports.Ports[j].Bind_host_id
		projid := ports.Ports[j].Tenant_id
		
		ln := len( ports.Ports[j].Fixed_ips )
		if ln > 0 {
			ip = make( []*string, ln )
			for i, v := range ports.Ports[j].Fixed_ips {
				ip[i] = &v.Ip_address
			}
		} else {
			ip = make( []*string, 0, 1 )
		}

		epmap[id] = Mk_endpt( id, mac, ip, netid, &projid, &phost )
		epmap[id].Set_router( true )
	}

	return
}

/*
	Get physical host
*/
func (ep *End_pt) Get_phost( ) ( *string ) {
	if ep == nil {
		return nil
	}

	return ep.phost
}
/*
	Get mac address
*/
func (ep *End_pt) Get_mac( ) ( *string ) {
	if ep == nil {
		return nil
	}

	return ep.mac
}
/*
	Get ip adderess
*/
func (ep *End_pt) Get_ip( n int ) ( *string ) {
	if ep == nil {
		return nil
	}

	if len( ep.ip ) > n && n > 0 {
		return ep.ip[n]
	}

	return nil
}

/*
	Get project id
*/
func (ep *End_pt) Get_project( ) ( *string ) {
	if ep == nil {
		return nil
	}

	return ep.project
}

/*
	Returns true if the endpoint is a router.
*/
func (ep *End_pt) Is_router( ) ( bool ) {
	return ep.router
}

/*
	By default and endpoint isn't a router, but this allows
	the router flag to be set.
*/
func (ep *End_pt) Set_router( flag bool ) {
	ep.router = flag
}

/*
	Implement stringer.
*/
func (ep *End_pt) String( ) (string) {
	s := fmt.Sprintf( "uuid=%s phost=%s mac=%s proj=%s netid=%s rtr=%v ips=[ ", *ep.id, *ep.phost, *ep.mac, *ep.project, *ep.network, ep.router )
	sep := ""
	if len( ep.ip ) > 0 {
		for _, v := range ep.ip {
			s += fmt.Sprintf( "%s%s", sep, *v )
			sep = ", "
		}
	}

	s += " ]"
	return  s
}
