
/*
	Mnemonic:	debug_ostack_auth
	Absract: 	Quick and dirty verification of some of the openstack interface. 
				This is a bit more flexible than the test module in ostack as it
				can take url/usr/password from the commandline.
	Author:		E. Scott Daniels
	Date:		7 August 2014
*/

package main

import (
	"flag"
	"fmt"
	"os"
	//"sync"
	"time"

	"forge.research.att.com/gopkgs/ostack"
)

func main( ) {
	var (
		o2 *ostack.Ostack = nil
		all_projects map[string]*string		// list of all projects from keystone needed by several tests

		pwd *string
		usr *string
		url *string
	)

	fmt.Fprintf( os.Stderr, "api debugger: v 1.6/1c094\n" )
	err_count := 0


	{	p := os.Getenv( "OS_USERNAME" ); usr = &p }			// defaults from environment (NOT project!)
	{	p := os.Getenv( "OS_AUTH_URL" ); url = &p }
	{	p := os.Getenv( "OS_PASSWORD" ); pwd = &p }
	
	dump_stuff := flag.Bool( "d", false, "dump stuff" )
	inc_project := flag.Bool( "i", false, "include project" )
	pwd = flag.String( "p", *pwd, "password" )
	project := flag.String( "P", "", "project" )
	token := flag.String( "t", "", "token" )
	usr = flag.String( "u", *usr, "user-name" )
	url = flag.String( "U", *url, "auth-url" )
	verbose := flag.Bool( "v", false, "verbose" )

	run_all := flag.Bool( "A", false, "run all tests" )
	run_fip := flag.Bool( "F", false, "run fixed-ip test" )
	run_gw_map := flag.Bool( "G", false, "run gw list test" )
	run_mac := flag.Bool( "H", false, "run mac-ip map test" )
	run_hlist := flag.Bool( "L", false, "run list-host test" )
	run_maps := flag.Bool( "M", false, "run maps test" )
	run_user := flag.Bool( "R", false, "run user/role test" )
	run_vfp := flag.Bool( "V", false, "run token valid for project test" )
	run_projects := flag.Bool( "T", false, "run projects test" )
	flag.Parse()									// actually parse the commandline

	if *token == "" {
		token = nil 
	}


	if *dump_stuff {
		ostack.Set_debugging( 0 )					// resets debugging counts to 0
	}

	if url == nil || usr == nil || pwd == nil {
		fmt.Fprintf( os.Stderr, "usage: debug_ostack_api -U URL -u user -p password [-d] [-i] [-v] [-A] [-F] [-L] [-M] [-T] [-V]\n" )
		os.Exit( 1 )
	}

	o := ostack.Mk_ostack( url, usr, pwd, nil )
	if o == nil {
		fmt.Fprintf( os.Stderr, "[FAIL] aborting: unable to make ostack structure\n" ) 
		os.Exit( 1 )
	}

	fmt.Fprintf( os.Stderr, "[OK]   created openstack interface structure for: %s %s\n", *usr, *url )

	err := o.Authorise( )
	if err != nil {
		fmt.Fprintf( os.Stderr, "[FAIL] aborting: authorisation failed: %s\n", err )
		os.Exit( 1 )
	}	

	fmt.Fprintf( os.Stderr, "\n[OK]   authorisation for %s successful  admin flag: %v\n", *usr, o.Isadmin()  )

	if *project == "" || *run_all || *run_projects {		// map projects that the user belongs to
		m1, _, err := o.Map_tenants( )
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] aborting: unable to generate a tennt list (required for some tests even if -M was not set): %s\n", err )
			fmt.Fprintf( os.Stderr, "Is %s an admin?\n", *usr )
			os.Exit( 1 )
		} 
	
		if * run_projects {
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
		fmt.Fprintf( os.Stderr, "[OK]   project used for remaining tests: %s\n", *project )
		o2 = ostack.Mk_ostack( url, usr, pwd, project )			// project specific creds
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
		fmt.Fprintf( os.Stderr, "[FAIL] did not capture a project name and -P not supplied on commenc line; cannot attempt any other tests\n" )
		os.Exit( 1 )
	}

	if *run_projects || *run_all || *run_user {				// needed later for both projects and user so get here first
		all_projects, _, err = o2.Map_all_tenants( )		// map all tenants using keystsone rather than compute service
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] unable to generate a coomplete tennt list from keystone: %s\n", err )
			all_projects = nil
		} 
	}

	if all_projects != nil && (*run_all || *run_projects) {				// see if we can get a full list of projects
		fmt.Fprintf( os.Stderr, "[OK]   ALL project map contains %d entries:\n", len( all_projects ) )
		if *verbose || ! *run_all {
			for k, v := range( all_projects ) {
				fmt.Fprintf( os.Stderr, "\t\tproject: %s --> %s\n", k, *v )
			}

			for k, _ := range( all_projects ) {
				o3 :=  ostack.Mk_ostack( url, usr, pwd, &k )
				o3.Insert_token( o2.Get_token() )
				startt := time.Now().Unix()
				hlist, err := o3.List_hosts( ostack.COMPUTE )
				endt := time.Now().Unix()
				if err == nil {
					fmt.Fprintf( os.Stderr, "[OK]    got hosts for %s: %s  (%d sec)\n", k, *hlist, endt - startt )
				} //else {
				//	fmt.Fprintf( os.Stderr, "[WARN] unable to get hosts for %s: %s", k, err )
				//}
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
		}

		startt = time.Now().Unix()
		hlist, err = o2.List_hosts( ostack.COMPUTE )
		endt = time.Now().Unix()
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error generating compute host list: %s\n", err )
			err_count++
		} else {
			fmt.Fprintf( os.Stderr, "[OK]   compute host list: %s  (%d sec)\n", *hlist, endt - startt )
		}
	
		startt = time.Now().Unix()
		hlist, err = o2.List_hosts( ostack.NETWORK )
		endt = time.Now().Unix()
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error generating network host list: %s\n", err )
			err_count++
		} else {
			fmt.Fprintf( os.Stderr, "[OK]   network host list: %s  (%d sec)\n", *hlist, endt - startt )
		}
	
		startt = time.Now().Unix()
		hlist, err = o2.List_hosts( ostack.COMPUTE | ostack.NETWORK )
		endt = time.Now().Unix()
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error generating combined compute and network host list: %s\n", err )
			err_count++
		} else {
			fmt.Fprintf( os.Stderr, "[OK]   comp & net host list: %s (%d sec)\n", *hlist, endt - startt )
		}

	}

	if !(*run_all || *run_maps) && *run_gw_map {
		gm1, _, err := o2.Mk_gwmaps( nil, nil, true, *inc_project )
		if err == nil {
			fmt.Fprintf( os.Stderr, "[OK]    generated gateway map with %d entries\n", len( gm1 ) )
			if *verbose {
				for k, v := range( gm1 ) {
					fmt.Fprintf( os.Stderr, "\tgw: %s --> %s\n", k, *v )
				}
			}
		} else {
			fmt.Fprintf( os.Stderr, "[FAIL]  error generating gateway map: %s\n", err )
			err_count++
		}
	}

	if  *run_all || *run_maps {
		m1, m2, m3, m4, err := o2.Mk_vm_maps( nil, nil, nil, nil, *inc_project )

	
		if err != nil {
			fmt.Fprintf( os.Stderr, "[FAIL] error generating maps: %s\n", err )
		} else {
			gm1, _, err := o2.Mk_gwmaps( nil, nil, true, *inc_project )
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

					for k, v := range( gm1 ) {
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
			os.Exit( 1 )
		}

	
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

	if (*run_all || *run_vfp)  && token != nil { 					// see if token is valid for the given project
		proj, id, err := o2.Token2project( token )
		if proj != nil {
			fmt.Fprintf( os.Stderr, "[OK]   token maps to project using o2 creds: %s == %s\n", *proj, *id )
		} else {
			fmt.Fprintf( os.Stderr, "\n[FAIL] unable to map token to a project using o2 creds: %s\n", err )
		}

		proj, id, err = o.Token2project( token )
		if proj != nil {
			fmt.Fprintf( os.Stderr, "[OK]   token maps to project using o1 creds: %s == %s\n", *proj, *id )
		} else {
			fmt.Fprintf( os.Stderr, "\n[FAIL] unable to map token to a project using o1 creds: %s\n", err )
		}

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
	} else {
		if  *run_vfp  && token == nil {
			fmt.Fprintf( os.Stderr, "\n[INFO]     no token supplied with -V option; test skipped\n" )	
		}
	}
	
	if err_count == 0 {
		fmt.Fprintf( os.Stderr, "\n[OK]     all tests passed\n" )
	} else {
		fmt.Fprintf( os.Stderr, "\n[WARN]   %d errors noticed\n", err_count )
	}
}

