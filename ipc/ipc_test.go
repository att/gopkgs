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


package ipc_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/att/gopkgs/ipc"
)

func TestIpc( t *testing.T ) {
	req := ipc.Mk_chmsg( );
	if req == nil {
		fmt.Fprintf( os.Stderr, "unable to create a request\n" )
		t.Fail();
		return;
	}

	start_ts := time.Now().Unix()
	fmt.Fprintf( os.Stderr, "this test runs for 60 seconds and will generate updates periodically....\n" )
	req.Response_ch = nil;		// use it to keep compiler happy
	

	data := "data for tickled bit";
	ch := make( chan *ipc.Chmsg );
	tklr := ipc.Mk_tickler( 6 );

	_, err := tklr.Add_spot( 30, ch, 30, &data, 0 );			// will automatically start the tickler

	_, err = tklr.Add_spot( 20, ch, 20, &data, 0 );			// will automatically start the tickler
	_, err = tklr.Add_spot( 15, ch, 15, &data, 0 );	
	id, err := tklr.Add_spot( 10, ch, 10, &data, 0 );		// nick the id so we can drop it later
	_, err = tklr.Add_spot( 10, ch, 1, &data, 2 );			// should drive type 1 only twice; 10s apart

	if err != nil {
		fmt.Fprintf( os.Stderr, "unable to add tickle spot: %s\n", err );
		t.Fail();
		return;
	}

	limited_count := 0;
	for count := 0; count < 2;  {
		req = <- ch;						// wait for tickle
		fmt.Fprintf( os.Stderr, "got a tickle: %d type=%d count=%d\n", time.Now().Unix() - start_ts, req.Msg_type, count );
		if req.Msg_type == 30 {
			if count == 0 {
				fmt.Fprintf( os.Stderr, "dropping type 10 from list; no more type 10 should appear\n" )
				tklr.Drop_spot( id );		// drop the 10s tickler after first 30 second one pops
			}
			count++;						// count updated only at 30s point
		}

		if req.Msg_type == 1 {
			if limited_count > 1 {
				fmt.Fprintf( os.Stderr, "limited count tickle was driven more than twice [FAIL]\n" );
				t.Fail();
			}

			limited_count++;
		}

		if req.Msg_type == 10 && count > 0 {
			fmt.Fprintf( os.Stderr, "req 10 driven after it was dropped  [FAIL]\n" );
			t.Fail();
		}
	}


	tklr.Stop();


	// when we get here there should only be three active ticklers in the list, so we should be
	// able to add 3 before we get a full error.
	// add more spots until we max out to test the error logic in tickle
	err = nil;
	for i := 0; i < 3 && err == nil; i++ {
		_, err = tklr.Add_spot( 20, ch, 20, &data, 0 );
		if err != nil {
			fmt.Fprintf( os.Stderr, "early failure when adding more: i=%d %s\n", i, err );
			t.Fail();
			return
		}
	}

	// the table should be full (6 active ticklers now) and this should return an error
	_, err = tklr.Add_spot( 10, ch, 10, &data, 0 );
	if err != nil {
		fmt.Fprintf( os.Stderr, "test to over fill the table resulted in the expected error: %s   [OK]\n", err );
	} else {
		fmt.Fprintf( os.Stderr, "adding a 7th tickle spot didn't cause an error and should have  [FAIL]\n" );
		t.Fail();
	}

}
