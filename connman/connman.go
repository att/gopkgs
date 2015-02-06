// vi: ts=4 sw=4:

/*
 Mnemonic:	connman.go
 Abstract:	Implements a connection manager that listens for TCP connections, accepts them and then
			manages the sessions.

			Creates and manages a connetion environment that allows the user to establish TCP connections
			(outgoing) and listens/accepts connection requests from remote processes.

			The connection manager object is created with NewManager which accepts a listen port and a
			channel on which Sess_data objects are written.  The listener is invoked and when a connection is
			accepted a connection reader is started and a ST_NEW Sess_data object is written on the user's
			channel. The reader sends all data via the channel (ST_DATA) and if the session is disconnected
			a Sess_data object with the ST_DISC state is written and the connection is cleaned up (no need for user to
			invoke the Close function in a disconnect case.

			When a user establishes a connection the remote IP address and a session id (name) are supplied along
			with a communication channel. The channel may be the same channel supplied when the connection manager
			object was created or it may be a different channel. 
		
			Data received from either a UDP listener, or on a connected TCP session is bundled into a Sess_data
			struct and placed onto the appropriate channel.  The struct contains, in addition to the received
			buffer, the ID of the session that can be used on a generic Write command to, the current state of
			the session (ST_ constants), and a string indicating some useful (humanised) data about the session.

 Date:		15 November 2009
 Author: 	E. Scott Daniels

 Mods:		24 Jan 2014 - added doc and error checking.
			30 Apr 2014 - incorporated in gopkgs.
			06 May 2014 - Changed debug message.
			15 Aug 2014 - Added better doc.
			17 Aug 2014 - Changed error messages to write to stderr.
			26 Oct 2014 - Fixed bug in newdata (copy).
			03 Dec 2014 - Session struct now implements true Writer interface.
						  Added direct UDP writing via Writer interface.
			06 Jan 2014 - Ensure goroutine exits when session is lost.
*/

package connman

import (
	//"bytes"
	"fmt"
	//"errors"
	"net"
	"os"
)

const (
					// state describing the session data returned on the user's channel
	ST_NEW = iota	// new connection
	ST_DATA			// data received
	ST_DISC			// disconnected connection
	ST_ACCEPTED		// session has been accepted
)

const(						// connection states
	ST_CLOSING int = 1 		// close already in progress
)

const (
	LISTENER string = "l0" 		// our magic session id for the default litener
)

type Cmgr struct {						// main session manager class
	clist	map[string] *connection 	// tcp connections, also tracks udp listeners
	llist	map[string] net.Listener 	// tcp listeners
	lcount	int 						// tcp listener count for id string generation
	ucount	int 						// udp 'listener' count for id string
	mcount	int 						// multicast 'listener' count for id string
}

/*
	Data that is returned on the channel to the user.
*/
type Sess_data struct {
	Buf		[]byte		// actual data
	Id		string		// identify src for user that receives data from multiple connections
	From	string		// message source address
	State	int			// ST_ constants indicating the session state
	Data	string		// maybe useful (humanised) data about the session or message; generally empty for data.
	sender	*connection		// enables the data block to be used as a writer
}


type connection struct {			// track specifics for a single 'connection'
	id			string 				// our assigned id (hashed)
	conn		net.Conn 			// connection interface
	uconn		*net.UDPConn 		// UDP socket
	uaddr		*net.UDPAddr		// udp address 'bound' to this struct (fast writes)
	data2usr	chan *Sess_data 	// channel to send data from this conn to user
	bytes_in	int64
	bytes_out	int64
	state		int 				// current state
}

/* -------------- private ------------------------------------------------------- */

// listen and accept connections
func (this *Cmgr) listener(  l net.Listener, data2usr chan *Sess_data ) {
	var n 	int = 0
	
	for {
		conn, err := l.Accept( )
		if err == nil {
			n += 1
			conn_data := new( connection )
			conn_data.id = fmt.Sprintf( "a%d", n )
			conn_data.conn = conn
			conn_data.data2usr = data2usr
			this.clist[conn_data.id] = conn_data 		// hash for write to session

			sdp := new( Sess_data ) 			// create and format accept msg back to user
			sdp.Id = conn_data.id
			sdp.From = conn.RemoteAddr().String()
			sdp.State = ST_ACCEPTED
			sdp.Data = fmt.Sprintf( "connection [%s] accepted from: %s", conn_data.id, sdp.From )
			data2usr <- sdp

			go this.conn_reader( conn_data )
		} else {
			return
		}
	}
}

