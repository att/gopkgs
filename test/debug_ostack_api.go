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
	Mnemonic:	debug_ostack_auth
	Absract: 	Quick and dirty verification of some of the openstack interface.
				This is a bit more flexible than the test module in ostack as it
				can take url/usr/password from the commandline.
	Author:		E. Scott Daniels
	Date:		7 August 2014

	Mod:		11 Jul 2015 : Changes to support new crack function for v2.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"strings"

	"github.com/att/gopkgs/ostack"
)

func find_in_list( list *string, target *string ) {
	if target != nil || *target != "" {
		if strings.Contains( *list, *target ) {
			fmt.Fprintf( os.Stderr, "[INFO] found in the list: %s\n", *target )
		} else {
			fmt.Fprintf( os.Stderr, "[INFO] NOT found in the list: %s\n", *target )
		}
	}
}

func main( ) {
	var (
		o2 *ostack.Ostack = nil
		all_projects map[string]*string		// list of all projects from keystone needed by several tests
		project_map map[string]*string		// list of projects we belong to

		pwd *string
		usr *string
		url *string
	)

	fmt.Fprintf( os.Stderr, "api debugger: v1.11/19235\n" )
	err_count := 0


	{	p := os.Getenv( "OS_USERNAME" ); usr = &p }			// defaults from environment (NOT project!)
	{	p := os.Getenv( "OS_AUTH_URL" ); url = &p }
	{	p := os.Getenv( "OS_PASSWORD" ); pwd = &p }
	
															// tests are -<capital> except -P
	chost_only := flag.Bool( "c", false, "list only compute hosts" )
	dump_stuff := flag.Bool( "d", false, "dump stuff" )
	host2find := flag.String( "h", "", "host search (if -L)" )
	inc_project := flag.Bool( "i", false, "include project in names" )
	show_latency := flag.Bool( "l", false, "show latency on openstack calls" )
	pwd = flag.String( "p", *pwd, "password" )
	project := flag.String( "P", "", "project for subsequent tests" )
	region := flag.String( "r", "", "region" )
	token := flag.String( "t", "", "token" )
	usr = flag.String( "u", *usr, "user-name" )
	url = flag.String( "U", *url, "auth-url" )
	verbose := flag.Bool( "v", false, "verbose" )
	target_vm := flag.String( "vm", "", "target VM ID" )

	run_all := flag.Bool( "A", false, "run all tests" )				// the various tests
	run_crack := flag.Bool( "C", false, "crack a token" )
	run_endpt := flag.Bool( "E", false, "test endpoint list gen" )
	run_fip := flag.Bool( "F", false, "run fixed-ip test" )
	run_gw_map := flag.Bool( "G", false, "run gw list test" )
	run_mac := flag.Bool( "H", false, "run mac-ip map test" )
	run_info := flag.Bool( "I", false, "run vm info map test" )
	run_if := flag.Bool( "IF", false, "run get interfaces test" )
	run_hlist := flag.Bool( "L", false, "run list-host test" )
	run_maps := flag.Bool( "M", false, "run maps test" )
	run_netinfo := flag.Bool( "N", false, "run netinfo maps" )
	run_user := flag.Bool( "R", false, "run user/role test" )
	run_subnet := flag.Bool( "S", false, "run subnet map test" )
	run_vfp := flag.Bool( "V", false, "run token valid for project test" )
	run_projects := flag.Bool( "T", false, "run projects test" )
	flag.Parse()									// actually parse the commandline

	if *token == "" {
		token = nil
	}


	if *dump_stuff {
		ostack.Set_debugging( -100 )					// resets debugging counts to 0
	}
	if *show_latency {
		ostack.Set_latency_debugging( true )
	}

	if url == nil || usr == nil || pwd == nil {
		fmt.Fprintf( os.Stderr, "usage: debug_ostack_api -U URL -u user -p password [-d] [-i] [-v] [-A] [-F] [-L] [-M] [-T] [-V]\n" )
		fmt.Fprintf( os.Stderr, "usage: debug_ostack_api --help\n" )
		os.Exit( 1 )
	}

	o := ostack.Mk_ostack_region( url, usr, pwd, nil, region )
	if o == nil {
		fmt.Fprintf( os.Stderr, "[FAIL] aborting: unable to make ostack structure\n" )
		os.Exit( 1 )
	}

	fmt.Fprintf( os.Stderr, "[OK]   created openstack interface structure for: %s %s\n", *usr, *url )

	region_str := "default"
	if *region  == "" {
		region = nil
	} else {
		region_str = *region
	}

	err := o.Authorise(  ) 			// generic auth without region since we don't give a project on the default creds
	if err != nil {
		fmt.Fprintf( os.Stderr, "[FAIL] aborting: authorisation failed: region=%s:  %s\n", region_str, err )
		os.Exit( 1 )
	}	

	fmt.Fprintf( os.Stderr, "\n[OK]   authorisation for %s (default creds) successful admin flag: %v\n", *usr, o.Isadmin()  )

	if *project == "" || *run_all || *run_projects {		// map projects that the user belongs to
		m1, _, err := o.Map_tenants( )
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] aborting: unable to generate a project list (required -A or -T tests): %s\n", err )
			fmt.Fprintf( os.Stderr, "Is %s an admin?\n", *usr )
			os.Exit( 1 )
		}
		project_map = m1
	
		if * run_projects {								// only announce if specific project test is on
			if *verbose {
				fmt.Fprintf( os.Stderr, "\n[OK]   project list generation ok:\n" )
			} else {
				fmt.Fprintf( os.Stderr, "\n[OK]   project list map contains %d entries\n", len( m1 ) )
			}
		}
		for k, v := range( m1 ) {
			if *project == "" {							// save first one for use later if user didn't set cmdline flag
				project = &k
			}
			if ! *run_projects {						// only list them if specific test is on
				break
			}
			fmt.Fprintf( os.Stderr, "\t\tproject: %s --> %s\n", k, *v )
		}
	}
	
	if *project != "" {
		fmt.Fprintf( os.Stderr, "[OK]   getting project creds for remaining tests: %s\n", *project )
		o2 = ostack.Mk_ostack_region( url, usr, pwd, project, region )			// project specific creds
		if o2 == nil {
			fmt.Fprintf( os.Stderr, "\n[FAIL] unable to alloc creds for specific project; %s\n", *project )
			os.Exit( 1 )
		}
		err = o2.Authorise( )
		if err != nil {
			fmt.Fprintf( os.Stderr, "\n[FAIL] unable to authorise creds for specific project; %s\n", *project )
			os.Exit( 1 )
		}
	} else {
		fmt.Fprintf( os.Stderr, "[FAIL] did not capture a project name and -P not supplied on command line; cannot attempt any other tests\n" )
		os.Exit( 1 )
	}

	if *run_all || *run_projects {
		fmt.Fprintf( os.Stderr, "[INFO]  sussing host list for each project....\n" )
		for k, _ := range( project_map ) {
			var hlist *string

			o3 :=  ostack.Mk_ostack_region( url, usr, pwd, &k, region )
			o3.Insert_token( o2.Get_token() )
			startt := time.Now().Unix()
			fetch_type := "compute & network"
			if *chost_only {
				hlist, err = o3.List_hosts( ostack.COMPUTE )
				fetch_type = "compute only"
			} else {
				hlist, err = o3.List_hosts( ostack.COMPUTE | ostack.NETWORK )
			}
			endt := time.Now().Unix()
			if err == nil {
				fmt.Fprintf( os.Stderr, "[OK]   got hosts (%s) for %s: %s  (%d sec)\n", k, fetch_type, *hlist, endt - startt )
			} else {
				fmt.Fprintf( os.Stderr, "[WARN] unable to get hosts (%s) for %s: %s  (%d sec)", fetch_type, k, err, endt-startt )
			}
		}
	}

	if *run_projects || *run_all || *run_user {				// needed later for both projects and user so get here first
		all_projects, _, err = o2.Map_all_tenants( )		// map all tenants using keystsone rather than compute service
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] unable to generate a complete tennt list from keystone: %s\n", err )
			all_projects = nil
		}
	}

	if all_projects != nil && (*run_all || *run_projects) {				// see if we can get a full list of projects
		fmt.Fprintf( os.Stderr, "[OK]   ALL project map contains %d entries:\n", len( all_projects ) )
		if *verbose || ! *run_all {
			for k, v := range( all_projects ) {
				fmt.Fprintf( os.Stderr, "\t\tproject: %s --> %s\n", k, *v )
			}

			if *verbose {
				fmt.Fprintf( os.Stderr, "[INFO]  sussing host list for all known project....\n" )
				for k, _ := range( all_projects ) {
					o3 :=  ostack.Mk_ostack_region( url, usr, pwd, &k, region )
					o3.Insert_token( o2.Get_token() )
					startt := time.Now().Unix()
					hlist, err := o3.List_hosts( ostack.COMPUTE )
					endt := time.Now().Unix()
					if err == nil {
						fmt.Fprintf( os.Stderr, "[OK]   got compute hosts for %s: %s  (%d sec)\n", k, *hlist, endt - startt )
					} else {
						fmt.Fprintf( os.Stderr, "[WARN] unable to get compute hosts for %s: %s", k, err )
					}
					startt = time.Now().Unix()
					hlist, err = o3.List_hosts( ostack.NETWORK )
					endt = time.Now().Unix()
					if err == nil {
						fmt.Fprintf( os.Stderr, "[OK]   got network hosts for %s: %s  (%d sec)\n", k, *hlist, endt - startt )
					} else {
						fmt.Fprintf( os.Stderr, "[WARN] unable to get network hosts for %s: %s", k, err )
					}
	
				}
			} else {
				fmt.Fprintf( os.Stderr, "[SKIP]  did not suss host list for all known project (use -T -v to do this)\n" )
			}
		}
	}

	if *run_all || *run_user {
		rm, err := o2.Map_roles() 				// map all roles
		if err == nil {
			fmt.Fprintf( os.Stderr, "[OK]   Roles found: %d\n", len( rm ) )
			for k, v := range( rm ) {
				fmt.Fprintf( os.Stderr, "\trole: %s = %s\n", k, *v );	
			}
		} else {
			fmt.Fprintf( os.Stderr, "[FAIL] unable to generate a role map from keystone: %s\n", err )
			err_count++
		}

/*
invoking groles when it's not supported causes the user roles request to fail with an auth failure.
don't know if openstack is invalidating the token, or what, but it works when global roles isn't
invoked.  bloody openstack.
		rm, err = o2.Map_user_groles() 				// map global roles for the user
		if err == nil {
			fmt.Fprintf( os.Stderr, "[OK]   Global roles found: %d\n", len( rm ) )
			for k, v := range( rm ) {
				fmt.Fprintf( os.Stderr, "\trole: %s = %s\n", k, *v );	
			}
		} else {
			fmt.Fprintf( os.Stderr, "[FAIL] unable to generate a global role map from keystone: %s\n", err )
			err_count++
		}
*/

		rm, err = o2.Map_user_roles( nil ) 			// map the user's roles for the project in the object
		if err == nil {
			fmt.Fprintf( os.Stderr, "[OK]   Project Roles for the user: %d\n", len( rm ) )
			for k, v := range( rm ) {
				fmt.Fprintf( os.Stderr, "\trole: %s = %s\n", k, *v );	
			}
		} else {
			fmt.Fprintf( os.Stderr, "[FAIL] unable to generate a role map from keystone: %s\n", err )
			err_count++
		}

		if *verbose {
			for p, pid := range( all_projects ) {
				if pid == nil {
					fmt.Fprintf( os.Stderr, "pid is nil for %s\n", p )
				} else {
					rm, err := o2.Map_user_roles( pid )
					if err != nil {
						fmt.Fprintf( os.Stderr, "\t%s seems not to be a member of %s\n", *usr, p )
					} else {
						fmt.Fprintf( os.Stderr, "\t%s has %d roles in %s\n", *usr, len( rm ), p )
					}
				}
			}
		}
	}

	if *run_all || *run_hlist {
		startt := time.Now().Unix()
		hlist, err := o2.List_enabled_hosts( ostack.COMPUTE )
		endt := time.Now().Unix()
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error generating enabled compute host list: %s\n", err )
			err_count++
		} else {
			fmt.Fprintf( os.Stderr, "[OK]   enabled compute host list: %s  (%d sec)\n", *hlist, endt - startt )
			find_in_list( hlist, host2find )
		}

		startt = time.Now().Unix()
		hlist, err = o2.List_hosts( ostack.COMPUTE )
		endt = time.Now().Unix()
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error generating compute host list: %s\n", err )
			err_count++
		} else {
			fmt.Fprintf( os.Stderr, "[OK]   compute host list: %s  (%d sec)\n", *hlist, endt - startt )
			find_in_list( hlist, host2find )
		}
	
		startt = time.Now().Unix()
		hlist, err = o2.List_hosts( ostack.NETWORK )
		endt = time.Now().Unix()
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error generating network host list: %s\n", err )
			err_count++
		} else {
			fmt.Fprintf( os.Stderr, "[OK]   network host list: %s  (%d sec)\n", *hlist, endt - startt )
			find_in_list( hlist, host2find )
		}
	
		startt = time.Now().Unix()
		hlist, err = o2.List_hosts( ostack.COMPUTE | ostack.NETWORK )
		endt = time.Now().Unix()
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error generating combined compute and network host list: %s\n", err )
			err_count++
		} else {
			fmt.Fprintf( os.Stderr, "[OK]   comp & net host list: %s (%d sec)\n", *hlist, endt - startt )
			find_in_list( hlist, host2find )
		}
	
		startt = time.Now().Unix()
		hlist, err = o2.List_enabled_hosts( ostack.COMPUTE | ostack.NETWORK )
		endt = time.Now().Unix()
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error generating enabled combined compute and network host list: %s\n", err )
			err_count++
		} else {
			fmt.Fprintf( os.Stderr, "[OK]   enabled comp & net host list: %s (%d sec)\n", *hlist, endt - startt )
			find_in_list( hlist, host2find )
		}

	}

	if *run_all || *run_endpt {
		eplist, err := o2.Map_endpoints( nil )
		if err == nil {
			eplist, err = o2.Map_gw_endpoints( eplist )
		}
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] unable to generate an enpoint list: %s\n", err )
		} else {
			fmt.Fprintf( os.Stderr, "[OK]   endpoint list has %d elements\n", len( eplist ) )
			if *verbose {
				for k, v := range eplist {
					fmt.Fprintf( os.Stderr, "\tep: %s %s\n", k, v )
		
				}
			}
		}
	}

	if *run_all || *run_if {							// interfaces (if) test
		o2.Get_interfaces( target_vm )
	}

	if *run_all || *run_info {
		m, err := o2.Map_vm_info( nil )
		if err == nil {
			fmt.Fprintf( os.Stderr, "[OK]   vm info map has %d entries\n", len( m ) )
			if *verbose {
				for _, v := range m {
					fmt.Fprintf( os.Stderr, "\tvm_info: %s\n", v )
				}
				fmt.Fprintf( os.Stderr, "\n" )
			}
		} else {
			fmt.Fprintf( os.Stderr, "[FAIL] error getting vm info map\n" )
		}
	}

	if !(*run_all || *run_maps) && *run_gw_map {						// just gateway maps (dup below if all maps are generated)
																		// order back:  mac2ip, ip2mac, mac2id, id2mac, id2phost
		gm1, _, gm2, gm3, gm4, gm5, err := o2.Mk_gwmaps( nil, nil, nil, nil, nil, nil, true, *inc_project )
		if err == nil {
			gw_id := ""

			fmt.Fprintf( os.Stderr, "[OK]    generated gateway map with %d entries, gw id map with %d entries\n", len( gm1 ), len( gm2 ) )
			if *verbose {
				fmt.Fprintf( os.Stderr, "\n\tgw mac -> gw-ip\n" )
				for k, v := range( gm1 ) {
					fmt.Fprintf( os.Stderr, "\tgw1: %s --> %s\n", k, *v )
				}

				/* skip ip->mac */
				fmt.Fprintf( os.Stderr, "\n\tgw mac -> gw-id\n" )
				for k, v := range( gm2 ) {
					fmt.Fprintf( os.Stderr, "\tgw2: %s --> %s\n", k, *v )
				}

				fmt.Fprintf( os.Stderr, "\n\tgw-id  -> mac\n" )
				for k, v := range( gm3 ) {
					fmt.Fprintf( os.Stderr, "\tgw3: %s --> %s\n", k, *v )
				}

				fmt.Fprintf( os.Stderr, "\n\tgw-id  --> phost\n" )
				for k, v := range( gm4 ) {
					gw_id = k
					extid, _  := o2.Gw2extid( &k )								// see if we can look up the external network id
					if extid != nil {
						fmt.Fprintf( os.Stderr, "\tgw4: %s --> %s   extid=%s\n", k, *v, *extid )
					} else {
						fmt.Fprintf( os.Stderr, "\tgw4: %s --> %s\n", k, *v )
					}
				}

				fmt.Fprintf( os.Stderr, "\n\tgw-ip  -> phost\n" )
				for k, v := range( gm5 ) {
					fmt.Fprintf( os.Stderr, "\tgw4: %s --> %s\n", k, *v )
				}
			}

			if gw_id != "" {													// just one phost look up to confirm it's working
				host, err := o2.Gw2phost( &gw_id )								// see if we can look up the phost
				if err == nil {
					fmt.Fprintf( os.Stderr, "\n[OK]    lookup of phys host for gateway ID %s:  %s\n", gw_id, *host )
				} else {
					fmt.Fprintf( os.Stderr, "\n[FAIL]  lookup of phys host for  gateway ID %s returned nil: %s\n", gw_id, err )
					err_count++
				}
			}

		} else {
			fmt.Fprintf( os.Stderr, "[FAIL]  error generating gateway map: %s\n", err )
			err_count++
		}

		gl1, err := o2.Mk_gwlist()
		if err == nil {
			fmt.Fprintf( os.Stderr, "[OK]    generated gateway list %d entries\n", len( gl1 ) )
			if *verbose {
				for i, v := range gl1 {
					fmt.Fprintf( os.Stderr, "\t gwlist: [%d] %s\n", i, v )
				}
			}
		} else {
			fmt.Fprintf( os.Stderr, "[FAIL]  error generating gateway list: %s\n", err )
			err_count++
		}

	}

	if  *run_all || *run_netinfo {
		m, err := o2.Mk_netinfo_map( )
		if err == nil {
			fmt.Fprintf( os.Stderr, "\n[OK]  network info map contains %d entries\n", len( m ) )
			if *verbose {
				for k,v := range m {
					fmt.Fprintf( os.Stderr, "\t net_info: %s --> %s\n", k, *v )	
				}
			}
		} else {
			fmt.Fprintf( os.Stderr, "[FAIL]  error generating net info map: %s\n", err )
			err_count++
		}
	}

	if  *run_all || *run_maps {
		m1, m2, m3, m4, m5, err := o2.Mk_vm_maps( nil, nil, nil, nil, nil, *inc_project )
	
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error generating maps: %s\n", err )
		} else {
			gm1, _, _, _, gm2, gm3, err := o2.Mk_gwmaps( nil, nil, nil, nil, nil, nil, true, *inc_project )		// yes this is a dup, but a shorter output of gw info
			if err != nil {
				fmt.Fprintf( os.Stderr, "[FAIL] error generating gw maps: %s\n", err )
			} else {
				if m1 == nil {
					fmt.Fprintf( os.Stderr, "\n[FAIL] unable to alloc map m1 for projet %s\n", *project )
					os.Exit( 1 )
				}
			
				if m2 == nil {
					fmt.Fprintf( os.Stderr, "\n[FAIL] unable to alloc map m2 for projet %s\n", *project )
					os.Exit( 1 )
				}
			
				if m3 == nil {
					fmt.Fprintf( os.Stderr, "\n[FAIL] unable to alloc map m3 for projet %s\n", *project )
					os.Exit( 1 )
				}
			
				if m4 == nil {
					fmt.Fprintf( os.Stderr, "\n[FAIL] unable to alloc map m4 for projet %s\n", *project )
					os.Exit( 1 )
				}

				if m5 == nil {
					fmt.Fprintf( os.Stderr, "\n[FAIL] unable to alloc map m5 for projet %s\n", *project )
					os.Exit( 1 )
				}

				if gm1 == nil || gm2 == nil || gm3 == nil {
					fmt.Fprintf( os.Stderr, "\n[FAIL] unable to alloc gateway maps for projet %s   (%v %v %v)\n", *project, gm2, gm2, gm3 )
					os.Exit( 1 )
				}

				if *verbose {
					fmt.Fprintf( os.Stderr, "\n[OK]   all VM maps were allocated for %s\n", *project )
					for k, v := range( m1 ) {
						fmt.Fprintf( os.Stderr, "\tm1: %s --> %s\n", k, *v )
					}
					fmt.Fprintf( os.Stderr, "\n" )
				
					for k, v := range( m2 ) {
						fmt.Fprintf( os.Stderr, "\tm2: %s --> %s\n", k, *v )
					}
					fmt.Fprintf( os.Stderr, "\n" )
				
					for k, v := range( m3 ) {
						fmt.Fprintf( os.Stderr, "\tm3: %s --> %s\n", k, *v )
					}
					fmt.Fprintf( os.Stderr, "\n" )
				
					for k, v := range( m4 ) {
						fmt.Fprintf( os.Stderr, "\tm4: %s --> %s\n", k, *v )
					}
					fmt.Fprintf( os.Stderr, "\n" )

					for k, v := range( m5 ) {
						fmt.Fprintf( os.Stderr, "\tm5: %s --> %s\n", k, *v )
					}
					fmt.Fprintf( os.Stderr, "\n" )

					for k, v := range( gm1 ) {
						fmt.Fprintf( os.Stderr, "\tgw: %s --> %s\n", k, *v )
					}
					fmt.Fprintf( os.Stderr, "\n" )

					for k, v := range( gm2 ) {
						fmt.Fprintf( os.Stderr, "\tgw: %s --> %s\n", k, *v )
					}
					fmt.Fprintf( os.Stderr, "\n" )

					for k, v := range( gm3 ) {
						fmt.Fprintf( os.Stderr, "\tgw: %s --> %s\n", k, *v )
					}
				} else {
					fmt.Fprintf( os.Stderr, "\tm1 contains %d entries\n", len( m1 ) )
					fmt.Fprintf( os.Stderr, "\tm2 contains %d entries\n", len( m2 ) )
					fmt.Fprintf( os.Stderr, "\tm3 contains %d entries\n", len( m3 ) )
					fmt.Fprintf( os.Stderr, "\tm4 contains %d entries\n", len( m4 ) )
					fmt.Fprintf( os.Stderr, "\tgw contains %d entries\n", len( gm1 ) )
				}
			}
		}
	}

	if *run_all || *run_mac {
		mac2tip,  err := o2.Mk_mac2tip( nil )
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error getting mac-info: %s", err )
			os.Exit( 1 )
		}

	
		if *verbose {
			fmt.Fprintf( os.Stderr, "\n[OK]   mac2tip address info fetched\n" )
			for mac, ip := range mac2tip {
				fmt.Fprintf( os.Stderr, "mac2ip: %-15s  --> %-15s\n", mac, *ip )
			}
			fmt.Fprintf( os.Stderr, "\n" )
		} else {
			fmt.Fprintf( os.Stderr, "\n[OK]   mac2ip map contains %d entries\n", len( mac2tip ) )
		}
	}

	if *run_all || *run_fip {
		ip2fip, fip2ip, err := o2.Mk_fip_maps( nil, nil, *inc_project )
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error getting fip-info: %s", err )
			err_count++
		} else {
			if *verbose {
				fmt.Fprintf( os.Stderr, "\n[OK]   floating IP address info fetched\n" )
				for ip, fip := range ip2fip {
					fmt.Fprintf( os.Stderr, "\tip2fip: %-15s  --> %-15s\n", ip, *fip )
				}
				fmt.Fprintf( os.Stderr, "\n" )
				for ip, fip := range fip2ip {
					fmt.Fprintf( os.Stderr, "\tfip2ip: %-15s  --> %-15s\n", ip, *fip )
				}
			} else {
				fmt.Fprintf( os.Stderr, "\n[OK]   floating IP address info fetched: fip2ip contains %d entries\n", len( fip2ip ) )
			}
		}
	}

	if (*run_all || *run_crack)  {
		if  token != nil { 										// crack the given token
			stuff, err := o.Crack_token( token )						// generic tegu style call (project is not known)
			if err != nil {
				fmt.Fprintf( os.Stderr, "[FAIL] unable to crack the token using unknown project: %s\n", err )
				err_count++
			} else {
				fmt.Fprintf( os.Stderr, "[OK]   token was cracked with unknown project: %s\n", stuff )
			}	

			stuff, err = o.Crack_ptoken( token, project, true  )
			if err != nil {
				fmt.Fprintf( os.Stderr, "[FAIL] unable to crack the token using V2: %s\n", err )
				err_count++
			} else {
				fmt.Fprintf( os.Stderr, "[OK]   token was cracked with V2: %s\n", stuff )
			}	

			stuff, err = o.Crack_ptoken( token, project, false  )
			if err != nil {
				fmt.Fprintf( os.Stderr, "[FAIL] unable to crack the token using V2: %s\n", err )
				err_count++
			} else {
				fmt.Fprintf( os.Stderr, "[OK]   token was cracked with V2: %s\n", stuff )
			}	

		} else {
			fmt.Fprintf( os.Stderr, "[SKIP] did not run crack test, no token provided\n" )
		}
	}

	if (*run_all || *run_vfp)  && token != nil { 					// see if token is valid for the given project

/*
		proj, id, err = o.Token2project( token )
		if proj != nil {
			fmt.Fprintf( os.Stderr, "[OK]   token maps to project using o1 creds: %s == %s\n", *proj, *id )
		} else {
			fmt.Fprintf( os.Stderr, "\n[FAIL] unable to map token to a project using o1 creds: %s\n", err )
		}
*/
		proj, id, err := o2.Token2project( token )
		if proj == nil {
			fmt.Fprintf( os.Stderr, "\n[FAIL] unable to map token to a project using o2 creds: %s\n", err )
		} else {
			state := o2.Equals_id( id )
			fmt.Fprintf( os.Stderr, "[OK]   token maps to a project using o2 creds: %s == %s pids are equal=%v\n", *proj, *id, state )

			result, err := o2.Valid_for_project( token, project )
			if err != nil {
				fmt.Fprintf( os.Stderr, "\n[FAIL] token NOT valid for project (%s): %s\n",  *project, err )
				if *verbose {
					fmt.Fprintf( os.Stderr, "\ttoken = %s\n", *token )
				}
				err_count++
			} else {
				if result {
					ptok := *token						// keep good messages to something smallish
					if len( ptok ) > 50 {
						ptok = ptok[0:50] + "...."
					}
					fmt.Fprintf( os.Stderr, "[OK]   token (%s) valid for project (%s)\n", ptok, *project )
				} else {
					err_count++
					fmt.Fprintf( os.Stderr, "\n[FAIL] token is NOT valid for project (%s) (no error message)\n", *project )
						fmt.Fprintf( os.Stderr, "\ttoken = (%s)\n", *token )
				}
			}

		}
	} else {
		if  *run_vfp  && token == nil {
			fmt.Fprintf( os.Stderr, "\n[INFO]     no token supplied with -V option; test skipped\n" )	
		}
	}

	if *run_all || *run_subnet {
		msn, mgw, err := o2.Mk_snlists( )
		if err == nil {
			if msn != nil {
				fmt.Fprintf( os.Stderr, "\n[OK]       subnet map contained %d entries\n", len( msn ) )
				if *verbose {
					for k, v := range msn {
						fmt.Fprintf( os.Stderr, "\tsnet: %s = %s\n", k, *v )
					}
				}
			}
			if mgw != nil {
				fmt.Fprintf( os.Stderr, "\n[OK]       gateway to cidr map contained %d entries\n", len( mgw ) )
				if *verbose {
					for k, v := range mgw {
						fmt.Fprintf( os.Stderr, "\tgw2cidr: %s = %s\n", k, *v )
					}
				}
			}
		} else {
			if err == nil {
				fmt.Fprintf( os.Stderr, "\n[FAIL]     subnet map was nil\n" )
			} else {
				fmt.Fprintf( os.Stderr, "\n[FAIL]     subnet map generation failed: %s\n", err )
			}
			err_count++
		}
	}
	
	// ----------------------------------------------------------------------------------------------------
	if err_count == 0 {
		fmt.Fprintf( os.Stderr, "\n[OK]     all tests passed\n" )
	} else {
		fmt.Fprintf( os.Stderr, "\n[WARN]   %d errors noticed\n", err_count )
	}

}

