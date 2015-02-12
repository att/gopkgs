// vi: ts=4 sw=4:
/*
	Mnemonic:	ssh_broker
	Abstract: 	Provides an interface to the ssh package that allows a local script (shell or python)
				to be read and sent to a remote host for execution.  The standard error and output are
				returned. Ssh connections are established using one or more private key files that
				are supplied when the broker object is created, and the connections persist until 
				the object is closed allowing for subsequent commands to be executed without the 
				overhead of the ssh setup.  

				Scripts must have a #! line which is used to execute the interpreter on the remote
				host.  The script is then written on stdin to the interpreter.  If python is in the 
				#! line, then leading whitespace isn't stripped as the script is sent to the remote host.  
				Commented lines (beginning with #), and blank lines are removed (might cause issues
				for strings with embedded newlines, but for now I'm not worrying about those).

				When the broker object is crated, a single script initiator is started, however it is 
				possible for the user application to start additional initiators in order to execute
				scripts in parallel.  The initiators read from the queue of scripts to run and 
				create a session, send the script, and wait for the results which are returned to the 
				caller directly, or as a message structure on a channel, depending on which function 
				was used. 

				There also seems to be a limit on the max number of concurrent sessions that one SSH
				connection will support.  This seems to be a host policy, rather than a blanket SSH
				constant, so an initiator will retry the execution of a script when it appears that the
				failrue is related to this limit.  All other errors are returned to the caller for 
				evaluation and possbile retry. 

				There are two functions that the user can invoke to run a script on a remote host:
				Run_on_host() and NBRun_on_host().  The former will block until the command is run 
				and the latter will queue the request with the caller's channel and the results message
				will be written to the channel when the script execution has been completed. 

				Basic usage:	(error checking omitted)
						// supply the key filenames that are recognised on the remote side in authorised keys
						keys := []string { "/home/scooter/.ssh/id_rsa", "/home/scooter/.ssh/id_dsa" }
						broker := Mk_broker( "scooter", keys )			// create a broker for user, with keys
						host := "cheetah"
						script := "/user/bin/do_something"					// can be in PATH, or qualified
						parms := "-t 10 /tmp/output"						// command line parameters
						stdout, stderr, err := broker.Run_on_host( host, script, parms )

				The script may be Korn shell, bash, or python and is fed into the interpreter as standard
				in put so $0 (arg[0]) will not be set.  The broker will attempt to set the variable 
				ARGV0 to be the script name should the script need it. Other than this small difference, 
				and there not being any standard input, the script should function as written.  It is possible
				that other script types can be used, though it is known that #!/usr/bin/env awk will fail and
				thus "pure awk" must be wrapped inside of a ksh or bash script.


				There are also two functions which support the running af a command on the remote host in a
				"traditional" SSH fasion.  Run_cmd (blocking) and NBRun_cmd (non-blocking) run the command
				in a similar fashion as the script execution methods.

	Author:		E. Scott Daniels
	Date: 		23 December 2014

	Mods:		15 Jan 2015 - Added ability to send an environment file before the named script file.
				01 Feb 2015 - Corrected bug, rsync happening on session2, not new connection.
				12 Feb 2015 - Dropped the ability to ditch leading/trailing whitspace when sending to 
					standard input.

	CAUTION:	This package reqires go 1.3.3 or later.
*/

package ssh_broker

import (
	"bytes"
	"bufio"
    "fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

    "code.google.com/p/go.crypto/ssh"
)

// ------ private structures -----------------------------------------------------------------------------
/*
	Manages an ssh connection to a particular host:port.
*/
type connection struct {
	host		string				// host name:port connected to
	schan		*ssh.Client			// the ssh supplied connection
	retry_ch	chan *Broker_msg	// retry channel for the host
}

// ------ public structures -----------------------------------------------------------------------------
/*
	Manages connection by host name, and the configuration that is needed to 
	create a new session.  Struct returned by Mk_broker
*/
type Broker struct  {
	conns		map[string]*connection	// current connections
    conns_lock	sync.RWMutex			// mutex to gate update to map

	config		*ssh.ClientConfig		// configuration that must be given to ssh on connect
	init_ch		chan *Broker_msg		// channel the initiators listen on for requests
	retry_ch	chan *Broker_msg		// channel where we'll queue retries popped from a host retry queue
	ninitiators	int						// number of initiators started
	was_closed	bool					// set to true if Close called on us so we don't try to reuse
	rsync_src	*string					// space separated list of files to rsynch to the other side
	rsync_dir	*string					// target directory for rsync
	verbose		bool					// we might get chatty if it's true
}