/*
	Create a new data object.
*/
func newdata( buf []byte, id string, state int, sender *connection, from *net.UDPAddr, data string ) (* Sess_data) {
	sdp := new( Sess_data )
	sdp.Buf = make( []byte, len( buf ) )
	sdp.sender = sender
	copy( sdp.Buf, buf )
	sdp.Id = id
	if from != nil {
		sdp.From = from.String()
	}

	sdp.State = state
	sdp.Data = data

	return sdp
}

// Read from sesion and write on the user's channel. The session can be either TCP or
// UDP even though there is no continuous UDP connection 'session' implies the listen
// port that was esabalished.
// On error (disc) we send the user a nil buffer on the channel and close things up
func (this *Cmgr) conn_reader( cp *connection )   {
	var buf []byte

	buf = make( []byte, 2048 )

	if cp.conn != nil {
		cp.data2usr <- newdata( nil, cp.id, ST_NEW, nil, nil, fmt.Sprintf( "%s", cp.conn.RemoteAddr()) )   // indicate new session
	}

	for {
		var nread 	int
		var err		error
		var from	*net.UDPAddr = nil 	// packet source if udp

		if cp.conn != nil {							// nil if this is udp, or if the session isn't connected
			nread, err = cp.conn.Read( buf );	
		} else {
			if cp.uconn != nil {
				nread, from, err = cp.uconn.ReadFromUDP( buf );	
			} else {
				return				// no session just stop the reader
			}
		}

/*
		--- in original go this logic was needed, though it seems that with the evolution it's not any longer ----
		if err != nil {
			//if e2common( err ) !=  os.EAGAIN {	// we can ignore this if we assume all errors mean close
			if e, ok := err.(os.PathError); ok && e.Error != os.EAGAIN {		//e is null if err isn't os.Error
				cp.data2usr <- newdata( nil, cp.id, ST_DISC, nil, "" ) 	// disco to the user programme	
				cp.data2usr = nil
				this.Close( cp.id ) 		// drop our side and stop
				return
			}
		} else {
			if cp.data2usr != nil {
				cp.bytes_in += int64( nread )
				cp.data2usr <- newdata( buf[0:nread], cp.id, ST_DATA, cp, from, "" )
			}
		}
*/
		if err != nil {					// assume that eagain has been implemented out
			cp.data2usr <- newdata( nil, cp.id, ST_DISC, nil, nil, "" ) 	// disco to the user programme	
			cp.data2usr = nil
			this.Close( cp.id ) 		// drop our side and stop
			return
		}

		if cp.data2usr != nil {							// a nil buffer signals end to caller, so only write if not nil
			cp.bytes_in += int64( nread )
			cp.data2usr <- newdata( buf[0:nread], cp.id, ST_DATA, cp, from, "" )
			buf = make( []byte, 2048 )					// new buffer to prevent overruns
		}
	}
}

/* ------ public ---------------------------------------------------- */

/*
	Starts a TCP listener, allowing the caller to supply type (tcp, tcp4, tcp6) and interface (0.0.0.0 for any)
	then opens and binds to the socket. A goroutine is started to actually do the listening and will
	accept sessions that connect. Generally the listen method will be driven during the construction of
	a cmgr object, though I user can use this if more than one listen port is required.

	Returns an ID which identifies the listener, and a boolean set to true if the listener was established
	successfully.
*/
func (this *Cmgr) Listen( kind string, port string,  iface string, data2usr chan *Sess_data ) ( lid string, err error ) {
	if port == ""  || port == "0" {		// user probably called constructor not wanting a listener
		return "", nil
	}


	lid = ""
	l, err := net.Listen( kind, fmt.Sprintf( "%s:%s", iface, port ) )
	if err != nil {
		err = fmt.Errorf( "unable to create a listener on port: %s; %s", port, err )
		return
	}

	lid = fmt.Sprintf( "l%d", this.lcount )
	this.lcount += 1

	this.llist[lid] = l
	go this.listener(  this.llist[lid], data2usr )
	return
}

