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
	Mnemonic:	ostack_vms
	Abstract:	Functions which generate translation maps of VM name and IP to and from
				mac address and VM-ID, and from VM-ID to physical hostname.

				In general VM IDs are mapped to IP addresses or tenant/IP addresses. The
				naming convention ip and tip are used to indicate plain IP and tenant/IP
				respectively.

				In most cases, if the caller passes in a map then the map is extended
				which allows for a map that span multiple tenants to be created.

				TODO: we need a way to map tenant name to tenant ID which is what we
					generate with tenant/IP mappings.  Having a tenant name to ID map
					would allow the user to supply something like cloudqos/daniels1.

	Date:		16 December 2013
	Author:		E. Scott Daniels

	Mod:		18 Apr 2014 - Added support for tenant/ip mappings, and several additional
					maping combinations to allow for Tenant/VM name to MAC mapping.
				27 Apr 2014 - Added functions allowing generation of all maps with a
					single call.
				19 May 2014 - Added floating ip to ip map generation functions.
				28 Jul 2014 - Changed tenant_id to project ID.
				 7 Aug 2014 - Corrected edge case where ostack returns "null" rather than
					omitting the value.
------------------------------------------------------------------------------------------------
*/

package ostack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// ------------- structs that are used to unbundle the json auth response data -------------------
// this is snarfed from doc: http://api.openstack.org/api-ref-compute.html
//

// -- port information -------------

type ip_address struct {
	Subnet_id string
	Ip_address string
}

type port struct {
	Status string
	Name string
	//admin_state_up bool
	Network_id string
	Tenant_id string
	Device_owner string
	//binding:capabilities  // what idiot created a name with a colon in it -- completely unusable here (unless tagged I guess)
	//Binding:vif_type string	// must have been the same idiot who decided on this name too.
	Mac_address string
	Fixed_ips []ip_address
	Id string
	Device_id string	
}

type port_list struct {
	Ports []port
}

// --- floating IP crap ----------
type floatip struct {
	Fixed_ip	*string		// fixed ip of the VM
	//Id		string		// id of the floating address (useless)
	Instance_id	*string		// ID of the VM assigned the IP
	Ip			*string		// the floating ip
	//Pool		string		// useless
}

type floatip_list struct {
	Floating_ips [] floatip
}

//-------------------- interface generation, requires one call per VM. ------------------------------------------

/*
	Get a list of interfaces for a VM.  Requires compute 2.1 interface. 
*/
func (o *Ostack) Get_interfaces( vmid *string ) ( err error ) {
	var (
		vm_data	generic_response	// "root" of the response goo after pulling out of json format
	)

	if o == nil {
		err = fmt.Errorf( "ostact struct was nil" )
		return
	}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	body := bytes.NewBufferString( "" )

	url := fmt.Sprintf( "%s/servers/%s/os-virtual-interfaces", *o.chost, *vmid  )
	err = o.get_unpacked( url, body, &vm_data, "get_interfaces:" )
	if err != nil {
		return
	}

/*
	for i, v := range vm_data.Virtual_interfaces {
		fmt.Fprintf( os.Stderr, ">>>> [%d] %s %s\n", i, vmid, v.Id )
	}
*/


	return
}
	

// ----------- ip -> vm-id and  vm-id -> ip mapping ----------------------------------------------------------------------