/*
	Used to pass information into an initator and then back to the requestor. External users
	(non-package functions) can use the get functions related to this struct to extract 
	information (none of the information is directly available).
*/
type Broker_msg struct {
	host	string					// host:port
	cmd		string					// command to execute (not local script)
	sname	string					// script name (fully qualified, or in PATH)
	parms	string					// command line parms to pass to script
	env		string					// file where the script's environment lives (optional)
	id		int						// caller assigned id to make response tracking easier
	ntries	int						// number of times we've retried this request
	startt	int64					// start/end times for elapsed info
	endt	int64
	stdout	bytes.Buffer
	stderr	bytes.Buffer
	err		error					// any resulting error
	resp_ch chan *Broker_msg		// channel used to send back results
}

// --------------------------------------------------------------------------------------------------

/*
	Given a filename, test to see if it is fully or partially qualified (has a slant). If it is not
	then the PATH is searched for a matching file and that fully qualified filename is returned or
	error is set.  If the filename isn't qualified in any way, then the file name is returned. 
	Error is set if path is searched and no match is found.
*/
func find_file( fname string ) ( pname string, err error ) {
	pname = fname
	if fname[0:1] != "." && fname[0:1] != "/" {					// not absolute or relative name
		pname, err = exec.LookPath( fname )						// search the path
	}

	return
}

/*
	Read from src and write to dest.
*/
func send_file( src *bufio.Reader, dest io.WriteCloser ) {
    for {
        rec, rerr := src.ReadBytes( '\n' );
        if rerr == nil {
			if len( rec ) > 0  &&  rec[0] != '#' {				// skip empty lines and comment lines
				dest.Write( rec )
			}
		} else {
			return
		}
	}
}

/*
	Expected to be invoked as a gorotine which runs in parallel to sending the ssh command to the 
	far side. This function reads from the input buffer reader br and writes to the target stripping 
	blank and comment lines as it goes.
*/
func send_script( sess *ssh.Session, argv0 string, env_file string, br *bufio.Reader ) {

	target, err := sess.StdinPipe( )				// we create the pipe here so that we can close here
	if err != nil {
		fmt.Fprintf( os.Stderr, "unable to create stdin for session: %s\n", err )
		return
	}
	defer target.Close( )

	if argv0 != "" {
		target.Write( []byte( "ARGV0=\"" + argv0 + "\"\n" ) )			// $0 isn't valid using this, so simulate $0 with argv0
	}

	if env_file != "" {										// must push out the environment first
		env_file, err = find_file( env_file )				// find it in the path if not a qualified name
		if err == nil {
			ef, err := os.Open( env_file )

			if err != nil {
				fmt.Fprintf( os.Stderr, "ssh_broker: could not open environment file: %s: %s\n", env_file, err )
			} else {
				ebr := bufio.NewReader( ef );								// get a buffered reader for the file
				send_file( ebr, target )
				ef.Close()
			}
		} else {
			fmt.Fprintf( os.Stderr, "ssh_broker: could not find  environment file: %s: %s\n", env_file, err )
		}
	}

	send_file( br, target )
}

