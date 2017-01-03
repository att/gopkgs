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
	This test is a bit different than most 'go test' function sets as the 
	test verifies that the package can be used, but output from the test
	is vetted by an external script which looks for various things, or the
	absense of things, in the log files that are created. Specifically, 
	the strings looked for in the output are:
		'.*should show.*'			(these are counted)
		'.*should NOT show.*'		(these cause an error if seen)
		'.*baa_some.*:.*'			(last field expected to be a multiple of m in n:m)
		'.*different time format.*'	(time stamp only, no date expected)

	as well as the labels big, little and black.  For the alternate time
	test, the timestamp is expected to present the time first (the date
	is not checked so omitting it, or presenting it second is fine).

	The vetting script also expects two files in /tmp to be created:
		bleat_test.log (containing messages from both little and big sheep)
		bleat_ls_test.log (containing only little sheep messages)

	It is possible for the 'go test' to pass, but for the vetting script to 
	be unhappy with the output and fail the overall test.  The vetting script
	should eliminate the human effort with respect to automating the test.
*/

package bleater_test

import (
	"testing"
	"os"
	"fmt"

	"github.com/att/gopkgs/bleater"
)

/*
	create three sheep that bleat.  Big sheep acts as the parent and thus
	the 'master' volume of all children can be controlled by setting the
	level on the big sheep.  The little sheep is given to the big sheep
	as a child, but black-sheep is left on its own. Thus, when the level
	of big-sheep is changed to 2, a bleat by little sheep with a setting
	of 2 should be heard, but when black sheep bleats at 2 it should still
	be silent as it's level has not been increased and it has no parent.
*/
func TestBleater( t *testing.T ) {
	big_sheep := bleater.Mk_bleater( 1, os.Stderr );			// create the sheep
	little_sheep := bleater.Mk_bleater( 1, os.Stderr );
	black_sheep := bleater.Mk_bleater( 1, os.Stderr );			// black sheep not added as child, so it should not get trickle down things

	big_sheep.Set_prefix( "big" );								// give them prefixes so we can see difference
	little_sheep.Set_prefix( "little" );
	black_sheep.Set_prefix( "black" );

	big_sheep.Add_child( little_sheep );						// big sheep gets baby


	little_sheep.Baa( 2, "%s", "should NOT show" );				// first tests from little sheep
	little_sheep.Baa( 1, "%s", "should show (1)" );

	big_sheep.Set_level( 2 );									// up the master level
	little_sheep.Baa( 2, "%s", "should show (2)" );			// now little sheep's level 2 bleats should be heard
	
	black_sheep.Baa( 2, "%s", "should NOT show" );				// but not those from blackie

	big_sheep.Inc_level( );
	little_sheep.Baa( 2, "%s", "should show after inc (2)" );	// both level 2 and 3 should continue to show
	little_sheep.Baa( 3, "%s", "should show after inc (3)" );	
	big_sheep.Dec_level( );
	little_sheep.Baa( 3, "%s", "should NOT show after dec (3)" );	// but not after the level was dec'd (down to 2)
	little_sheep.Baa( 2, "%s", "should show after inc (2)" );		// but level 2 should continue to show

	big_sheep.Set_level( 0 );												// off the master volume
	little_sheep.Baa( 2, "%s", "should NOT show after mster off (1)" );	 	// with master off and little set to 1 this should not show
	little_sheep.Inc_level( );												// little should now be two and next should show
	little_sheep.Baa( 2, "%s", "should show after little sheep inc (2)" );	


	little_sheep.Set_tsformat( "15:04" );
	little_sheep.Baa( 1, "%s", "should show with different time format" );				// little sheeps level2 bleats should be heard

	// --- test rolling of the log; big and little messages should cease to appear in stderr  ------
	fname := "/tmp/bleater_test.log"
	big_sheep.Baa( 0, "switching log file to %s; all big and little sheep messages should go there now; black messages should continue to stderr", fname )

	err := big_sheep.Append_target( fname, true )				// will push to the child so both big and little sheep should write to file now
	if err != nil {
		fmt.Fprintf( os.Stderr, "failed to open log file for roll: %s: %s\n", fname, err )
		t.Fail( )
		os.Exit( 1 )
	}

	big_sheep.Baa( 0, "big-sheep message should appear in new log file" )
	little_sheep.Baa( 0, "little sheep message should appear in log file" )
	black_sheep.Baa( 0, "black sheep should still be writing to stderr" )


	black_sheep.Baa( 0, "Testing baa_some now" )
	for i := 0; i < 50; i++ {
		black_sheep.Baa_some( "foo", 15, 1, "foo baa_some message 1:15 %d", i  )
		black_sheep.Baa_some( "bar", 5, 1, "bar baa_some message 1:5 %d", i  )
	}

	black_sheep.Baa_some_reset( "foo" )		// foo should write straight away, but not bar
	for i := 0; i < 50; i++ {
		black_sheep.Baa_some( "foo", 15, 1, "after reset: foo baa_some message 1:15 %d", i  )
		black_sheep.Baa_some( "bar", 5, 1, "after reset: bar baa_some message 1:5 %d", i  )
	}
	
	black_sheep.Baa( 0, "testing baa some reset" )
	black_sheep.Baa_some_reset( "foo" )
	black_sheep.Baa_some_reset( "bar" )
	black_sheep.Set_level( 0 )
	black_sheep.Baa( 1, "should NOT show [FAIL]" )
	for i := 0; i < 15; i++ {				// these should NOT be written
		black_sheep.Baa_some( "foo", 15, 1, "after reset: foo baa_some message 1:15 %d should NOT apear!   [FAIL]", i  )
		black_sheep.Baa_some( "bar", 5, 1, "after reset: bar baa_some message 1:5 %d  should NOT appear!   [FAIL]", i  )
	}
	black_sheep.Baa( 0, "end suppression test (no lines should say 'fail' above." )
	
	black_sheep.Set_level( 1 )					// reset level, these should both appear!
	black_sheep.Baa_some( "foo", 15, 1, "after reset: foo baa_some message (after level reset) 1:15 0"  )
	black_sheep.Baa_some( "bar", 5, 1, "after reset: bar baa_some message (after level reset) 1:5 0"  )
	black_sheep.Baa( 0, "two lines should have been written between the end suppression message and this" )

	fname = "/tmp/bleater_ls_test.log"
	little_sheep.Create_target( fname, true )		// direct little sheep off to it's own log now
	black_sheep.Baa( 0, "should not go into little sheep log file" )
	big_sheep.Baa( 0, "should not go into little sheep log file" )
	little_sheep.Baa( 0, "should go into little sheep log file" )
}