/*
	Provides both the ip_vmid and tip_vmid mapping which is the same with one boolean flag check.
	Generates a map that allows for the translation of [tenant/]IP-address to VM-id (uuid).

	usr_jdata allows the caller to get the data from openstack and pass in.  This is used when
	the same openstack json can be used to generate several maps so that only one api request
	needs to be made.

	If the reverse flag is set, then the translation map is from VM-id to address which is similar
	to the output of vm2xip except that only the VM-ids are added to the map as keys.

	This function can provide all of the logic needed to build both the ip and tip translation maps
	with the presence of the inc_tenant flag. When set, the tenant information is added to the IP-address
	as it is used as either the key or value. Further, if save_name is true, the map is generated
	with the name as the value rather than the id.
*/
func (o *Ostack) xip2vmid( deftab map[string]*string, inc_tenant bool, usr_jdata []byte, save_name bool, reverse bool ) ( symtab map[string]*string, err error ) {
	var (
		vm_data	generic_response	// "root" of the response goo after pulling out of json format
		jdata	[]byte				// raw json response data
		addr 	string
	)

	symtab = deftab

	if usr_jdata != nil {
		jdata = usr_jdata
	} else {
		err = o.Validate_auth()						// reauthorise if needed
		if err != nil {
			return
		}

		jdata = nil
		body := bytes.NewBufferString( "" )
	
		url := *o.chost + "/servers/detail"
		dump_url( "1:servers", 10, url )
		jdata, _, err = o.Send_req( "GET",  &url, body )
		dump_json( "servers", 10, jdata )
	
		if err != nil {
			return
		}
	}

	err = json.Unmarshal( jdata, &vm_data )			// unpack the json into jif
	if err != nil {
		dump_json(  fmt.Sprintf( "vm2ip: unpack err: %s\n", err ), 30, jdata )
		return
	}

	if deftab != nil {								// use caller's table if there
		symtab = deftab
	} else {
		symtab = make( map[string]*string )
	}

	for i := range vm_data.Servers {							// for each vm
		have_reverse := false									// we only capture first address when mapping by vmid

		for _, v := range vm_data.Servers[i].Addresses {		// for each network interface (addresses is a poor choice) (does NOT provide the interface uuid)
			var dup_vminfo string

			if len( v ) > 0 {									// it is possible to have an interface with no addresses, so prevent failures
				for j := range v {								// for each address assigned to this interface
					addr = v[j].Addr

					if inc_tenant {
						addr = vm_data.Servers[i].Tenant_id  + "/" + addr
					}

					dup_addr := addr								// MUST assign to a new variable for each declared HERE not on the stack
					if save_name {
						if inc_tenant {
							dup_vminfo = vm_data.Servers[i].Tenant_id + "/" + vm_data.Servers[i].Name
						} else {
							dup_vminfo = vm_data.Servers[i].Name
						}
					} else {
						dup_vminfo = vm_data.Servers[i].Id
					}
		
					if reverse && ! have_reverse {
						symtab[dup_vminfo] = &dup_addr
						have_reverse = true
					} else {
						symtab[dup_addr] = &dup_vminfo
					}
				}
			}
		}
	}

	return
}

/*
	Returns a map allowing for translation of tenant/IP-address to UUID (VM-id)
*/
func (o *Ostack) Mk_tip2vmid( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.xip2vmid( deftab, true, nil, false, false )
}

/*
	Returns a map allowing for translation of IP-address to UUID (VM-id)
*/
func (o *Ostack) Mk_ip2vmid( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.xip2vmid( deftab, false, nil, false, false )
}

/*
	Returns a map allowing for translation of UUID (VMid) to tenant/IP-address
*/
func (o *Ostack) Mk_vmid2tip( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.xip2vmid( deftab, true, nil, false, true )
}

/*
	Returns a map allowing for translation of UUID (VMid) to IP-address.
*/
func (o *Ostack) Mk_vmid2ip( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.xip2vmid( deftab, false, nil, false, true )
}

/*
	Returns a map allowing for translation of tenant/IP-address to vmname
*/
func (o *Ostack) Mk_tip2vm( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.xip2vmid( deftab, true, nil, true, false )
}

/*
	Returns a map allowing for translation of IP-address to vm name
*/
func (o *Ostack) Mk_ip2vm( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.xip2vmid( deftab, false, nil, true, false )
}

//----------- vm-id -> IP  and vm-name -> IP mapping -------------------------------------------------------

