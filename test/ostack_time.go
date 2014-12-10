// simple test to confirm how to convert ostack bloody human readable time into 
// something that is usable. 

package main

import (
	"time"
	"fmt"
)

func main( ) {

	go_std_time := "2006-01-02T15:04:05Z"		// go's reference time in openstack format
	hr_time := "2014-08-18T02:34:48Z"		// time as returned by openstack

	
	 t, _ := time.Parse( go_std_time, hr_time ) 
    fmt.Printf( "%d\n", t.Unix() )
}

