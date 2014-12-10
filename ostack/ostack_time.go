/*
-----------------------------------------------------------------------------------------
	Mnemonic:	ostack_time
	Abstract:	Functions that make it possible to deal  with openstack's uncanny ability 
				to generate only  human readable times.
	Author:		E. Scott Daniels
	Date:		16 August 2014

	Mod:		09 Nov 2014 - Check for nil pointer in Unix_time.
-----------------------------------------------------------------------------------------
*/

package ostack

import (
	"time"
)


const (
	GO_STD_TIME string = "2006-01-02T15:04:05Z"       // go's reference time in openstack format
)

/*
	Accept a human readable string from openstack (presumed to be in the format 
    2014-08-18T02:34:48Z) and convert it into a usable, unix timestamp.

	If there is an error, err is set and the returned time will be 0.
*/
func Unix_time( os_time *string ) ( utime int64, err error ) {
	if os_time == nil || *os_time == "" {
		utime = time.Now().Unix() + 300				// no time given by openstack?  assume 5 minutes
		return
	}

	utime = 0
	t, err := time.Parse( GO_STD_TIME, *os_time )
	if err == nil {
		utime = t.Unix( ) 
	}

	return
}