/*
	Starts a UDP listener which will forward received data back to the application using
	the supplied channel.  The listener ID is returned along with a boolian indication
	of success (true) or failure.
*/
func (this *Cmgr) Listen_udp( port int, data2usr chan *Sess_data ) ( uid string, err error) {
	var addr	net.UDPAddr

	uid = ""
	addr.IP = net.IPv4( 0, 0, 0, 0 )
	addr.Port = port
	uconn, err := net.ListenUDP( "udp",  &addr )
	if err != nil {
		err = fmt.Errorf( "unable to create a udp listener on port: %d; %s", port, err )
		return
	}

	uid = fmt.Sprintf( "u%d", this.ucount ) 	// successful bind to port
	this.ucount += 1

	cp := new( connection )
	cp.conn = nil
	cp.uconn = uconn
	cp.data2usr = data2usr 		// session data written to the channel
	cp.id = uid 			// user assigned session id
	this.clist[uid] = cp 		// hash for write to session
	go this.conn_reader( cp ) 	// start reader; will discard if data2usr is nil
	
	this.clist[uid] = cp
	return
}

/*
	Joins a multicast group as a listener on the named interface
*/
func (this *Cmgr) Listen_mc( ifname string, addr string, data2usr chan *Sess_data ) ( sessid string, err error ) {

	sessid = "";
	iface, err := net.InterfaceByName( ifname )
	if err != nil {
		return
	}

	sessid = fmt.Sprintf( "m%d", this.mcount ) 	// successful bind to port
	this.mcount++

	uaddr, err := net.ResolveUDPAddr( "udp", addr )
	if err != nil {
		return
	}

	uconn, err := net.ListenMulticastUDP( "udp", iface, uaddr )
	if err != nil {
		return
	}

	cp := new( connection )
	cp.conn = nil
	cp.uconn = uconn
	cp.data2usr = data2usr 		// session data written to this channel
	cp.id = sessid 				// user assigned session id
	this.clist[sessid] = cp 	// hash for write to session

	go this.conn_reader( cp ) 	// start reader; will discard if data2usr is nil
	
	return sessid, err
}

	

/*
	Provides a more consistant interface with the Listen_udp name convention and is just
	a wrapper for Listen().
*/
func (this *Cmgr) Listen_tcp( port string, data2usr chan *Sess_data ) ( string, error ) {
	return this.Listen( "tcp", port, "0.0.0.0", data2usr )
}

/*
	Lists the current statistics about connections to the standard output device.
*/
func (this *Cmgr) List_stats(  ) {
	var ucount int = 0 		// count of udp 'listeners' to dec conn count by

	fmt.Fprintf( os.Stderr, "%d tcp listeners:\n", len( this.llist ) ) 		// tcp listeners
	for l := range this.llist {
		fmt.Printf( "\t%s on %s\n", l, this.llist[l].Addr().String()  )
	}

	for cname := range this.clist {				// udp listeners
		cp := this.clist[cname]
		if cp.uconn != nil {
			ucount += 1
			fmt.Printf( "\t%s UDP on %s  %5d %5d\n", cp.id,  cp.uconn.LocalAddr().String(), cp.bytes_in, cp.bytes_out )
		}
	}

	fmt.Printf( "%d tcp connections:\n", len( this.clist ) - ucount ) 		// established tcp connections
	for cname := range this.clist {
		cp := this.clist[cname]
		if cp.conn != nil {
			fmt.Printf( "\t%s -> %s %5d %5d\n", cp.id, cp.conn.RemoteAddr(), cp.bytes_in, cp.bytes_out )
		}
	}

}

/*
	Establishes a connection to the target process (ip:port) and starts a reader listening
	for data on the session.  Any received data will be forwarded to the user application
	via the channel provided.
*/
func (this *Cmgr) Connect( target string, uid string, data2usr chan *Sess_data ) ( err error ){
	err = nil;
	if this == nil {
		err = fmt.Errorf( "cannot connect; nil object passed in" );
		return
	}

	cp := new( connection )
	cp.conn, err = net.Dial( "tcp", target )
	//if( err != nil ) {
		//fmt.Printf( "session_connect: unable to create session to: %s: %s\n", target, err )
	//} else {
	if err == nil {
		cp.data2usr = data2usr 		// session data written to the channel
		cp.id = uid 				// user assigned session id
	}

	this.clist[uid] = cp 		// hash for write by id to session
	go this.conn_reader( cp ) 	// start reader; will discard if data2usr is nil

	return
}