/*
	Run On A Remote.  Opens the named file and looks at the first line for #!. 
	Allocates stdin on the session and then runs the #! named command. If sname 
	is a relative or absolute path then it is opened directly. If it is not, then 
	PATH is searched for the script.  This function assumes that the session has 
	already been set up with stdout/err if needed. 

*/
func ( b *Broker ) roar( req *Broker_msg ) ( err error ) {
	if req == nil {
		err = fmt.Errorf( "no request block" )
		return
	}

	sess, err := b.session2( req.host )							// get a connection and session
	if err != nil {
		return
	}

	sess.Stdout = &req.stdout
	sess.Stderr = &req.stderr

	pname, err := find_file( req.sname )
	if err != nil {
		return
	}


	f, err := os.Open( pname )							// open script and read first line here to suss off shell
	if err != nil {
		return
	}
	defer f.Close()

	br := bufio.NewReader( f );								// get a buffered reader for the file
	rec, rerr := br.ReadBytes( '\n' );						// read first line
	if rec[0] == '#' && rec[1] == '!' && rerr == nil {
		rec = bytes.Trim( rec, "\n" )						// zap the newline

		shell := string( rec[2:] ) + " -s -- " + req.parms		// assume ksh or bash
		if strings.Index( shell, "python" ) > 0 {
			shell = string( rec[2:] ) + " - " +  req.parms		// shell to execute with python style command line for stdin
		}

		go send_script( sess, pname, req.env, br )			// write the remainder of the script in parallel
		err = sess.Run( shell )
	} else {
		err = fmt.Errorf( "not run: run on a remote requires script to have leading #! directive on the first line: %s\n", pname )
	}

	return
}

/*
	Run a command on a remote host.
*/
func ( b *Broker ) rcmd( req *Broker_msg ) ( err error ) {
	if req == nil {
		err = fmt.Errorf( "no request block" )
		return
	}

	sess, err := b.session2( req.host )							// get a connection and session
	if err != nil {
		return
	}

	sess.Stdout = &req.stdout
	sess.Stderr = &req.stderr

	err = sess.Run( req.cmd )

	return
}

/*
	Given a key file, open, and read it, then convert its contents into a "signer" that 
	is needed by ssh in the config auth list. 
*/
func read_key_file( kfname string ) ( s ssh.Signer, err error ) {
	s = nil

	kf, err := os.Open( kfname )
	if err != nil {
		return
	}
	defer kf.Close()

	buf := make( []byte, 4096 )			// we could check file length and base on that
	_, err = kf.Read( buf ) 
	if err != nil {
		return
	}

	s, err = ssh.ParsePrivateKey(  buf )		// convert to a "signer"
	if err != nil {
		s = nil
		return
	}

	return
}

/*
	Find or create our connection to the named host. If a connection doesn't
	exist, then we'll create one. If the rsync data is present, then we'll
	rsynch stuff over while we have the lock.
*/
func ( b *Broker ) connect2( host string ) ( c *connection, err error ) {
	err = nil

	if strings.Index( host, ":" ) < 0  {
		host = host + ":22"							// add default port if not supplied
	}

	b.conns_lock.RLock()								// get a read lock
	c = b.conns[host]
	b.conns_lock.RUnlock()
	if c != nil {										// we've already connected, just return
		return
	}

	if b.rsync_src != nil && b.rsync_dir != nil {		// no connection, rsynch if we need to
		toks := strings.Split( host, ":" )				// must split off port for rsynch
		b.synch_host( &toks[0] )
	}

	b.conns_lock.Lock( )								// get a write lock
	defer b.conns_lock.Unlock()							// hold until we return

	c = b.conns[host]
	if c != nil {										// someone created while we were waiting on lock
		return											// just send it back now
	}

	c = &connection{ host: host }
	c.retry_ch = make( chan *Broker_msg, 1024 )			// the host retry queue
	c.schan, err = ssh.Dial( "tcp", host, b.config )	// establish the tcp session (ssh channel)
	if err != nil {
		c = nil
		return
	}
	
	b.conns[host] = c									// finally, add to our map (host:port)
	return
}

/*
	Create a new sesson to the named host establishing the connection if
	we must. 
*/
func ( b *Broker ) session2( host string ) ( s *ssh.Session, err error ) {

	s = nil
	c, err := b.connect2( host )			// ensure we have a connection first
	if err != nil {
		return
	}

	s, err = c.schan.NewSession( )
	if err != nil {
		s = nil
	}

	return
}

