// vi: sw=4 ts=4:
/*
	Mnemonic:	rabbit_hole
	Abstract:	See below.

	Date:		30 March 2016
	Author:		E. Scott Daniels

	Mods:		19 Nov 2016 - Adde reonnect logic when session for read or write is lost.

	Useful links:
			https://godoc.org/github.com/streadway/amqp#example-Channel-Consume
*/

/*
	Rabbit_hole creates a simple interface to a RabbitMQ exchange
	allowing setup via one function call to Mk_mqreader() for reading
	and Mk_mqwriter() for sending. Rabbit_hole also provides for a 
	channel listening interface which can be paused by the user.

	The user programme can create a listener via Mk_mqreader() and 
	then can either listen directly on the lister.Port for 
	amqp.Delivery messages, or can invoke listener.Eat() and
	supply a channel where Eat() will write received messages
	(alowing a central user function to process all messages from 
	multiple listeners).

	User programme creates a sender via Mk_mqwriter() which 
	returns a struct that is used to start the driver. Once
	the driver is started, the user can pass messages on the 
	struct.Port and the driver will push it out to the message
	exchange.
*/
package rabbit_hole

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

const (
	DURABLE		bool = true				// bloody amqp parms are true false w/o constants to make code readable
	AUTO_DEL	bool = true
	INTERNAL	bool = true
	WAIT		bool = false
	EXCLUSIVE	bool = true
	AUTO_ACK	bool = true
	LOCAL		bool = true
	MANDITORY	bool = true
	IMMED		bool = true
)

/* 
	Manages a reader connection.
*/
type Mq_reader struct {
	conn *amqp.Connection
	ch	*amqp.Channel
	qu 	*amqp.Queue					// random (generic) queue
	ex	string						// exchange name
	qname string					// name or "" for random queue
	Port <-chan amqp.Delivery		// this is exposed so that the user can listen directly
	url		string					// url used for connect
	key		string					// key we are listing with
	etype	string					// exhange type

	durable	bool					// exchange type flags
	auto_del bool
	internal bool

	qautodel bool					// specifc queue types
	qdurable bool
	qexclusive bool

	quiet	bool					// prevents diagnostics when true
	stop	bool					// stop flag
	paused	bool					// paused flag
}

/*
	Struct which manages a connection to an exchange for writing.
*/
type Mq_writer struct {
	conn *amqp.Connection
	ch	*amqp.Channel
	ex	string						// exchange name
	key	string						// default key used to send messages
	Port	chan interface{}		// anything user writes to port is sent to the exhange (accepts string, *string and []byte blobs)
	url		string					// url created on first connect call; used on reconnect attempts
	etype	string					// exhange type
	durable	bool					// exchange type flags
	auto_del bool
	internal bool
	stop	bool					// flag that the write goroutne should halt
	quiet	bool					// if true, diagnostic messages are supressed
	alive	bool					// true if the write connection is thought to b alive and messages can be sent
	Notify chan string				// user can send request to us (quiet, stop, etc.)
}

/*
	User can send this on the channel to use a specific key.
*/
type Mq_msg struct {
	Key	string
	Data	[]byte
}


/*
	Create a connection and then open a channel using the url that exists in the writer. 
	The caller must close both before exit.
*/
func (wrtr *Mq_writer) connect2url( ) ( err error ) {
	if wrtr == nil  || wrtr.url == "" {
		return fmt.Errorf( "no struct given" )
	}

	wrtr.alive = false
	wrtr.conn, err = amqp.Dial( wrtr.url )
	if err != nil {
		return err
	}

	wrtr.ch, err = wrtr.conn.Channel()
	if err != nil {
		wrtr.conn.Close()
		return err
	}

	wrtr.alive = true
	return nil
}

/*
	Builds a url given host, port and user credential information,  then calls connect2url to connect to it.
*/
func (wrtr *Mq_writer) connect2( host string, port string, user string, pw string ) ( err error ) {
	if wrtr == nil {
		return fmt.Errorf( "no struct given" )
	}

	wrtr.url = fmt.Sprintf( "amqp://%s:%s@%s:%s/", user, pw, host, port )
	return wrtr.connect2url()
}

