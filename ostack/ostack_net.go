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


/*
------------------------------------------------------------------------------------------------
	Mnemonic:	ostack_net
	Abstract:	Functions to support getting information from the network component (o.nhost)
				of openstack.

	Date:		03 February 2014 (hbdkrd)
	Author:		E. Scott Daniels

	Related:	Doc for agent api request is (of course) not with all of the other openstack
				networking doc, it's here:
					http://specs.openstack.org/openstack/neutron-specs/specs/api/agent_management.html

				used to be here, but this seems broken:
					http://docs.openstack.org/api/openstack-network/2.0/content/List_Agents.html

				Other api doc:
					http://developer.openstack.org/api-ref-networking-v2.html

	Mods:		24 July 2014 - bloody icehouse has no backward compat inasmuch as the complete
					list of gateways cannot be fetched on one call. Users will need to
					make a call per project, and we need to provde the support to update
					an existing map.
				13 Aug 2014 - Added error checking centered round missing urls.
					Moved deprecated code out and added a function which (through a hackish
					request) generates a list of physical hosts with network services running on them.
				30 Aug 2014 - Added more description to error message (object to string output).
				30 Sep 2014 - Looks like bloody openstack is returning host names which are of
					the form "host": "c1r2:1ed04397-35fb-51ca-a932-29d8e09be240". Why can't it
					drop the bloody uuname and make it a separate field. We now assume that
					colon is not a legal character in a host name and will split and drop
					anything after the first colon.
				14 Oct 2014 - Changed list_net_hosts to list only the OVS hosts.
				04 Oct 2014 - Changed list_net_hosts to look for the agent string "quantum-openvswitch-agent"
					to be compatible with grizzly (bloody openstack renaming mid-stream).
				04 Dec 2014 - Now reports network host only if service shows alive.
				10 Dec 2014 - Added support to look up a specific gateway (router) in order to
					suss out it's physical location.
				10 Jan 2015 - Lots of updates to support wa interface.
				21 May 2015 - Now looks for either neutron-l3-agent or neutron-openvswitch-agent
					as an indication that the node is a network supporting node.
				28 Jun 2015 - General cleanup and some dissection of duplicated code.
------------------------------------------------------------------------------------------------
*/

package ostack

import (
	"bytes"
	"fmt"
	"strings"
)

// ----- internal support -------------------------------------------------------------------
/*
	Fetch information for gwmac2xip to use and return a generic_response struct with
	the json expansion already done. We need to fetch in several different spots and
	this keeps the url conststruction in one place.
*/
func (o *Ostack) fetch_gwmac_data( ) ( response *generic_response, err error ) {
	response = &generic_response{}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	url := fmt.Sprintf( "%s/v2.0/ports?device_owner=network:router_interface&tenant_id=%s", *o.nhost, *o.project_id )				// lists just the gateways
	body := bytes.NewBufferString( "" )
	err = o.get_unpacked( url, body, response, "fetch_gwmac:" )

	return
}

// ------------------- public ----------------------------------------------------------------

/*
	Given a gateway ID, make the call to dig out the external network id.
*/
func (o *Ostack) Gw2extid( id *string ) ( extid *string, err error ) {
	var (
		resp generic_response		// top level data mapped from ostack json
	)

	extid = nil
	if o == nil {
		err = fmt.Errorf( "net_netinfo: openstack creds were nil" )
		return
	}
	if id == nil || *id == "" {
		err = fmt.Errorf( "id was not supplied" )
		return
	}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	if o.nhost == nil || *o.nhost == "" {
		err = fmt.Errorf( "no network host url to query %s", o.To_str()  )
		return
	}

	body := bytes.NewBufferString( "" )

	//url := fmt.Sprintf( "%s/v2.0/routers/%s/l3-agents.json", *o.nhost, *id )
	//url := fmt.Sprintf( "%s/v2.0/routers/%s/l3-agents", *o.nhost, *id )
	//url := fmt.Sprintf( "%s/v2.0/routers", *o.nhost )

	url := fmt.Sprintf( "%s/v2.0/routers/%s", *o.nhost, *id )
	err = o.get_unpacked( url, body, &resp, "gw2extid:" )
	if err != nil {
		return
	}

	extid = nil
	if resp.Router != nil &&  resp.Router.External_gateway_info != nil {
		dup_str := resp.Router.External_gateway_info.Network_id
		extid = &dup_str
	} else {
		err = fmt.Errorf( "Router or external gateway info was missing in openstack data" )
	}
	
	return
}