/*
	An initiator runs as a goroutine and pulls requests from the initiator channel for 
	processing. The result is folded back into the request and written to the user channel 
	contained in the request. If a request fails with an error that contains the phrase
	"administraively prohibited", then it is retried as this is usually a bump against the
	number of sessions permitted to any single host.   The rerty logic is this:

		Queue the request on the specific host's (channel) retry queue. When the next 
		iterator processing a script on that host completes, it will check the retry
		queue for the host and move it to the main broker retry queue which will then 
		be picked up by an initiator.  If we were to move the request straight to the 
		broker's retry queue, it might be picked up before any of others had finished
		creating a tight loop that accomplishes nothing. 
*/
func ( b *Broker ) initiator( id int ) {
	var (
		is_open bool
		req		*Broker_msg
	)

	for {
		select { 							// read from main or retry channel and return if channel is closed
			case req, is_open = <- b.retry_ch: 
				if !is_open {
					return
				}

			case req, is_open = <- b.init_ch:
				if !is_open {
					return
				}

		}

		if req.cmd != "" {								// remote command to execute rather than a local script
			req.startt = time.Now().Unix()
			req.err = b.rcmd( req )						// run it 
			req.endt = time.Now().Unix()
		} else {
			req.startt = time.Now().Unix()
			req.err = b.roar( req )						// local script: send the request to the remote host
			req.endt = time.Now().Unix()
		}

		if req.err != nil {
			if req.ntries < 10  &&  strings.Contains( fmt.Sprintf( "%s", req.err ), "administratively prohibited" ) {	// likely over max sessions
				c, err := b.connect2( req.host )			// find the connection
				if err == nil { 							// no error finding it, then queue the request to be retried
					req.ntries++
					c.retry_ch <- req
					req = nil								// no more processing here
				}
			} 
		} 

		if req != nil {
			c, err := b.connect2( req.host )					// find the connection; look for retries that queued while we were running

			if  req.resp_ch != nil {
				req.resp_ch <- req								// return the request to the caller
			}

			if err == nil {
				if len( c.retry_ch ) > 0 {					// must pop and queue it onto the master retry channel
					r := <-c.retry_ch
					b.retry_ch <- r
				} 
			}
		}
	}

}


// ----- public msg functions ------------------------------------------------------------------------------
/*
	Returns the standard out, standard error elapsed time (sec) and the overall error state contained in 
	the message.
*/
func (m *Broker_msg) Get_results( ) ( stdout bytes.Buffer, stderr bytes.Buffer, elapsed int64, err error ) {
	return m.stdout, m.stderr, m.endt - m.startt, m.err
}

/*
	Returns the host, script name and ID contained in the 
*/
func (m *Broker_msg) Get_info( ) ( host string, sname string, id int ) {
	return m.host, m.sname, m.id
}


// ------------ public broker functions ---------------------------------------------------------------------

/*
	Create a broker for the given user and with the given key files.
*/
func Mk_broker( user string, keys []string ) ( broker *Broker ) {
	if len( keys ) <= 0 {
		broker = nil
		return
	}

	broker = &Broker { }
	broker.conns = make( map[string]*connection, 100 )		// value is a hint, not limit
	broker.was_closed = false

	auth_list := make( []ssh.AuthMethod, len( keys ) )

	j := 0
	for i := range keys {
		s, err := read_key_file( keys[i]  ) 
		if err == nil {										// error isn't fatal to the process, but not having the key later might cause issues
			auth_list[j] = ssh.PublicKeys( s )
			j++
		}
	}

	if j <= 0 {												// didn't find a suitable key
		fmt.Fprintf( os.Stderr, "mk_broker: no suitable key found\n" )
		broker = nil
		return
	}

	broker.config = &ssh.ClientConfig {						// set up the config info that ssh needs to open a connection
		User: user,
		Auth: auth_list,
		ClientVersion: "",
	}

	broker.init_ch = make( chan *Broker_msg, 2048 )
	broker.retry_ch = make( chan *Broker_msg, 2048 )
	go broker.initiator( 0 )									// by default single threaded
	broker.ninitiators = 1
	return
}

/*
	Start n initiators which allow n scripts to be executed in parallel.
*/
func ( b *Broker ) Start_initiators( n int ) {
	if n <= 0 {
		return
	}

	if b.was_closed {
		b.was_closed = false			// this reopens it
	}

	for i := 0; i < n; i++  {
		go b.initiator( i + b.ninitiators )
	}

	b.ninitiators += n
}