// ----------- generic write functions allowing writes to a named session ---------------------------------------

/*
	Writes the byte array to the named connection.
*/
func (this *Cmgr) Write( id string, buf []byte ) ( err error ) {
	var (
		n	int
		nw	int
	)

	err = nil

	if cp, ok := this.clist[id]; ok {
		cp.bytes_out += int64( len( buf ) )

		for n = len( buf ) ; n >0 ; {
			nw, err = cp.conn.Write( buf ) 	// ignore error assuming that reader will catch and close things up
			n -= nw;
			if err != nil {
				return
			}
		}
	}

	return
}

/*
	Writes n bytes from the byte array to the named session.
*/
func (this *Cmgr) Write_n( id string, buf []byte, n int ) ( err error ){
	var (
		nw	int
	)

	err = nil

	if cp, ok := this.clist[id]; ok {
		cp.bytes_out += int64( len( buf ) )

		for  ; n >0 ; {
			nw, err = cp.conn.Write( buf ) 	// ignore error assuming that reader will catch and close things up
			n -= nw;
			if err != nil {
				return
			}
		}
	} 

	return
}

/*
	Writes the string to the named session.
*/
func (this *Cmgr) Write_str( id string, buf string ) ( err error ) {
	return this.Write( id,  []byte( buf ) )
}

/*
	Writes the byte array to the udp address given in 'to'.  The address is expected to be host:port format.
*/
func (this *Cmgr) Write_udp( id string, to string, buf []byte ) ( err error ) {
	err = nil

	if cp, ok := this.clist[id]; ok {
		addr, e := net.ResolveUDPAddr( "ip", to ) 		// parm1 is either ip, ip4 or ip6
		if e != nil {
			fmt.Fprintf( os.Stderr, "unable to convert address: %s\n", to )
			err = e
			return
		}

		cp.bytes_out += int64( len( buf ) )
		_, err = cp.uconn.WriteToUDP( buf, addr ) 	// ignore error assuming that reader will catch and close things up
	}

	return
}

// -------------------------- writes to a specific connection (faster than named writes) --------------------------------------------------------

/*
	Given a connection name, or UDP listener name, return the struct that can be used as the direct
	writer.
*/
func ( c *Cmgr ) Get_writer( id string ) ( *connection ) {
	return  c.clist[id]
}

/*
	Return a writer that can be used to write directly to the address. Faster than the generic
	id oriented writes because there is no need for id lookup to find the writer, and the address
	is already constructed. 
*/
func ( c *Cmgr ) Get_udp_writer( id string, addr string ) ( newcp *connection, err error ) {
	cp := c.clist[id]
	if cp == nil {
		err = fmt.Errorf( "get_udp_write: cannot find named session to generate writer from: %s", id )
		return nil, err
	}

	newcp = &connection {
		id: "unmapped",				// this is a standalone connection that isn't hash reachable
		uconn: cp.uconn,			// it references the same UDP connection struct
		data2usr: nil,				// this struct isn't used to receive data
		bytes_out: 0,
	}

	newcp.uaddr, err = c.String2udp_addr ( addr )			// convert user string for use by write()

	return newcp, err
}

/*
	These functions allow a 'direct' reply based on a message that was written to the user's channel
*/

/*
	Writes the contents of buf (bytes) to the process that sent the data represented by Sess_data.
	This implements the writer interface so things like fmt.Fprintf( ) can use the struct.

	Returns actual number written and error if underlying environment had issues.
*/
func (this *connection) Write( buf []byte ) ( nw int, err error ) {
	var (
		n	int				// number to write
		tpnw int				// number written this pass
	)

	err = nil

	
	for n = len( buf ); n > 0; {
		if this.conn != nil {
			tpnw, err = this.conn.Write( buf ) 			// connection oriented
		} else {
			if this.uaddr != nil {
				tpnw, err = this.uconn.WriteToUDP( buf, this.uaddr )		// udp oriented
			} else {
				nw = 0
				err = fmt.Errorf( "no address associated with the connection structure for a UDP write" )
				return
			}
		}

		this.bytes_out += int64( tpnw )
		n -= tpnw;
		nw += tpnw;
		if err != nil {
			return
		}
	}

	return
}