/*
	Given a gateway ID, make the call to dig out the physical host that the gateway lives on.
	(Gateway is Openstack's term for L3 router.)
*/
func (o *Ostack) Gw2phost( id *string ) ( host *string, err error ) {
	var (
		resp generic_response		// top level data mapped from ostack json
	)

	host = nil
	if o == nil {
		err = fmt.Errorf( "net_netinfo: openstack creds were nil" )
		return
	}
	if id == nil || *id == "" {
		err = fmt.Errorf( "id was not supplied" )
		return
	}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	if o.nhost == nil || *o.nhost == "" {
		err = fmt.Errorf( "no network host url to query %s", o.To_str()  )
		return
	}

	body := bytes.NewBufferString( "" )
	url := fmt.Sprintf( "%s/v2.0/routers/%s/l3-agents", *o.nhost, *id )
	err = o.get_unpacked( url, body, &resp, "gw2phost:" )
	if err != nil {
		return
	}


	host = nil
	if resp.Agents != nil  && len ( resp.Agents ) > 0 {
		host = resp.Agents[0].Host
	} else {
		err = fmt.Errorf( "agent list missing in openstack output" )
	}
	
	return
}

/*
	Generate a string containing a space separated list of physical host names which
	are associated with the particular type of agent(s) that are passed in.

	Udup_list is a map of host names that have already been encountered (dups) and should be
	ignored; it can be nil.  The dup map generated is returned.
*/
func (o *Ostack) List_net_hosts( udup_list map[string]bool, limit2neutron bool ) ( hlist *string, dup_map map[string]bool, err error ) {
	var (
		rdata generic_response		// stuff back from openstack
	)

	empty_str := ""
	hlist = &empty_str
	dup_map = udup_list				// ensure it goes back even on error

	if o == nil {
		err = fmt.Errorf( "net_netinfo: openstack creds were nil" )
		return
	}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	if o.nhost == nil || *o.nhost == "" {
		err = fmt.Errorf( "no network host url to query in %s", o.To_str() )
		return
	}

	if dup_map == nil {
		dup_map = make( map[string]bool, 24 )
	}

	body := bytes.NewBufferString( "" )
	url := fmt.Sprintf( "%s/v2.0/agents", *o.nhost )				// nhost is the host where the network service is running
	err = o.get_unpacked( url, body, &rdata, "mk_net_phost:" )
	if err != nil {
		return
	}

	wstr := ""
	sep := ""
	for i := range rdata.Agents {
		if (limit2neutron == false ||
			*rdata.Agents[i].Binary == "neutron-l3-agent"  ||
			*rdata.Agents[i].Binary == "neutron-openvswitch-agent"  ||
			*rdata.Agents[i].Binary == "quantum-l3-agent"  ||
			*rdata.Agents[i].Binary == "quantum-openvswitch-agent") &&
			rdata.Agents[i].Alive {									// list only if service is alive (assume host is also up)

			tokens := strings.SplitN( *rdata.Agents[i].Host, ".", 2 )	// ostack isn't consistent, these might come back fully qualified with domain; strip
			tokens = strings.SplitN( tokens[0], ":", 2 )				// and it sometimes adds :uuid to the name so trash that too
			
			if ! dup_map[tokens[0]] {
				wstr += sep + tokens[0]
				sep = " "
				dup_map[tokens[0]]  = true
			}
		}
	}

	hlist = &wstr
	return
}