/*
	Clean up things that need to be closed when the user makes us goaway.
*/
func ( b *Broker ) Close( ) {
	if b == nil {
		return
	}

	for k, c := range b.conns {
		if c != nil {
			c.schan.Close()
			b.conns[k] = nil
		}
	}

	close( b.init_ch )					// close the initiator channel which should cause iniators to stop
	close( b.retry_ch )

	b.was_closed = true					// prevent using it unless more initiators are opened
	b.ninitiators = 0
}

/*
	Close a session to the named host.
*/
func ( b *Broker ) Close_session( name *string ) ( err error ) {
	err = nil

	if b == nil {
		err = fmt.Errorf( "close_session: broker pointer was nil" )
	}

	if b.was_closed {
		return
	}
	
	c := b.conns[*name]
	if c == nil {					// nothing to close
		return
	}

	err = c.schan.Close()
	b.conns[*name] = nil

	return
}


/*
	Execute the script (a local shell/python script) on the remote host. 
	This is done by creating a request and putting it on the initiator queue and 
	waiting for the response on the dedicated channel allocated here. 

	A nil error indicates success, otherwise there was an error setting up for or 
	executing the command.  If stderr is nil, the error was related to setup rather 
	than execution.

*/
func ( b *Broker ) Run_on_host( host string, script string, parms string, env_file string ) ( stdout *bytes.Buffer, stderr *bytes.Buffer, err error ) {
	if b == nil || b.was_closed {
		err = fmt.Errorf( "run_on_host: broker pointer was nil, or broker has been closed" )
		return
	}

	req := &Broker_msg {
		host: 	host,
		sname:	script,
		env:	env_file,
		parms:	parms,	
	}
	req.resp_ch = make( chan *Broker_msg )			// we'll listen on this channel for response
	defer close( req.resp_ch )

	stdout = nil
	stderr = nil

	b.init_ch <- req						// send request to initiator queue
	req = <- req.resp_ch					// wait on the response
	stdout = &req.stdout
	stderr = &req.stderr
	req.resp_ch = nil
	err = req.err

	return
}

/*
	Put the execution request onto the initiator queue, but do not block. The response
	is put onto the channel provided by the user. If the channel is nil, the request
	is still queued with the assumption that the caller does not want the results. 
*/
func ( b *Broker ) NBRun_on_host( host string, script string, parms string,  uid int, uch chan *Broker_msg  ) ( err error ) {
	if b == nil || b.was_closed {
		err = fmt.Errorf( "nbrun_on_host: broker pointer was nil, or broker has been closed" )
		return
	}

	req := &Broker_msg {
		host: 	host,
		sname:	script,
		parms:	parms,	
		id:		uid,
		resp_ch:	uch,
	}

	b.init_ch <- req						// send request to initiator queue

	return
}

/*
	Execute the command on the named host. The cmd string is expected to be the 
	complete command line.
*/
func ( b *Broker ) Run_cmd( host string, cmd string ) ( stdout *bytes.Buffer, stderr *bytes.Buffer, err error ) {
	if b == nil || b.was_closed {
		err = fmt.Errorf( "run_cmd: broker pointer was nil, or broker has been closed" )
		return
	}

	req := &Broker_msg {
		host: 	host,
		cmd:	cmd,
	}
	req.resp_ch = make( chan *Broker_msg )			// we'll listen on this channel for response
	defer close( req.resp_ch )

	stdout = nil
	stderr = nil

	b.init_ch <- req						// send request to initiator queue
	req = <- req.resp_ch					// wait on the response
	stdout = &req.stdout
	stderr = &req.stderr
	err = req.err
	req.resp_ch = nil

	return
}

/*
	Run a command on the remote host, non-blocking. The cmd string is expected to 
	be the complete command line. 
*/
func ( b *Broker ) NBRun_cmd( host string, cmd string,  uid int, uch chan *Broker_msg  ) ( err error ) {
	if b == nil || b.was_closed {
		err = fmt.Errorf( "nbrun_cmd: broker pointer was nil, or broker has been closed" )
		return
	}

	req := &Broker_msg {
		host: 	host,
		cmd:	cmd,
		sname:	"",
		parms:	"",
		id:		uid,
		resp_ch:	uch,
	}

	b.init_ch <- req						// send request to initiator queue

	return
}

/*
	Set/reset verbose.
*/
func ( b *Broker ) Set_verbose( value bool ) {
	b.verbose = value
}
