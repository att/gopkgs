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


/*

	Mnemonic:	sheep_herder
	Abstract:	A simple go routine to cause the log file that sheep are writing to to be
				rolled now and again.

	Date:		30 April 2014
	Author:		E. Scott Daniels

	Mods:		16 Jul 2014 - Corrected bug that was causing log file name to be gnerated
					based on local time and not zulu time.
				02 Jul 2014 - Corrected typos in Baa message.
*/

package bleater

import (
	"fmt"
	"time"

	//"forge.research.att.com/gopkgs/bleater"
)

/*
	Generate a new logfile name based on the current date, the log directory passed in
	and the cycle period.

	The file name created will have the syntax:
		<log_dir>/<prefix>.log.<date>

	where <prefix> is taken from the bleater.

	If period is >86400 seconds, then the file name is day based.
	If period is <86400 but > 3600 seconds, then file is based on hour, else
	the file includes the minute and is of the general form yyyymmddhhmm.
*/
func (b *Bleater) Mk_logfile_nm( ldir *string, period int64 ) ( *string ) {
	ref := "200601021504"

	if period >= 86400 {
		ref = "20060102"
	
	} else {
		if period >= 3600 {
			ref = "2006010215"
		}
	}

	s := fmt.Sprintf( "%s/%s.log.%s", *ldir, b.pfx, time.Now().UTC().Format( ref ) )		// 20060102 is Go's reference date and allows us to specify how we want to see it
	return &s
}

/*
	Go routine that manages the rolling over of the bleater log.
	We assume there is already a log open.  Period defines when the log is rolled (e.g. 300 causes
	the log to be rolled on 5 minute boundaries, 3600 hour boundaries, and 86400 at midnight).
*/
func (b *Bleater) Sheep_herder(  ldir *string, period int64 ) {
	b.Baa( 1, "sheep herder started with a period of %ds, first log roll in %d seconds", period, period - (time.Now().Unix() % period ) )

	if period < 60 {			// dictate some sanity
		period = 60
	}

	for {
		now := time.Now().Unix()
		time.Sleep(  time.Duration( (period - (now % period )) ) * time.Second )	// should wake us at midnight
		lfn := b.Mk_logfile_nm( ldir, period )
		b.Baa( 0, "herder is rolling the log into: %s", *lfn )
		
		err := b.Append_target( *lfn, true )			// create the next directory
		if err != nil {
			b.Baa( 0, "ERR: unable to roll the log to %s: %s", lfn, err )
		} else {
			b.Baa( 0, "herder rolled the log" )
		}

		time.Sleep( 5 * time.Second )				// let some time pass before we recalc midnight so as to not loop during the same second
	}
}