/*
	This generates a single map that allows either a VM-id or VM-name to translate to a
	[tenant/]IP address. With the setting of one boolean switch this can provide the
	function for either the ip or tip translation.  This is similar to xip2vmid() function
	except that it produces both sets of keys (vm-name and vm-id) in the same map.

	Because a VM may have multiple interfaces, only the first interface ip address is mapped
	to the name or ID.
*/
func (o *Ostack) vm2xip( deftab map[string]*string, inc_tenant bool, usr_jdata []byte ) ( symtab map[string]*string, err error ) {
	var (
		vm_data	generic_response	// "root" of the response goo after pulling out of json format
		jdata	[]byte				// raw json response data
		addr 	string
		vmname	string
	)

	symtab = deftab

	if usr_jdata != nil {
		jdata = usr_jdata
	} else {
		err = o.Validate_auth()						// reauthorise if needed
		if err != nil {
			return
		}
	
		jdata = nil
		body := bytes.NewBufferString( "" )
	
		url := *o.chost + "/servers/detail"
		dump_url( "2:servers", 10, url )
		jdata, _, err = o.Send_req( "GET",  &url, body )
	
		if err != nil {
			return
		}
	}

	err = json.Unmarshal( jdata, &vm_data )			// unpack the json into jif
	if err != nil {
		dump_json(  fmt.Sprintf( "vm2xip: unpack err: %s\n", err ), 30, jdata )
		return
	}

	if deftab != nil {								// use caller's table if there
		symtab = deftab
	} else {
		symtab = make( map[string]*string )
	}

	for i := range vm_data.Servers {						// for each VM
		for _, v := range vm_data.Servers[i].Addresses {	// for each interface -- addresses is a poor choice of label
			addr = ""
			for j := range v {
				if v[j].Addr != "" {
					addr = v[j].Addr
					break
				}
			}
	
			if addr != "" {
				if inc_tenant {
					addr = vm_data.Servers[i].Tenant_id  + "/" + addr
					vmname = vm_data.Servers[i].Tenant_id  + "/" + vm_data.Servers[i].Name
				} else {
					vmname = vm_data.Servers[i].Name
				}
			
				dup_addr := addr								// MUST create a new string for each hash here rather than a statically defined variable
				dup_name := vmname
			
				symtab[vm_data.Servers[i].Id] = &dup_addr		// two sets of keys: ID and [tenant/]name
				symtab[dup_name] = &dup_addr		

				break											// exit loop through interfaces
			}
		}
	}

	return
}

/*
	Generate a map allowing translation of VM name or VM ID (uuid) to VM IP
*/
func (o *Ostack) Mk_vm2ip( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.vm2xip( deftab, false, nil )
}

/*
	Generate a map allowing translation of VM name or VM ID (uuid) to tenant/VM IP
*/
func (o *Ostack) Mk_vm2tip( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.vm2xip( deftab, true, nil )
}


//----------- vm-id -> vm-name -----------------------------------------------------------

/*
	This generates a map that allows a vm-id (uuid) to be translated to a VM name. If
	reverse is true, then the map translates from name to VM-id. If inc_tenant is true
	then the name is tenant/name.

*/
func (o *Ostack) vmname2vmid( deftab map[string]*string, inc_tenant bool, usr_jdata []byte, reverse bool ) ( symtab map[string]*string, err error ) {
	var (
		vm_data	generic_response	// "root" of the response goo after pulling out of json format
		jdata	[]byte				// raw json response data
		vmname	string
	)

	symtab = deftab

	if usr_jdata != nil {
		jdata = usr_jdata
	} else {
		err = o.Validate_auth()						// reauthorise if needed
		if err != nil {
			return
		}

		jdata = nil
		body := bytes.NewBufferString( "" )
	
		url := *o.chost + "/servers/detail"
		dump_url( "3:servers", 10, url )
		jdata, _, err = o.Send_req( "GET",  &url, body )
	
		if err != nil {
			return
		}
	}

	err = json.Unmarshal( jdata, &vm_data )			// unpack the json into jif
	if err != nil {
		dump_json(  fmt.Sprintf( "vmname2vmid: unpack err: %s\n", err ), 30, jdata )
		return
	}

	if deftab != nil {								// use caller's table if there
		symtab = deftab
	} else {
		symtab = make( map[string]*string )
	}

	for i := range vm_data.Servers {
		if inc_tenant {
			//addr = vm_data.Servers[i].Tenant_id  + "/" + addr
			vmname = vm_data.Servers[i].Tenant_id  + "/" + vm_data.Servers[i].Name
		} else {
			vmname = vm_data.Servers[i].Name
		}

		dup_id := vm_data.Servers[i].Id			// MUST create a new string for each hash here rather than a statically defined variable
		dup_name := vmname
		
		if reverse {
			symtab[dup_id] = &dup_name			// generating id->name map
		} else {
			symtab[dup_name] = &dup_id			// generating name->id map
		}
	}

	return
}