/*
	Generate a map that is keyed by the network name with each entrying beign a three tuple, space
	separated, string of: physical net, type (gre,vlan,etc), and segment id.
*/
func (o *Ostack) Mk_netinfo_map( ) ( nmap map[string]*string, err error ) {
	var (
		net_list generic_response	// top level data mapped from ostack json
	)

	nmap = nil
	if o == nil {
		err = fmt.Errorf( "net_netinfo: openstack creds were nil" )
		return
	}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	if o.nhost == nil || *o.nhost == "" {
		err = fmt.Errorf( "no network host url to query %s", o.To_str()  )
		return
	}

	body := bytes.NewBufferString( "" )
	url := fmt.Sprintf( "%s/v2.0/networks", *o.nhost )				// nhost is the host where the network service is running
	err = o.get_unpacked( url, body, &net_list, "mk_netinfo:" )
	if err != nil {
		return
	}

	nmap = make( map[string]*string, 101 )				// size is a hint, not a limit
	for _, n := range net_list.Networks {
		dup_str :=fmt.Sprintf( "%s %s %s %d", n.Id, n.Phys_net, n.Phys_type, n.Phys_seg_id )
		nmap[n.Name] = &dup_str
	}

	return
}

/*
	Reads through the openstack crap passed in, or rquests the crap, and generates
	maps for each gateway (router):
		1) mac -> tenant/ip			(umap_ad)
		2) mac -> gateway-id		(umap_id)
		3) ten/ip -> phost				(umap_phost)

	If the reverse option is set, then
		1) tenant/ip -> mac
		2) gateway-id -> mac
		3) gateway-id -> physical host	(umap_phost)

*/
func (o *Ostack) gwmac2xip(  umap_ad map[string]*string, umap_id map[string]*string, umap_phost map[string]*string, usr_resp *generic_response, inc_tenant bool, reverse bool ) (
	m_ad map[string]*string,
	m_id map[string]*string,
	m_phost map[string]*string,
	err error ) {

	var (
		ports 	*generic_response	// unpacked json from response
		addr	string				// ip address
	)

	m_ad = umap_ad					// ensure that if we bail the original map goes back on return
	m_id = umap_id
	m_phost = umap_phost
	if o == nil {
		err = fmt.Errorf( "net_gwmac2xip: openstack creds were nil" )
		return
	}

	if o.nhost == nil || *o.nhost == "" {
		err = fmt.Errorf( "no network host url to query %s", o.To_str()  )
		return
	}

	if usr_resp != nil {							// user supplied already fetched and unpacked data
		ports = usr_resp							// just use it
	} else {
		ports, err =  o.fetch_gwmac_data( )			// no user supplied stuff, fetch it
		if err != nil {
			return
		}
	}

	if m_ad == nil {								// create maps if user didn't supply one/both
		m_ad = make( map[string]*string )
	}

	if m_id == nil {								
		m_id = make( map[string]*string )
	}

	if m_phost == nil {
		m_phost = make( map[string]*string )
	}

	for j := range ports.Ports {
		if inc_tenant {
			addr = ports.Ports[j].Tenant_id + "/" + ports.Ports[j].Fixed_ips[0].Ip_address
		} else {
			addr = ports.Ports[j].Fixed_ips[0].Ip_address
		}

		dup_addr := addr						// MUST duplicate them
		dup_id := ports.Ports[j].Device_id				// id of the device that this port is on (the router
		dup_mac := ports.Ports[j].Mac_address
		dup_phost := ports.Ports[j].Bind_host_id

		if reverse {
			m_ad[dup_addr] = &dup_mac
			m_id[dup_id] = &dup_mac
			m_phost[dup_addr] = &dup_phost
		} else {
			m_ad[dup_mac] = &dup_addr
			m_id[dup_mac] = &dup_id
			m_phost[dup_id] = &dup_phost
		}
	}

	return
}

