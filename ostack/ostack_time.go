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
	Unix_time accepts a human readable string from openstack (presumed to be in the format
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