/*
	Reconnect atempts to establish the session assuming that the URL was already deined during the initial 
	connection process. A reconnect attempt will try indefinately until the process is stopped, or a conection
	is established.
*/
func ( wrtr *Mq_writer ) reconnect( ) {
	if wrtr == nil || wrtr.ch != nil {
		return
	}

	announced := false
	wrtr.alive = false						// should be set but be sure

	for {
		err := wrtr.connect2url(  )
		if err == nil {								// connect ok, set up the exchange 
			err = wrtr.ch.ExchangeDeclare( wrtr.ex, wrtr.etype, wrtr.durable, wrtr.auto_del, wrtr.internal, true, nil )
			if err == nil {
				if !wrtr.quiet {
					fmt.Fprintf( os.Stderr, "writer reconnected: %s %v alive=%v\n", wrtr.ex, wrtr, wrtr.alive )
				}
				return
			}

			wrtr.Close( )					// couldn't attach to exchange, so close an try again
			wrtr.alive = false				// ensure this wasn't accidentally set
		}

		if !announced && !wrtr.quiet {
			fmt.Fprintf( os.Stderr, "attempting writer reconnection: %s: %s\n", wrtr.ex, err )
			announced = true
			time.Sleep( 2 * time.Second )		// try every 2 seconds
		}
	}
}

/*
	Ensure that everything is closed down before we go.
*/
func (wrtr *Mq_writer ) Close( ) {
	if wrtr != nil {
		if wrtr.ch != nil {
			wrtr.ch.Close()
			wrtr.ch = nil
		}
		if wrtr.conn != nil {
			wrtr.conn.Close()
			wrtr.conn = nil
		}
	}
}

/*
	Listens on the channel for some kind of 'blob' and then sends it off to the exchange.
	The blob can be string, *string, []byte, or Mq_msg.  Mq_msg is the only way of
	using a key other than the default given when the driver was started.

	If the write channel encounters an error, an attempt to reconnct will be made, and writing 
	will resume when a connection is established, however any messages received on our inbound
	channel will be dropped. The user can test the state after each burst (or individual message)
	using the Is_alive() function which will return false if the connection is down.  This does 
	not make any attempt to preserve messages during an outage.
*/
func (wrtr *Mq_writer) Driver( ) {
	var (
		msg	*amqp.Publishing 
	)

	if wrtr == nil {
		fmt.Fprintf( os.Stderr, "canntot start writer diver: struct is nil\n" )
		return
	}

	mcount := 0					// messages written to prevent complete run away we'll stop at 50 consecutive

	for {
		select {
			case iblob := <- wrtr.Port:			// something to send
				if wrtr.alive {
					msg = nil
					key := wrtr.key

					switch blob := iblob.(type) {
						case string:
  							msg = &amqp.Publishing {
          							ContentType: "text/plain",
          							Body:        []byte( blob ),
  								}
	
						case *string:
  							msg = &amqp.Publishing {
          							ContentType: "text/plain",
          							Body:        []byte( *blob ),
  								}
	
						case []byte:
  							msg = &amqp.Publishing {
          							ContentType: "text/plain",
          							Body:        blob,
  								}
	
						case Mq_msg:									// a struct with the desired write key
  							msg = &amqp.Publishing {
          							ContentType: "text/plain",
          							Body:        blob.Data,
  								}
							key = blob.Key
			
						default:
							if !wrtr.quiet && mcount < 50 {
								fmt.Fprintf( os.Stderr, "rabbit_hole: blob passed to writer is not string, *string, or []byte; not sent\n" )
								mcount++
								if mcount == 50 {
									fmt.Fprintf( os.Stderr, "rabbit_hole: squelching all future blob type messages until a good one received\n" )
								}
							}
					} 

					if msg != nil {
						mcount = 0				// reset 
						err := wrtr.ch.Publish( wrtr.ex, key, !MANDITORY, !IMMED, *msg ) 
						if err != nil {
							if ! wrtr.quiet {
								fmt.Fprintf( os.Stderr, "send error ex=%s: %s\n", wrtr.ex, err )
							}
							wrtr.Close()				// shut it down
							wrtr.alive = false			// prevent further wrte attempts
							go wrtr.reconnect()			// attempt reconnection in parallel
						}
					}
				}

			case req :=  <- wrtr.Notify:		// user request
				switch req {
					case "stop":
						wrtr.stop = true

					case "quiet":
						wrtr.quiet = true

					case "verbose":
						wrtr.quiet = false
				}
		}

		if wrtr.stop {
			if !wrtr.quiet {
					fmt.Fprintf( os.Stderr, "writer driver for %s is stopping\n", wrtr.ex )
			}
			return
		}
	}
}