/*
	Allow udp from address to be captured for fast replies -- users should avoid doing
	this as all writes will go to the bound address rather than to the sender if 
	sess_data.Write() is used after this.
*/
func (s *Sess_data) Bind2sender( ) ( err error ) {
	err = nil
	if s.sender.uconn != nil { 
		s.sender.uaddr, err = net.ResolveUDPAddr( "udp", s.From )
	} else {
		return fmt.Errorf( "cannot bind to sender: no uconn in connection" )
	}

	return
}

/*
	Allows a previously bound sender to be dropped.
*/
func (s *Sess_data) Unbind_sender( ) {
	s.sender.uaddr = nil
}

/*
	Allows session data to be used as a Writer interface for connection oriented sessions.
*/
func (s *Sess_data) Write( buf []byte ) ( nw int, err error ) {
	return s.sender.Write( buf )
}

/*
	Writes n bytes to the process that sent the data represented by Sess_data
*/
func (this *Sess_data ) Write_n( buf []byte, n int ) ( err error ) {
	var (
		nw int
	)
	
	err = nil
	if this == nil {
		return
	}
	if this.sender == nil {
		err = fmt.Errorf( "sender not associated with session" )
		return
	}
	if this.sender.conn == nil {
		err = fmt.Errorf( "connection not associated with session" )
		return
	}

	this.sender.bytes_out += int64( n )
	for ; n > 0; {
		nw, err = this.sender.conn.Write( buf[0:n] ) 	// ignore error assuming that reader will catch and close things up

		n -= nw
		if err != nil {
			return
		}
	}

	return
}

/*
	Writes the string to the process that sent the data represented by Sess_data
*/
func (this *Sess_data ) Write_str( buf string ) ( n int, err error ) {
	return this.Write( []byte( buf ) )
}

/*
    Writes the buffer to the address assoiated with id.
*/
func (this *Cmgr) Write_udp_addr( id string, addr *net.UDPAddr, buf []byte ) {
	if cp, ok := this.clist[id]; ok {
		cp.bytes_out += int64( len( buf ) );
		cp.uconn.WriteToUDP( buf, addr );
	}
}

/*
   Build an address from the IP addres:port string passed in; for use with Write_udp_addr().
*/
func (this *Cmgr) String2udp_addr ( astr string ) ( addr *net.UDPAddr, err error ) {
	addr, err = net.ResolveUDPAddr( "udp", astr )
	return
}



// -------------------------------------------------------------------------------------------------------------------
func ( cm *Cmgr ) Get_conn( id string ) ( conn *connection ) {
	conn = cm.clist[id]
	return
}

/*
	Might be a faster way to send udp than using the Cmgr interface
*/
func ( conn *connection ) Direct_udp_send( addr *net.UDPAddr, buf []byte ) {
		conn.uconn.WriteToUDP( buf, addr );
}
// -------------------------------------------------------------------------------------------------------------------


/*
	Closes the named connection.
*/
func (this *Cmgr) Close( id string ) {
	
	sess, ok := this.clist[id] 		// map id to the session data
	if ok {
		if( sess.state != ST_CLOSING ) { // if close called, read will call us when it popps; in case we are preempted
			sess.state = ST_CLOSING
			_ = sess.conn.Close( )
			sess.conn = nil
			delete( this.clist, id )
		}

		return
	}

	ls, ok := this.llist[id] 			// listener
	if ok {
		ls.Close( )
		delete( this.llist, id )
	}
}

/*
	Creates a new connection manager object. Normally the establishment of a TCP listener is a two step process (create
	the manager, and then allocate a TCP listener), however this can be reduced to a single call if the port is greater
	than zero. In this case a TCP listener will be started on the port and "attached" to the manager with any data
	received by sessions connected to the port sent to the user application using the channel provided.  If the listener
	cannot be established a nil object is returned.  To create a UDP listener, the two step process is needed and this
	function is invoked with a port of zero and a nill data2user pointer.
	
*/
func NewManager( port string, data2usr chan *Sess_data  ) ( *Cmgr ) {
	this := new( Cmgr )
	this.clist = make( map[string] *connection ) 	// must allocate the maps first
	this.llist = make( map[string] net.Listener );	
	this.lcount = 0

	
	_, err := this.Listen_tcp( port, data2usr ) 		// if port is "" or "0", Listen will ignore and return ok
	if  err != nil  {
		fmt.Fprintf( os.Stderr, "unable to initialise session manager: %s\n", err )
		return nil
	}
	
	return this
}

