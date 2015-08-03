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