/*
	Starts the writer driver for the exchange using the given key as the default.
*/
func( wrtr *Mq_writer ) Start_writer( key string ) {
	if wrtr == nil {
		return
	}

	wrtr.key = key
	go wrtr.Driver( )
}

/*
	Create a new writer which connects to the RMQ server and binds to the named exchange.
*/
func Mk_mqwriter( host string, port string, user string, pw string, ex string, ex_type string, key *string ) ( wrtr *Mq_writer, err error ) {

	wrtr = &Mq_writer { 
		ex: ex,
		url: "",
		durable: false,
		auto_del: true,
		internal: false,
	}

	err = wrtr.connect2( host, port, user, pw )
	if err != nil {
		return nil, err
	}

	etokens := strings.Split( ex_type, "+" )
	wait := true
	wrtr.etype = etokens[0]				// fanout, direct, etc.

	for _, v := range etokens {
		switch v {
			case "du":	wrtr.durable = true
			case "!du":	wrtr.durable = false
			case "ad":	wrtr.auto_del = true
			case "!ad":	wrtr.auto_del = false
			case "in":	wrtr.internal = true
			case "!in":	wrtr.internal = false
		}
	}

	err = wrtr.ch.ExchangeDeclare( wrtr.ex, wrtr.etype, wrtr.durable, wrtr.auto_del, wrtr.internal, wait, nil )
	if err != nil {
		wrtr.Close( )
		return nil, err
	}

	wrtr.Port = make( chan interface{}, 1024 )		// user writes things here for driver to send; allow 1k to queue
	wrtr.Notify = make( chan string, 5 )			// allow 5 to queue

	return wrtr, nil
}


/*
	Delete the queue and exchange associated with the writer.
	If force is not true, then deletion only happens if there are 
	no consumers on the queue.
*/
func( wrtr *Mq_writer ) Delete( force bool ) ( err error ) {
	if wrtr == nil || wrtr.ch == nil {
		return
	}

/*	
	_, err = wrtr.ch.QueueDelete( wrtr.qname, force, false, false )		// delete, even if messages queued and wait
	if err {
		wrtr.ch = nil
		return err
	}
*/

	err = wrtr.ch.ExchangeDelete( wrtr.ex, force, false )				// delete the exchange and wait

	wrtr.Close()

	return err	
}

/*
	Stop sets the stop flag in the writer.  If the writer is active it will 
	stop and return.
*/
func ( wrtr *Mq_writer ) Stop( ) {
	if wrtr != nil {
		wrtr.Notify <- "stop"			// break the write loop and stop the driver
	}
}

/*
	Quiet sets the writer's quiet option using the given on/off flag.
*/
func ( wrtr *Mq_writer ) Quiet( onoff bool ) {
	if wrtr != nil {
		wrtr.quiet = onoff
	}
}

/*
	Is alive checks the known state of the connection and returns true if it is thought
	to be alive.
*/
func ( wrtr *Mq_writer ) Is_alive( ) ( bool ) {
	if wrtr != nil {
		return wrtr.alive
	}

	return false
}

// ----------------reader things ----------------------------------------------------------
/*
	Ensure that everything is closed down before we go.
*/
func (rdr *Mq_reader) Close( ) {
	if rdr != nil {
		if rdr.ch != nil {
			rdr.ch.Close()
			rdr.ch = nil
		}
		if rdr.conn != nil {
			rdr.conn.Close()
			rdr.conn = nil
		}

		rdr.qu = nil
		rdr.Port = nil
	}
}

