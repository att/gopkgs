
package main

import (
	"fmt"
	"os"
	"strings"
)

func add_phost_suffix( old_list *string, suffix *string ) ( *string ) {
	if suffix == nil || old_list == nil || *old_list == "" {
		return old_list
	}

	nlist := ""
	sep := ""

	htoks := strings.Split( *old_list, " " )
	for i := range htoks {
		if htoks[i] != "" {
		if (htoks[i])[0:1] >= "0" && (htoks[i])[0:1] <= "9" {
			nlist += sep + htoks[i]										// assume ip address, put on as is
		} else {
			if strings.Index( htoks[i], "." ) >= 0 {					// fully qualified name
				dtoks := strings.SplitN( htoks[i], ".", 2 )				// add suffix after first node in the name
				nlist += sep + dtoks[0] + *suffix  + "." + dtoks[1]		
			} else {
				nlist += sep + htoks[i] + *suffix
			}
		}

		sep = " "
		}
	}

	return &nlist
}

func do_it( list string, suffix *string ) {
	
	ns := add_phost_suffix( &list, suffix )
	fmt.Fprintf( os.Stderr, "gave: (%s)\ngot:  (%s)\n\n", list, *ns )
}

func main( ) {
	var suffix *string = nil 

	fmt.Fprintf( os.Stderr, "These should all be the same out as went in:\n" )
	do_it( "c2r3 c2r1 c1r6 o11r32 s3e4 charlie robert", suffix )
	do_it( "c2r3.research.att.com c2r1.att.com c1r6.research.att.com o11r32.research.att.com s3e4.research.att.com charlie.research.att.com robert.research.att.com", suffix )
	do_it( "c2r3 c2r1 c1r6 o11r32 s3e4 charlie robert.research.att.com", suffix )
	do_it( "c2r3 c2r1 192.168.32.45 o11r32 9.23.4.16 charlie robert.research.att.com", suffix )

	s := "-ops"
	fmt.Fprintf( os.Stderr, "=========\nThese should all have the suffix in the right place\n" )
	suffix = &s
	do_it( "", suffix )
	do_it( "  ", suffix )
	do_it( " ", suffix )
	do_it( "c2r3 c2r1 c1r6 o11r32 s3e4 charlie robert", suffix )
	do_it( "c2r3.research.att.com c2r1.att.com c1r6.research.att.com o11r32.research.att.com s3e4.research.att.com charlie.research.att.com robert.research.att.com", suffix )
	do_it( "c2r3 c2r1 c1r6 o11r32 s3e4 charlie robert.research.att.com", suffix )
	do_it( "c2r3 c2r1 192.168.32.45 o11r32 9.23.4.16 charlie robert.research.att.com", suffix )
	do_it( "c2r3", suffix )
}