/*
	Generate a map allowing translation of VM name to VM id (uuid).
*/
func (o *Ostack) Mk_vmname2vmid( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.vmname2vmid( deftab, false, nil, false )
}

/*
	Generate a map allowing translation of tenant/VM name to VM id (uuid).
*/
func (o *Ostack) Mk_vmtname2vmid( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.vmname2vmid( deftab, true, nil, false )
}

/*
	Generate a map allowing translation of VM ID (uuid) to VM name.
*/
func (o *Ostack) Mk_vmid2vmname( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.vmname2vmid( deftab, false, nil, true )
}

/*
	Generate a map allowing translation of VM ID (uuid) to tenant/VM name.
*/
func (o *Ostack) Mk_vmid2vmtname( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.vmname2vmid( deftab, true, nil, true )
}


// ---------------- vm -> mac   and mac -> vm mapping -----------------------------------------------------

/*
	Generate a map that allows VMid to mac translation. If the caller passes in a map then the map
	is extended, otherwise a new map is created.
*/
func (o *Ostack) Mk_vmid2mac( def_table map[string]*string ) ( table map[string]*string, err error ) {
	table = nil
	if o == nil {
		err = fmt.Errorf( "net_subnets: openstack creds were nil" )
		return
	}
	
	ip2vmid, err := o.Mk_tip2vmid( nil )
	if err != nil {
		return
	}

	ip2mac, err := o.Mk_tip2mac( nil )
	if err != nil {
		return
	}

	if def_table == nil {						// no user table given to extend, create a new one
		table = make( map[string]*string )
	} else {
		table = def_table
	}

	for ip, mac := range ip2mac {
		vmid := ip2vmid[ip]
		if vmid != nil {
			table[*vmid] = mac
		}
	}

	return
}

/*
	Generates a symtable that maps mac addresses to IP addresses and IP to mac.
	This is accomplished by making a network 'ports' request to openstack and interpreting
	the resulting data.

	Usr_jdata allows the caller to prefetch the data; likely to build several maps from
	the same call.
*/
func (o *Ostack) mac2xip( def_table map[string]*string, inc_tenant bool, usr_jdata []byte, reverse bool ) ( table map[string]*string, err error ) {
	var (
		ports	port_list			// data from ostack
		jdata	[]byte				// raw json response data
	)

	table = def_table
	if o == nil {
		err = fmt.Errorf( "ostack/mac2xip: openstack creds were nil" )
		return
	}

	if usr_jdata != nil {
		jdata = usr_jdata
	} else {
		err = o.Validate_auth()						// reauthorise if needed
		if err != nil {
			fmt.Fprintf( os.Stderr, "ostack/mac2xip: validation error: %s\n", err )		//TESTING
			return
		}
	
		jdata = nil
		body := bytes.NewBufferString( "" )
	
		url := fmt.Sprintf( "%s/v2.0/ports", *o.nhost )		// tennant id is built into chost
		dump_url( "1:ports", 10, url )
		jdata, _, err = o.Send_req( "GET",  &url, body );
	
		if err != nil {
			fmt.Fprintf( os.Stderr, "ostack/mac2xip: error: %s\n", err )		//TESTING
			return
		}
		dump_json(  "mac2xip", 10, jdata )
	}

	err = json.Unmarshal( jdata, &ports )			// unpack the json into jif
	if err != nil {
		//fmt.Fprintf( os.Stderr, "ostack/mac2xip: unable to unpack json: %s\n", err )		//TESTING
		//fmt.Fprintf( os.Stderr, "offending_json=%s\n", jdata )
		dump_json(  fmt.Sprintf( "mac2xip: unpack err: %s\n", err ), 30, jdata )
		return
	}

	if def_table == nil {						// no user table given to extend, create a new one
		table = make( map[string]*string )
	} else {
		table = def_table
	}

	for _, p := range ports.Ports {

		if len( p.Fixed_ips ) >  0 {
			dup_addr := ""
			if inc_tenant {
				dup_addr = p.Tenant_id + "/" + p.Fixed_ips[0].Ip_address		// mac->first ip in the list
			} else {
				dup_addr = p.Fixed_ips[0].Ip_address							// mac->first ip in the list
			}
 			dup_mac := p.Mac_address			
			if reverse {
				for i := range p.Fixed_ips {
					if inc_tenant {
						dup_addr = p.Tenant_id + "/" + p.Fixed_ips[i].Ip_address		
					} else {
						dup_addr = p.Fixed_ips[i].Ip_address
					}
					table[dup_addr] = &dup_mac 									// we'll map all IPs to their mac however
				}
			} else {
				table[p.Mac_address] = &dup_addr
			}
		}
	}

	return
}