/*
	Create a queue on the other side of the channel with sane defaults.
	The queue will be created with the properties in the rdr struct
	(qautodel etc.).  If the string 'random' was given, a randomly generated
	queue name will be used. If durable == true and autodel == false, then the 
	queue must be named (random not allowed).  

	The queues will persist on the server based on the settings:
		durable !autodelete -- forever (be careful)
		!durable autodelete -- until last connection is dropped
		!durable !autodelete - forever (again, be careful)
		durable autodelete  -- meaningless according to the doc

		durable/!durable can only be connected to the same type of exchange.

	For the long lasting queues if there is no 'listener' the messages will 
	pile on the server and could result in messages being paged to disk 
	which could result in full disk. These queues must also be named.

*/
func (rdr *Mq_reader) mk_queue( ) ( err error ) {
	if rdr == nil {
		return fmt.Errorf( "nil reader passed" )
	}
	if rdr.ch == nil {				// blasted underlying lib doesn't nil check
		return fmt.Errorf( "channel is nil, unable to alloc queue" )
	}

	if rdr.qname == "random" {
		rdr.qname = ""
	}

	if !rdr.qautodel && rdr.qname == "" {
		return fmt.Errorf( "perminant queues (!autodel) must be named" )
	}

	//qu, err := rdr.ch.QueueDeclare( rdr.qname, !DURABLE, !AUTO_DEL, !EXCLUSIVE, WAIT, nil )
	qu, err := rdr.ch.QueueDeclare( rdr.qname, rdr.qdurable, rdr.qautodel, rdr.qexclusive, WAIT, nil )
	if err == nil {
		rdr.qu = &qu
	}

	return  err
}

/*
	Create a randomly named queue to bind to an exchange; deleted when the 
	channel closes. Random queues are _always_ not durable, auto delete and
	exclusive so that they are deleted on close.
*/
func (rdr *Mq_reader) mk_rand_queue( ) ( err error ) {
	qu, err := rdr.ch.QueueDeclare( "", !DURABLE, AUTO_DEL, EXCLUSIVE, WAIT, nil )
	if err == nil {
		rdr.qu = &qu
	}
	return err
}

/*
	Bind an exchange to a queue. If queue is nil, then a random queue will be created.
*/
func (rdr *Mq_reader) bind_q2ex( key *string ) ( err error ) {
	//if rdr.qu == nil {
	//	err = rdr.mk_rand_queue( ) 
	//	if err != nil {
	//		err = fmt.Errorf( "random queue failed: %s\n", err )
	//		return err
	//	}
	//}

	err = rdr.mk_queue( )
	if err != nil {
		return fmt.Errorf( "unable to create queue: %s", err )
	}

	return rdr.ch.QueueBind( rdr.qu.Name, *key, rdr.ex, false, nil )
}

/*
	Create a connection and then open a channel using the url that exists in the reader. 
	The caller must close both before exit.
*/
func (rdr *Mq_reader) connect2url( ) ( err error ) {
	if rdr == nil || rdr.url == "" {
		return fmt.Errorf( "no struct provided" )
	}

	if rdr.ch  != nil {
		rdr.Close()
	}
	
	rdr.conn, err = amqp.Dial( rdr.url )
	if err != nil {
		return err
	}

	rdr.ch, err = rdr.conn.Channel()
	if err != nil {
		rdr.conn.Close()
		return err
	}

	return nil
}

/*
	Connect drives all of the steps needed to connect to an exchange once the url and exchange info
	have been placed into the structure.  It can be called for an initial connection, or to reconnect
	when a session is lost. Error on return is nil if successful.
*/
func ( rdr *Mq_reader ) connect( ) ( err error ) {

	if rdr.ch!= nil {
		rdr.Close( )
	}

	err = rdr.connect2url( ); 			// attempt a reconnect using prebuilt url
	if err == nil {
		err = rdr.ch.ExchangeDeclare( rdr.ex, rdr.etype, rdr.durable, rdr.auto_del, rdr.internal, true, nil )
		if err == nil {
			err = rdr.bind_q2ex( &rdr.key ) 
			if err == nil {
				rdr.Port, err = rdr.ch.Consume( rdr.qu.Name, "", AUTO_ACK, !EXCLUSIVE, !LOCAL, WAIT, nil )
			}
		}
	}

	if err != nil {
		rdr.Port = nil
		rdr.Close()
	}

	return err
}

