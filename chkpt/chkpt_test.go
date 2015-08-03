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


package chkpt_test

import (
	"fmt"
	"time"
	"testing"

	"github.com/att/gopkgs/chkpt"
)

func TestChkpt( t *testing.T ) {
	c := chkpt.Mk_chkpt(  "/tmp/foo", 3, 5 );
	c.Add_md5()

	if c == nil {
		t.Errorf( "creation of chkpt obj failed" );
		return;
	}

	err := c.Create( );

	if err != nil {
		t.Errorf( "open file for write failed: %s", err );
		return
	}

	fmt.Fprintf( c, "this is a test! %d\n", time.Now().Unix() );		// Chkpt can be used directly in a format statement if open
	fname, err := c.Close( );

	if err != nil {
		fmt.Printf( "error reported by close: %s\n", err )
		t.Fail()
	}
	fmt.Printf( "output chkpt to: %s\n", fname )
}