/*
	Generate a map that allows for translation from mac to IP. If the caller passes in a map, then
	the map is added to, otherwise a new map is created.
*/
func (o *Ostack) Mk_mac2ip( def_table map[string]*string ) ( table map[string]*string, err error ) {
	return o.mac2xip( def_table, false, nil, false )
}

/*
	Generate a map that allows for translation fro mac to tenant/IP.  If the caller passes in a map, then
	the map is added to, otherwise a new map is created.
*/
func (o *Ostack) Mk_mac2tip( def_table map[string]*string ) ( table map[string]*string, err error ) {
	return o.mac2xip( def_table, true, nil, false )
}

/*
	Generate a map that allows for translation from IP to MAC. If the caller passes in a map, then
	the map is added to, otherwise a new map is created.
*/
func (o *Ostack) Mk_ip2mac( def_table map[string]*string ) ( table map[string]*string, err error ) {
	return o.mac2xip( def_table, false, nil, true )
}

/*
	Generate a map that allows for translation fro tenant/IP  to MAC. If the caller passes in a map, then
	the map is added to, otherwise a new map is created.
*/
func (o *Ostack) Mk_tip2mac( def_table map[string]*string ) ( table map[string]*string, err error ) {
	return o.mac2xip( def_table, true, nil, true )
}


// ------------- physical host mapping --------------------------------------------------------------------------------------

/*
	Generate a map that allows the vmid to be used to map to the physical host.
*/
func (o *Ostack) vmid2host( deftab map[string]*string,  usr_jdata []byte ) ( symtab map[string]*string, err error ) {
	var (
		vm_data	generic_response	// "root" of the response goo after pulling out of json format
		jdata	[]byte				// raw json response data
	)

	symtab = deftab

	if usr_jdata != nil {
		jdata = usr_jdata
	} else {
		err = o.Validate_auth()						// reauthorise if needed
		if err != nil {
			return
		}

		jdata = nil
		body := bytes.NewBufferString( "" )
	
		url := *o.chost + "/servers/detail"
		dump_url( "4:servers", 10, url )
		jdata, _, err = o.Send_req( "GET",  &url, body )
	
		if err != nil {
			return
		}
	}

	err = json.Unmarshal( jdata, &vm_data )			// unpack the json into jif
	if err != nil {
		//fmt.Fprintf( os.Stderr, "ostack/vm2ip: unable to unpack json: %s\n", err )		//TESTING
		//fmt.Fprintf( os.Stderr, "offending_json=%s\n", jdata )
		dump_json(  fmt.Sprintf( "vmid2phost: unpack err: %s\n", err ), 30, jdata )
		return
	}

	if deftab != nil {								// use caller's table if there
		symtab = deftab
	} else {
		symtab = make( map[string]*string )
	}

	for i := range vm_data.Servers {
		dup_id := vm_data.Servers[i].Id			// MUST create a new string for each hash here rather than a statically defined variable
		dup_host := vm_data.Servers[i].Host_name
		
		symtab[dup_id] = &dup_host
	}

	return
}