/*
	Create a reader for a given user/pw host/port exchange 5-tuple.
	Creates a connection and channel. 

	The exchange type has the following syntax:
		[type][+eop+eop..][>[qname][+qop+qop...]

	Type is fanout etc. Eop is one of the following and can be !eop to 
	negate:
		ad == autodel
		du == durable
		in == internal

	qname is the name of the queue, if >+qop is coded, then a random name
	is gnerated. The qops are:
		ad == auto delete
		du == durable
		ex == exclusive

	Any attribute may be prefixed with ! to negate it (e.g. !ad) and the order
	of either type of op is NOT important.

	The queue type must match the exchange type (durable or !durable) and setting the
	type for the exchage will also set the type for the queue, and thus setting
	du or !du in the queue is optional.
*/
func Mk_mqreader( host string, port string, user string, pw string, ex string, ex_type string, key *string ) ( rdr *Mq_reader, err error ) {
	rdr = &Mq_reader { 
		ex: ex,
		url: fmt.Sprintf( "amqp://%s:%s@%s:%s/", user, pw, host, port ),
		key: *key,
		qname: "",

		durable: false,
		auto_del: true,
		internal: false,

		qdurable: false,
		qautodel: true,
		qexclusive: false,

		qu:	nil,
	}

	tokens := strings.Split( ex_type, ">" )					// possibly [type][+op+op]>[queue][+op+op]
	etokens := strings.Split( tokens[0], "+" )				// tease out any exchange options
	for _, v := range etokens {
		switch v {
			case "du":	rdr.durable = true
						rdr.qdurable = true					// since these must match, we'll set both here so they can be omitted in the queue set if desired

			case "!du":	rdr.durable = false
						rdr.qdurable = false

			case "ad":	rdr.auto_del = true
			case "!ad":	rdr.auto_del = false
			case "in":	rdr.internal = true
			case "!in":	rdr.internal = false
		}
	}
	rdr.etype = etokens[0]				// will be "" if no type  +op>qname or just >qname

	if len( tokens ) > 1 {								// etype also had >queue name or >qname+option(s)
		qtokens := strings.Split( tokens[1], "+" )
		for _, v := range qtokens {						// if >+op  then random queue and no name given
			switch v {
				case "ad":	rdr.qautodel = true
				case "!ad":	rdr.qautodel = false
				case "du":	rdr.qdurable = true
				case "!du":	rdr.qdurable = false
				case "ex":	rdr.qexclusive = true
				case "!ex":	rdr.qexclusive = false
			}
		}

		rdr.qname = qtokens[0]
	}

	err =  rdr.connect( )			// for initial connection we try just once
	if err == nil {
		return rdr, nil
	}

	return nil, err
}

/*
	Wait for messages and write them on the channel. 
	If we have a disconnect, then we attempt to reconnect and 
	will resume reading once connected. If the stop funciotn is 
	called, then we will abort. If pause is invoked then the 
	messages read are dropped.
*/
func (rdr *Mq_reader) eat( usr_ch chan amqp.Delivery ) {
	dcount := 0

	for {
		for msg := range rdr.Port {		// read until close (session broken)
			if rdr.paused {
				if dcount % 100 == 0  {
					if( ! rdr.quiet ) {
						fmt.Fprintf( os.Stderr, "rabbit_hole: read paused -- %d messages dropped\n", dcount + 1 )
					}
					dcount = 0;
				} 
				dcount++
			} else {
				usr_ch <- msg
			}
		}

		//TODO: check for stop and stop :)

		if !rdr.quiet {
			fmt.Fprintf( os.Stderr, "rabbit_hole: read session disconnected, reconnection in progress: %s\n", rdr.ex )
		}
		for {
			if rdr.connect( ) == nil {
				break
			}
			time.Sleep( 2 * time.Second )
		}

		if !rdr.quiet {
			fmt.Fprintf( os.Stderr, "rabbit_hole: read session reconnected: %s\n", rdr.ex )
		}
	}
}

/*
	Pause will set the pause flag in the reader based on the on/off state that is
	passed in.  When paused is true, then messages which are receivd from the 
	Rabbit exchange are dropped. When paused, and quiet is off, periodic dropped
	counts are written to the standard error device.
*/
func ( rdr *Mq_reader ) Pause( onoff bool ) {
	if rdr != nil {
		rdr.paused = onoff
	}
}

/*
	Stop sets the stop flag in the reader.  If the reader is active it wll 
	stop and return.
*/
func ( rdr *Mq_reader ) Stop( ) {
	if rdr != nil {
		rdr.stop = true
	}
}

/*
	Quiet sets the reader's quiet option using the given on/off flag.
*/
func ( rdr *Mq_reader ) Quiet( onoff bool ) {
	if rdr != nil {
		rdr.quiet = onoff
	}
}


/*
	Start_eating causes the package to beging processin messages from the rabbit 
	channel and passing them on the user channel, or dropping them if paused.
*/
func( rdr *Mq_reader) Start_eating( usr_ch chan amqp.Delivery ) {
	go rdr.eat( usr_ch )
}
