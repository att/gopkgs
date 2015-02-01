// vi: ts=4 sw=4:
/*
	Mnemonic:	rsync.go
	Abstract: 	This module contains the support to allow an rsync command to be run on the 
				localhost for each new remote host connection. The rsync connections are 
				run "outside" of the ssh environment that is managed by this package.

	Author:		E. Scott Daniels
	Date: 		20 January 2015

	Mods:		

	CAUTION:	This package reqires go 1.3.3 or later.
*/

package ssh_broker

import (
	//"bytes"
	//"bufio"
    "fmt"
	//"io"
	"os"
	//"os/exec"
	//"strings"
	//"sync"

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
*/
func ( b *Broker ) synch_host( host *string ) ( err error ) {

	if b == nil || b.rsync_src == nil || b.rsync_dir == nil  || host == nil {
		return
	}

	cmd := fmt.Sprintf( "rsync %s %s@%s:%s", *b.rsync_src, b.config.User, *host, *b.rsync_dir )
	
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