/*
	Make all 4 possible vm based maps:
		[tenant]/IP <-> vmid	(2 tables)
		vm  -> [tenant/]IP		(1 table maps both [tenant/]VM-name and VM-ID to [tenant/]IP-address)
		vmID -> physical-host	(1 table)

	Parmeters are default maps allowing the maps to be added to (if processing multiple tenants).
	This needs to be invoked for each project (ostack object) in order to get full coverage.
*/
func (o *Ostack) Mk_vm_maps(
			def_vmid2ip map[string]*string,
			def_ip2vmid map[string]*string,
			def_vm2ip map[string]*string,
			def_vmid2host map[string]*string,
			def_ip2vm map[string]*string,
		inc_tenant bool ) (
			vmid2ip map[string]*string,
			ip2vmid map[string]*string,
			vm2ip map[string]*string,
			vmid2host map[string]*string,
			ip2vm map[string]*string,
		err error ) {

	var (
		jdata	[]byte				// raw json response data
	)

	vmid2ip = def_vmid2ip 		// capture defaults if supplied
	ip2vmid = def_ip2vmid
	vm2ip = def_vm2ip
	vmid2host = def_vmid2host
	ip2vm = def_ip2vm

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	jdata = nil									// fetch the necessary data from openstack; data for 4 of the 6 maps
	body := bytes.NewBufferString( "" )
	url := *o.chost + "/servers/detail"
   	dump_url( "5:servers", 10, url )
	jdata, _, err = o.Send_req( "GET",  &url, body )
	dump_json( "vm-maps", 10, jdata )

	if err != nil {
		return
	}

	ip2vmid, err = o.xip2vmid( ip2vmid, inc_tenant, jdata, false, false )
	vmid2ip, err = o.xip2vmid( vmid2ip, inc_tenant, jdata, false, true )
	ip2vm, err = o.xip2vmid( ip2vm, inc_tenant, jdata, true, false )	// map ip to vm name
	vm2ip, err = o.vm2xip( vm2ip, inc_tenant, jdata )					// maps both vmname and vmid to ip
	vmid2host, err = o.vmid2host( vmid2host, jdata )


	// mac2xip needs ports data, so we must fetch from openstack

	return
}

/*
	Make both possible mac maps:
		[tenant/]IP <-> mac		(2 tables)
	
	Because port information seems unrelated to tenant id, this needs to be invoked once
	regardless of how many tenants are involved.
*/
func (o *Ostack) Mk_mac_maps( def_ip2mac map[string]*string, def_mac2ip map[string]*string, inc_tenant bool ) ( ip2mac map[string]*string, mac2ip map[string]*string, err error ) {
	var (
		jdata []byte
	)

	ip2mac = def_ip2mac
	mac2ip = def_mac2ip
	
	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	if o.nhost == nil {
		err = fmt.Errorf( "no network host (nhost) available for: user: %s project: %s", *o.user, *o.project )
	}

	jdata = nil
	body := bytes.NewBufferString( "" )
	url := fmt.Sprintf( "%s/v2.0/ports", *o.nhost )		// tennant id is built into chost
	dump_url( "2:ports", 10, url )
	jdata, _, err = o.Send_req( "GET",  &url, body );

	if err != nil {
		return
	}

	mac2ip, err = o.mac2xip( mac2ip, inc_tenant, jdata, false )
	if err == nil {
		ip2mac, err = o.mac2xip( ip2mac, inc_tenant, jdata, true )
	}

	return
}


// ------------------ floating IP to IP translation -----------------------------------------------------------

