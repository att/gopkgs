// vi: sw=4 ts=4:

/*
------------------------------------------------------------------------------------------------
	Mnemonic:	ostack_vminfo
	Abstract:	Functions that generate and/or operate on a VM_info struct.

	Date:		01 May 2015
	Author:		E. Scott Daniels

	Mod:
------------------------------------------------------------------------------------------------
*/

package ostack

import (
	"bytes"
	"encoding/json"
	"fmt"
)

/*
	Returns a map of VM information keyed by VM id.
	If umap is passed in (not nil), then the information is added to that map, otherwise
	a new map is created. 
*/
func (o *Ostack) Map_vm_info( umap map[string]*VM_info ) ( info map[string]*VM_info, err error ) {
	var (
		vm_data	generic_response	// "root" of the response goo after pulling out of json format
		jdata	[]byte				// raw json response data
	)

	if umap != nil {
		info = umap
	} else {
		info = make( map[string]*VM_info, 256 )			// 256 is a hint, not a hard limit
	}

	if o == nil {
		err = fmt.Errorf( "ostact struct was nil" )
		return
	}

	err = o.Validate_auth()						// reauthorise if needed
	if err != nil {
		return
	}

	jdata = nil
	body := bytes.NewBufferString( "" )

	url := *o.chost + "/servers/detail"
	dump_url( "get_vm_info", 10, url )
	jdata, _, err = o.Send_req( "GET",  &url, body ) 
	dump_json( "get_vm_info", 10, jdata )

	if err != nil {
		return
	}

	err = json.Unmarshal( jdata, &vm_data )			// unpack the json into jif
	if err != nil {
		dump_json(  fmt.Sprintf( "get_vm_info: unpack err: %s\n", err ), 30, jdata )
		return
	}


	// TODO -- add address information 
	for i := range vm_data.Servers {							// for each vm
		vm := vm_data.Servers[i]
		id := vm.Id
		info[id] = &VM_info {
			id:			id,										// must carry id so String and To_json work
			zone:		vm.Azone,
			created:	vm.Created,
			flavour:	vm.Flavor.Id,
			hostid:		vm.Hostid,
			host_name:	vm.Host_name,
			image:		vm.Image.Id,
			name:		vm.Name,
			status:		vm.Status,
			tenant_id:	vm.Tenant_id,
			updated:	vm.Updated,
			launched:	vm.Launched,
    		terminated:	vm.Terminated,
		}
	}

	return
}


func (vi *VM_info) Get_name() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.name
}

func (vi *VM_info) Get_zone() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.zone
}

func (vi *VM_info) Get_flavour() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.flavour
}

func (vi *VM_info) Get_created() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.created
}

func (vi *VM_info) Get_hostid() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.hostid
}

func (vi *VM_info) Get_hostname() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.host_name
}

func (vi *VM_info) Get_image() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.image
}

func (vi *VM_info) Get_status() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.status}

func (vi *VM_info) Get_tenantid() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.tenant_id
}

func (vi *VM_info) Get_updated() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.updated}

func (vi *VM_info) Get_launched() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.launched}

func (vi *VM_info) Get_terminated() ( string ) {
	if vi == nil {
		return ""
	}

	return vi.terminated
}


/*
	Implements the string interface. 
*/
func (vi *VM_info) String() ( string ) {
	s := ""
	if vi == nil {
		return s
	}	

	return fmt.Sprintf( "%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", 
			vi.id, vi.name, vi.hostid, vi.host_name, vi.status,
			vi.tenant_id, vi.flavour, vi.image, vi.zone, vi.created,
			vi.launched, vi.updated, vi.terminated )
}


/*
	Generates a json string representing the struct.
*/
func (vi *VM_info) To_json() ( string ) {
	if vi == nil {
		return "{ }"
	}	

	return fmt.Sprintf( `{ "id": %q, "name": %q, "hostid": %q, "host_name": %q, "status": %q, "tenant_id": %q, "flavour": %q, "image": %q, "zone": %q, "created": %q, "launched": %q, "updated": %q, "terminated": %q }`, 
			vi.id, vi.name, vi.hostid, vi.host_name, vi.status, vi.tenant_id, vi.flavour, vi.image, vi.zone, vi.created,
			vi.launched, vi.updated, vi.terminated )
}


