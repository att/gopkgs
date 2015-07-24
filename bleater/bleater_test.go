//vi: sw=4 ts=4:
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


package bleater_test

import (
	"testing"
	"os"
	"fmt"

	"codecloud.web.att.com/gopkgs/bleater"
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

	big_sheep.Add_child( little_sheep );						// big sheep gets baby


	little_sheep.Baa( 2, "%s", "should not show" );				// first tests from little sheep
	little_sheep.Baa( 1, "%s", "should  show (1)" );

	big_sheep.Set_level( 2 );									// up the master level
	little_sheep.Baa( 2, "%s", "should  show (2)" );				// little sheeps level2 bleats should be heard
	
	black_sheep.Baa( 2, "%s", "should  NOT show" );				// but not those from blackie

	big_sheep.Inc_level( );
	little_sheep.Baa( 2, "%s", "should  show after inc (2)" );	
	little_sheep.Baa( 3, "%s", "should  show after inc (3)" );	
	big_sheep.Dec_level( );
	little_sheep.Baa( 3, "%s", "should  NOT show after dec (3)" );

	big_sheep.Set_level( 0 );									// off the master volume
	little_sheep.Baa( 2, "%s", "should  NOT show after mster off (2)" );	
	little_sheep.Inc_level( );									// little should now be two and this should show
	little_sheep.Baa( 2, "%s", "should  show after little sheep inc (2)" );	


	little_sheep.Set_tsformat( "15:04" );
	little_sheep.Baa( 2, "%s", "should  show with different time format" );				// little sheeps level2 bleats should be heard

	// --- test rolling of the log ------
	fname := "/tmp/bleater_test.log"
	big_sheep.Baa( 0, "switching log file to %s; all messages should go there now", fname )

//	f, err := os.Create( fname )
	//err := big_sheep.Create_target( fname, true )
	err := big_sheep.Append_target( fname, true )
	if err != nil {
		fmt.Fprintf( os.Stderr, "failed to open log file for roll: %s: %s\n", fname, err )
		t.Fail( )
		os.Exit( 1 )
	}

	//big_sheep.Set_target( f, true )			// close old and push in our new
	
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
	black_sheep.Baa( 1, "should not appear [FAIL]" )
	for i := 0; i < 15; i++ {				// these should NOT be written
		black_sheep.Baa_some( "foo", 15, 1, "after reset: foo baa_some message 1:15 %d should NOT apear!   [FAIL]", i  )
		black_sheep.Baa_some( "bar", 5, 1, "after reset: bar baa_some message 1:5 %d  should NOT appear!   [FAIL]", i  )
	}
	black_sheep.Baa( 0, "end suppression test (no lines should say 'fail' above." )
	
	black_sheep.Set_level( 1 )					// reset level, these should both appear!
	black_sheep.Baa_some( "foo", 15, 1, "after reset: foo baa_some message 1:15  (after level reset)"  )
	black_sheep.Baa_some( "bar", 5, 1, "after reset: bar baa_some message 1:5  (after level reset)"  )
	black_sheep.Baa( 0, "two lines should have been written between the end suppression message and this" )
}