/*
	Generate/extend a map of [tenant/]ip addresses to floating ip addresses.
	If reverse is true, then the map is keyed on the floating address instead of the IP address of 	
	the VM that has been assigned the floating address. If jdata is passed in we will use it, else
	we'll make the openstack call to get the data.  This allows forward and reverse maps to be
	generated from a single call to openstack with the driving function doing the single call.
*/
func (o *Ostack) xip2fip( deftab map[string]*string, inc_tenant bool, usr_jdata []byte, reverse bool ) ( symtab map[string]*string, err error ) {
	var (
		fip_list	floatip_list	// the goo returned after unpacking
		jdata	[]byte				// raw json response data
	)

	symtab = deftab

	if usr_jdata != nil {
		jdata = usr_jdata
	} else {
		err = o.Validate_auth()						// reauthorise if needed
		if err != nil {
			return
		}
	
		jdata = nil
		body := bytes.NewBufferString( "" )
	
		url := *o.chost + "/os-floating-ips"
		dump_url( "1:floating-ips", 10, url )
		jdata, _, err = o.Send_req( "GET",  &url, body )
	
		if err != nil {
			return
		}
	}

	err = json.Unmarshal( jdata, &fip_list )			// unpack the json into jif
	if err != nil {
		dump_json(  fmt.Sprintf( "xip2fip: unpack err: %s\n", err ), 30, jdata )
		return
	}

	if deftab != nil {								// use caller's table if there
		symtab = deftab
	} else {
		symtab = make( map[string]*string )
	}

	for i := range fip_list.Floating_ips {
		fip := fip_list.Floating_ips[i]

		if fip.Fixed_ip != nil {					// who knows why it would be
			ip := *fip.Fixed_ip
			if inc_tenant {
				ip = *o.project_id + "/" + ip
			}
			dup_fip := *fip.Ip
	
			if reverse {
				symtab[dup_fip] = &ip
			} else {
				symtab[ip] = &dup_fip
			}
		}
	}
	
	return
}


/*
	Returns a map which translates tenant/ip addresses to floating (external) IP addresses.
	If deftab is not nil it is extended, otherwise a new table is created.
*/
func (o *Ostack) Mk_tip2fip( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.xip2fip( deftab, true, nil, false )	// inc ten, no-reverse
}

/*
	Returns or extends a map that translate ip addresses to floating IP addresses
	If deftab is not nil it is extended, otherwise a new table is created.
*/
func (o *Ostack) Mk_ip2fip( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.xip2fip( deftab, false, nil, false )	// no ten, no reverse
}

/*
	Returns a map which translates floating (external) IP address to tenant/ip addresses.
	If deftab is not nil it is extended, otherwise a new table is created.
*/
func (o *Ostack) Mk_fip2tip( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.xip2fip( deftab, true, nil, true )	// inc ten, reverse
}

/*
	Returns a map which translates floating (external) IP address to ip addresses.
	If deftab is not nil it is extended, otherwise a new table is created.
*/
func (o *Ostack) Mk_fip2ip( deftab map[string]*string ) ( symtab map[string]*string, err error ) {
	return o.xip2fip( deftab, false, nil, true )	// no ten, no-reverse
}


/*
	Generate both the IP to floating IP map and the floating IP to IP map.  Tennant ID is included
	as a part of the IP address if the parameter inc_tenant is true. Using this function reduces
	overhead as only one call to openstack for the information is needed per tenant.
*/
func (o *Ostack) Mk_fip_maps(
		def_ip2fip map[string]*string,
		def_fip2ip map[string]*string,
		inc_tenant bool ) (
			ip2fip map[string]*string,
			fip2ip  map[string]*string,
			err error ) {
	var (
		jdata	[]byte				// raw json response data
	)

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	jdata = nil
	body := bytes.NewBufferString( "" )

	url := *o.chost + "/os-floating-ips"
	dump_url( "2:floating-ips", 10, url )
	jdata, _, err = o.Send_req( "GET",  &url, body )

	if err != nil {
		return
	}

	ip2fip, err = o.xip2fip( def_ip2fip, inc_tenant, jdata, false )
	if err == nil {
		fip2ip, err = o.xip2fip( def_fip2ip, inc_tenant, jdata, true )
	}
		
	return
}
