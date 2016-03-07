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
	ch := make( chan *ipc.Chmsg, 10 );						// allow 10 messages to queue on the channel
															// to test non-blocking aspect, set to 1 test should run longer than 60 seconds

	tklr := ipc.Mk_tickler( 6 );							// allow max of 6 ticklers; we test 'full' error at end

	tklr.Add_spot( 35, ch, 30, &data, 0 );		// will automatically start the tickler
	tklr.Add_spot( 20, ch, 20, &data, 0 );
	tklr.Add_spot( 15, ch, 15, &data, 0 );	
	id, err := tklr.Add_spot( 10, ch, 10, &data, 0 );		// nick the id so we can drop it later
	_, err = tklr.Add_spot( 10, ch, 1, &data, 2 );			// should drive type 1 only twice; 10s apart

	if err != nil {
		fmt.Fprintf( os.Stderr, "unable to add tickle spot: %s\n", err );
		t.Fail();
		return;
	}

	fmt.Fprintf( os.Stderr, "type 10 and type 1 written every 10 seconds; type 1 only written twice\n" );
	fmt.Fprintf( os.Stderr, "type 10 will be dropped after 35 seconds\n" )
	fmt.Fprintf( os.Stderr, "type 15 will appear every 15 seconds\n" )
	fmt.Fprintf( os.Stderr, "type 20 will appear every 20 seconds\n" )
	fmt.Fprintf( os.Stderr, "type 30 will appear every 35 seconds\n" )

	limited_count := 0;
	for count := 0; count < 2;  {
		req = <- ch;						// wait for tickle
		fmt.Fprintf( os.Stderr, "got a tickle: @%ds type=%d count=%d\n", time.Now().Unix() - start_ts, req.Msg_type, count );
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


func add_nbmsg( ch chan *ipc.Chmsg, mt int ) {
	r := ipc.Mk_chmsg()
	fmt.Fprintf( os.Stderr, "\tsending message type %d\n", mt );
	r.Send_nbreq( ch, nil, mt, nil, nil )
	fmt.Fprintf( os.Stderr, "\tsent message type %d\n", mt );
}

/*
	Test the non-blocking send.
*/
func TestNBSend( t *testing.T ) {
	var count int = 0;

	ch := make( chan *ipc.Chmsg, 1 )				// make a channel which holds 1 message
	go add_nbmsg( ch, 1 )							// this write should be successful
	go add_nbmsg( ch, 2 )							// this write shouldn't make the pipe, but shouldn't block the go routine either

	time.Sleep( 4 * time.Second )					// pause to let go routines do their thing for sure

	for {
		select {
			case m := <-ch:								// read from channel; this should happen only once
				fmt.Fprintf( os.Stderr, "[INFO] read from channel: %d\n", m.Msg_type );
				count++
	
			default:									// this will happen when no data left to read
				if count != 1 {
					fmt.Fprintf( os.Stderr, "[FAIL] expected a read count of 1, got %d\n", count );
					t.Fail()
				} else {
					fmt.Fprintf( os.Stderr, "[OK]   expected a read count of 1, got 1\n" );
				}
				return
		}
	}
}
