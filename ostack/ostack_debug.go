package ostack

/*
	Mnemonic:	ostack_debug
	Abstract:	Some quick debug functions.
	Author:		E. Scott Daniels
	Date:		7 August 2014
*/

import (
	"fmt"
	"os"
)

/*
	Dump the url if the counter max is not yet reached. 
*/
func dump_url( id string, max int, url string ) {
	if  dbug_url_count < max {
		dbug_url_count++

		fmt.Fprintf( os.Stderr, "%s >>> url= %s\n", id, url )
	}
}

/*
	Dump the raw json array if the counter max has not yet been reached.
*/
func dump_json( id string, max int, json []byte ) {
	if  dbug_json_count < max {
		dbug_json_count++

		fmt.Fprintf( os.Stderr, "%s >>> json= %s\n", id, json )
	}
}

/*
	Allow debugging to be reset or switched off. Set high (20) to 
	turn off, 0 to reset. 
*/
func Set_debugging( count int ) {
	dbug_json_count = count
	dbug_url_count = count
}