/*
	Generates gateway [tenant/]ip to mac and mac to [tenant/]ip maps and gateway-id to mac and mac
	to gateway-id maps.  Needs only one call to openstack to generate all maps.  A fifth map,
	translating uuid to phost, is also generated.

	The u* maps are updated if supplied. If nil is passed, a new map is created.
	Use_project is deprecated and supported only for backwards compatibility.

	If use_project is true, then the request is made using the project_id, otherwise the
	project_id is not submitted. In versions before icehouse, submitting without the project
	ID,  with an admin user ID, resulted in a complete list of gateways. With icehouse, it
	seems that we must request for each project.
*/
func (o *Ostack) Mk_gwmaps( umac2ip map[string]*string,
			uip2mac map[string]*string,
			umac2id map[string]*string,
			umid2mac map[string]*string,
			uid2phost map[string]*string,
			uip2phost map[string]*string,
			inc_tenant bool, use_project bool ) (

			mac2ip map[string]*string,
			ip2mac map[string]*string,
			mac2id map[string]*string,
			id2mac map[string]*string,
			id2phost map[string]*string,
			ip2phost map[string]*string,
			err error ) {
	var (
		response *generic_response
	)

	ip2mac = uip2mac							// ensure we return the user maps on error
	mac2ip = umac2ip
	mac2id = umac2id
	id2mac = umac2id
	id2phost = uid2phost

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	if o.nhost == nil || *o.nhost == "" {
		err = fmt.Errorf( "no network host url to query %s", o.To_str() )
		return
	}

	response, err = o.fetch_gwmac_data( )			// suss out the data from ostack so we can use it multiple times
	if err != nil {
		return
	}

	ip2mac, id2mac, ip2phost, err = o.gwmac2xip( ip2mac, id2mac, id2phost, response, inc_tenant, true )
	if err != nil {
		return
	}
	mac2ip, mac2id, id2phost, err = o.gwmac2xip( mac2ip, mac2id, nil, response, inc_tenant, false )

	return
}

/*
 	Creates a list of IP addresses that are gateways. 	
*/
func (o *Ostack) Mk_gwlist( ) ( gwlist []string, err error ) {
	var (
		ports 	generic_response			// unpacked json from response
		url		string
	)

	gwlist = nil

	if o == nil {
		err = fmt.Errorf( "net_subnets: openstack creds were nil" )
		return
	}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}


	if o.nhost == nil || *o.nhost == "" {
		err = fmt.Errorf( "no network host url to query %s", o.To_str() )
		return
	}

	body := bytes.NewBufferString( "" )
	if o.project != nil {
		url = fmt.Sprintf( "%s/v2.0/ports?device_owner=network:router_interface&tenant_id=%s", *o.nhost, *o.project_id )				// lists just the gateways
	} else {
		url = fmt.Sprintf( "%s/v2.0/ports?device_owner=network:router_interface", *o.nhost )		// before icehouse all are returned on single generic call, so nil project is acceptable
	}

	err = o.get_unpacked( url, body, &ports, "mk_gwlist:" )
	if err != nil {
		return
	}

	gwlist = make( []string, len( ports.Ports ) )
	i := 0
	for j := range ports.Ports {
		gwlist[i] = fmt.Sprintf( "%s %s", ports.Ports[j].Mac_address, ports.Ports[j].Fixed_ips[0].Ip_address )
		i++
	}

	return
}


/*
 	Creates several maps based on subnet information:
		snlist	is a map of subnet information indexed by subnet ID. Each entry in the map is a string of space
				separated values in the following order: Name, Tenant ID, CIDR, Gateway IP.
		gw2cidr is a map of gateway project-id/ipaddress to cidr
*/
func (o *Ostack) Mk_snlists( ) ( snlist map[string]*string, gw2cidr map[string]*string, err error ) {
	var (
		resp 	generic_response		// unpacked json from response
	)

	snlist = nil
	gw2cidr = nil

	if o == nil {
		err = fmt.Errorf( "mk_snlist: openstack creds were nil" )
		return
	}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	if o.nhost == nil || *o.nhost == "" {
		err = fmt.Errorf( "no network host url to query %s", o.To_str() )
		return
	}

	body := bytes.NewBufferString( "" )
	url := fmt.Sprintf( "%s/v2.0/subnets", *o.nhost )				// nhost is the host where the network service is running
	err = o.get_unpacked( url, body, &resp, "mk_snlists:" )
	if err != nil {
		return
	}

	snlist = make( map[string]*string )
	gw2cidr = make( map[string]*string )
	for j := range resp.Subnets {
		list := resp.Subnets[j].Name + " " + resp.Subnets[j].Tenant_id + " " + resp.Subnets[j].Cidr + " " + resp.Subnets[j].Gateway_ip
		snlist[resp.Subnets[j].Id] = &list

		dup_str := resp.Subnets[j].Cidr
		gw2cidr[resp.Subnets[j].Tenant_id + "/" + resp.Subnets[j].Gateway_ip] = &dup_str
	}

	return
}


