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
	Mnemonic:	rsync.go
	Abstract: 	This module contains the support to allow an rsync command to be run on the
				localhost for each new remote host connection. The rsync connections are
				run "outside" of the ssh environment that is managed by this package.

	Author:		E. Scott Daniels
	Date: 		20 January 2015

	Mods:		13 Apr 2015 - Added explicit ssh command for rsync to use.

	CAUTION:	This package reqires go 1.3.3 or later.
*/

package ssh_broker

import (
    "fmt"
	"os"

	"codecloud.web.att.com/gopkgs/extcmd"
)

/*
	Accepts the setup for rsync commands.  We assume src is one or more
	source files and dest_dir is the name of the directory on the remote
	host.  When the rsync command is actually run, the two strings are
	combined (unchanged) with user@host: inserted just before the dest
	directory.
*/
func ( b *Broker ) Add_rsync( src *string, dest_dir *string ) {
	if b != nil {
		b.rsync_src = src

		b.rsync_dir = dest_dir
	}
}

func ( b *Broker ) Rm_rsync( ) {
	if b != nil {
		b.rsync_src = nil
		b.rsync_dir = nil
	}
}


// ---- private functions -----------------------------------------------------------------

/*
	Actually run the rsync command.  If verbose is true, then output is written
	to stderr, otherwise it is ignored. Error is returned and will be nil if
	the command successfully executed.

	The -e option is used to force rsync to use ssh with a nokey checking option
	so that we don't get caught by the reinstall of a machine which changes its
	key and would possibly block our rsync attempt.
*/
func ( b *Broker ) synch_host( host *string ) ( err error ) {

	if b == nil || b.rsync_src == nil || b.rsync_dir == nil  || host == nil {
		return
	}

	cmd := fmt.Sprintf( `rsync -e "ssh -o StrictHostKeyChecking=no" %s %s@%s:%s`, *b.rsync_src, b.config.User, *host, *b.rsync_dir )
	
	verbose := b.verbose
	if verbose {
		fmt.Fprintf( os.Stderr, "synch: %s\n", cmd )	
	}

	stdout, stderr, err := extcmd.Cmd2strings( cmd )
	if err != nil {
		fmt.Fprintf( os.Stderr, "ssh-broker: unable to rsync for host %s: %s\n", *host, err )
		fmt.Fprintf( os.Stderr, "ssh-broker: command: %s\n", cmd )
		verbose = true
	}
	if verbose {
		for i := range stdout {
			fmt.Fprintf( os.Stderr, "%s\n", stdout[i] )
		}

		for i := range stderr {
			fmt.Fprintf( os.Stderr, "%s\n", stderr[i] )
		}
	}

	return
}
