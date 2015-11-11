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

	Mnemonic:	msgrtr
	Abstract:	Provides a message routing interface.  The user application invokes
				the Start() function to begin listening using http on the given port.
				Messages received are 'unpacked' and then routed to the user application's
				threads.

				The 'url' passed on the start call is the only one which is supported.
				Threads in the application can register and provide a channel on which
				event structures are delivered for matching events.  The registration
				defines the 'path' which match messages (e.g. network/router/add). Paths
				may be 'short' (e.g. network/router) to receive all messages which match
				to that depth.  Multiple listeners can register for the same path; all 
				messages are 'broadcast' to all listeners; there is no concept of 
				bubble up or handling which prevent the event from being delivered to 
				some of the listeners.

				The listening thread can respond (reply) to an event, and at least one
				listener must reply if the ack field is set to true. If no listeners
				reply the sending application will hang until timeout. Only one reply 
				per event is sent, the first, all others are silently discarded. Replys
				are sent by using the event's reply function.

				Event messages are expected to be posted to the url as a json object
				with some known set of fields, and optionally some meta information 
				which the listener(s) may need.   The 'band' field describes the 
				'path' (e.g. network/router) and the action type will be appended to 
				the path provided that the action is not empty or missing.

				When an event is received it is written on the channel of all listeners
				which have registered for the event type.  The contents pushed on the 
				channel is a *msgrtr.Event which has public fields so that they are 
				easily accessed by the user programme. Speifically, the event type
				and paylod are probably what is of the most interest. The payload is
				is a map, indexed by string, which references interface{} elements. 
				The event type is the _complete_ type; not just the portion of the 
				type that the listener registered.  For example, if the listener
				registered network.swtich (wanting all events for network switches)
				the event types generated would include:  network.switch.add, network.switch.del
				network.switch.mod, etc.  If a listener registers only for a specific type,
				then that will be the only type delivered.  

				The third  field of interest, and to which the user process must pay attention
				to, is the Ack field.  If true, one of the listeners _must_ call event.Reply()
				to send a reply to the sender.  

				Event types are determined by the message generator, the process sending
				the http request to this process, and are _not_ controlled by this package.
				Same goes for the payload map.  The keys are up to the message generator.

	Date:		30 Oct 2015
	Author:		E. Scott Daniels

*/

package msgrtr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/att/gopkgs/bleater"
	"github.com/att/gopkgs/ipc"
)

const (
	NOP 		int =	0
	SEND_ACK 	int = 	iota 	// user send an ack
	RAW_BLOCK					// raw data block from http
	REGISTER					// listener thread registration
	UNREGISTER
)

/*
	Because of the nature of the http listener which doesn't provide for a mechanism
	to pass data to the functions that are called, we are forced to rely on global
	data.  
*/
var (
	sheep *bleater.Bleater		// bleater, possibly from the user
	disp_ch chan *ipc.Chmsg		// dispatcher's listen channel
)

/*
	Message user app sends to register for a message or set of messages.
*/
type Reg_msg struct {
	ch			chan *Event		// channel that the user wants messages on
	band		string			// message "band" to listen to (e.g. network router add)
}

/*
	Block passed to the dispatcher to, well, dispatch.
*/
type Data_block struct {
	Events []*Event					// events from the json http message (unmarshal)

									// info added to track the event(s)
	rel_ch	chan int				// releases the http dealwith function for the request
	out		http.ResponseWriter		// where json should be written
	acks_needed int					// number of acks needed
	ack_count	int					// number of acks sent
}


// ------------ HTTP listener functions -------------------------------------------

/*
    Pull the data from the request (the -d stuff from churl;  -D stuff from rjprt).
	We really want an array of events, but some senders might not want to play that
	game, so we will first look for an array, and if we don't find one we will assume
	that it's a singleton, and will wrap the json in '{event: [ <json> ] }' so that 
	we end up with one. Output is written directly to the interface datablock
	that is passed in (assuming it has Events field).
*/
func dig_data( resp *http.Request, data_blk interface{} ) ( err error ) {
	data, err := ioutil.ReadAll( resp.Body )
	resp.Body.Close( )
	if( err != nil ) {
		sheep.Baa( 1, "unable to dig data from the request: %s", err )
		return 
	}


	err = json.Unmarshal( data, data_blk )
	if err != nil {
		sheep.Baa( 1, "msgrtr: unable to extract events from data: bad json: %s", err )
	}

	if err != nil {	
		return err
	}

	if db, ok := data_blk.( *Data_block ); ok {
		if len( db.Events ) <= 0 {										// no events, assume it was just a singleton, reformat and reparse
			sdata := `{ "events": [ ` + string( data ) + ` ] }`
			err = json.Unmarshal( []byte( sdata ), data_blk )
			if err != nil {
				sheep.Baa( 1, "msgrtr: unable to extract events from reformatted data: bad json: %s", err )
			}
		}
	}

	return err
}


/*
	Deal with input from the other side sent to the http url. 
	This function is invoked directly by the http listener and as such we get no 'user data'
	so we rely on globals in order to be able to function.  (There might be a way to deal
	with this using a closure, but I'm not taking the time to go down that path until
	other more important things are implemented.)

	We assume that the body contains one complete json struct which might contain
	several messages.

	This is invoked as a goroutine by the http environment and when this returns the 
	session to the requestor is closed. So, we must ensure that we block until all
	output has been sent to the session before we return.  We do this by creating a 
	channel and we wait on a single message on that channel.  The channel is passed in 
	the datablock. Once we have the message, then we return.
*/
func deal_with( out http.ResponseWriter, in *http.Request ) {
	var (
		state	string = "ERROR"
		msg		string
	)

	out.Header().Set( "Content-Type", "application/json" )				// announce that everything out of this is json
	out.WriteHeader( http.StatusOK )									// if we dealt with it, then it's always OK; requests errors in the output if there were any

	sheep.Baa( 0, "dealing with a request" )

	data_blk := &Data_block{}
	err := dig_data( in, data_blk )

	if( err != nil ) {													// missing or bad data -- punt early
		sheep.Baa( 1, "msgrtr/http: missing or badly formatted data: %s: %s", in.Method, err )
		fmt.Fprintf( out, `{ "status": "ERROR", "comment": "missing or badly formatted data: %s", err }` )	// error stuff back to user
		return
	}

	switch in.Method {
		case "PUT":
			msg = "PUT requests are unsupported"

		case "POST":
			sheep.Baa( 1, "deal_with called for post" )

			if len( data_blk.Events ) <= 0 {
				sheep.Baa( 1, "data block has no events?????" )
			} else {
				data_blk.out = out
				data_blk.rel_ch = make( chan int, 1 )

				state = "OK"
				sheep.Baa( 1, "data: type=%s", data_blk.Events[0].Event_type )
				req := ipc.Mk_chmsg( )
				req.Send_req( disp_ch, nil, RAW_BLOCK, data_blk, nil )			// pass to dispatcher to process
				<- data_blk.rel_ch												// wait on the dispatcher to signal ok to go on; we don't care what comes back
			}
			

		case "DELETE":
			msg = "DELETE requests are unsupported"

		case "GET":
			msg = "GET requests are unsupported"

		default:
			sheep.Baa( 1, "deal_with called for unrecognised method: %s", in.Method )
			msg = fmt.Sprintf( "unrecognised method: %s", in.Method )
	}

	if state == "ERROR" {
		fmt.Fprintf( out, fmt.Sprintf( `{ "endstate": { "status": %q, "comment": %q } }`, state, msg ) )		// final, overall status and close bracket
	}
}



// ------------ private functions -------------------------------------------

/*
	Expected to be started as a go routine which:
		1) listens for registration requests from the user application
		2) receives messages from the http environment so that we can ensure
			sequential distribution should an event contain multiple 
			messages.

	The dispatcher is single threaded and thus guarentees that the events
	received in a single message are distributed in the order received. The
	number of acks needed is recorded, and if none are asked for, then we 
	release the http deal with function straight away, otherwise we hold it
	until the number of expected send_ack messages have been recceived back
	from the user programme.
*/
func dispatcher( ch chan *ipc.Chmsg ) {
	gallery := mk_audience( "", nil )						// initialise the audence tree

	for {
		req := <- disp_ch									// listen for requests from http world, or from users
		if req == nil {										// parnoia saves us in the long run
			sheep.Baa( 1, "nil event on channel" )
			continue
		}

		switch req.Msg_type {

			case RAW_BLOCK:
				ec := 0											// error count
				if db, ok := req.Req_data.( *Data_block ); ok {
					db.acks_needed = 0
					for i := range db.Events {
						e := db.Events[i]
						e.add_mutex( )
						e.add_db( db )							// reference the db for acks
						ac, err := e.bcast( gallery )
						if err != nil {
							ec++
							sheep.Baa( 1, "dispatch: no listener for event: %s", e )
						} else {
							db.acks_needed += ac
						}
					}

					if ec > 0 {									// any error invalidates the whole chain, toss a warning back now
							fmt.Fprintf( db.out, fmt.Sprintf( `{ "endstate": { "status": %q, "comment": %q } }`, "ERROR", "no listener for some/all events requiring ack" ) )
							db.rel_ch <- 1						// release the deal_with instance
					} else {
						if db.acks_needed <= 0 {
							fmt.Fprintf( db.out, fmt.Sprintf( `{ "endstate": { "status": %q, "comment": %q } }`, "OK", "Got it" ) )		// nothing to wait on, just send response now
							db.rel_ch <- 1						// release the deal_with instance
						}
					}
				} else {
					sheep.Baa( 1, "dispatch: internal mishap processing raw block: doesn't seem to be a block ptr" )
				}

			case REGISTER:									// msg from user to register for an event band
				if reg, ok := req.Req_data.( *Reg_msg ); ok {
					sheep.Baa( 1, "dispatch: registering a listener for: %s", reg.band )
					gallery.add_listener( reg.band, reg.ch )
					
					sheep.Baa( 1, "audience: %s", gallery )
				} else {
					sheep.Baa( 1, "dispatch: internal mishap: bad message struct on register" )
				}

			case SEND_ACK:									// send an ack/response message for an event (event expected in Req_data and ack json in Response_data)
				if e, ok := req.Req_data.( *Event ); ok {
					if db := e.get_db(); db != nil {
						fmt.Fprintf( db.out, "%s", e.Get_msg() )
						db.acks_needed--
						if db.acks_needed <= 0 {
							db.rel_ch <- 1						// release the deal_with instance as all needed acks were sent
						}
					}
				} else {
					sheep.Baa( 1, "dispatch: internal mishap: bad event on send-ack" )
				}

			case UNREGISTER:								// msg from user to unregister for an event band
				if reg, ok := req.Req_data.( *Reg_msg ); ok {
					sheep.Baa( 1, "dispatch: unregistering a listener for: %s", reg.band )
					gallery.drop_listener( reg.band, reg.ch )
					
					sheep.Baa( 1, "audience: %s", gallery )
				} else {
					sheep.Baa( 1, "dispatch: internal mishap: bad msg struct on unregister" )
				}

			default:
				sheep.Baa( 1, "dispatch: unrecognised request received on channel was ignored: type=%d", req.Msg_type )
		}
	}
}

/*
	Ivoked by start() as a go routine since the http function doesn't return.
	This sets up for, and then invokes the http listener which will send all 
	http requests to our function that deals with such things.
*/
func listen( url string, port string ) {

	/*
		FUTURE:   this needs to be extended to support https
		if  create_cert {
			http_sheep.Baa( 1, "creating SSL certificate and key: %s %s", *ssl_cert, *ssl_key )
			dns_list := make( []string, 3 )
			dns_list[0] = "localhost"
			this_host, _ := os.Hostname( )
			tokens := strings.Split( this_host, "." )
			dns_list[1] = this_host
			dns_list[2] = tokens[0]
			cert_name := "tegu_cert"
			err = security.Mk_cert( 1024, &cert_name, dns_list, ssl_cert, ssl_key )
    		if err != nil {
				http_sheep.Baa( 0, "ERR: unable to create a certificate: %s %s: %s  [TGUHTP001]", ssl_cert, ssl_key, err )
			}
		}
		err = http.ListenAndServeTLS( ":" + *api_port, *ssl_cert, *ssl_key,  nil )		// drive the bus
	*/

	sheep.Baa( 1, "msgrtr: listening on port %s for %s", port, url )

	if url[0:1] != "/" {									// handle func registry needs the lead slant; reason unknown
		url = "/" + url										// so add it if missing.
	}

	http.HandleFunc( url, deal_with )						// invoke deal_with function for all messages received on the url
	if strings.Index( port, ":" ) < 0 {
		port = ":" + port
	}
	err := http.ListenAndServe( port, nil )			// drive the bus
	if err != nil {
		sheep.Baa( 0, "msgrtr: unable to initialise http interface on url, port %s %s", url, port )
	}
}


// ------------- public functions -------------------------------------------

/*
	A wrapper allowing a user thread to register with a function call rather than 
	having to send a message to the dispatcher.
*/
func Register( band string, ch chan *Event ) {
	reg := &Reg_msg {
		band: band,
		ch: ch,
	}

	msg := ipc.Mk_chmsg()
	msg.Send_req( disp_ch, nil, REGISTER, reg, nil )            // send the registration to dispatcher for processing
}

/*
	A wrapper allowing a user thread to unregister with a function call rather than 
	having to send a message to the dispatcher.
*/
func Unregister( band string, ch chan *Event ) {
	reg := &Reg_msg {
		band: band,
		ch: ch,
	}

	msg := ipc.Mk_chmsg()
	msg.Send_req( disp_ch, nil, UNREGISTER, reg, nil )            // send the registration to dispatcher for processing
}

/*
	Initialises the message router and returns the channel that it will 
	accept retquest (ipc structs) on allowing the user thread(s) to 
	register for messages.  Port is the port that the http listener should
	camp on, and url is the url string that should be used.  Port may be of either of
	these two forms:
		interface:port
		port

	If interface is supplied, then the listener will be started only on that interface/port
	combination. If interface is omitted, then the listener will listen on all interfaces.
	This funciton may be invoked multiple times, with different ports, but be advised that 
	all messages are funneled to the same handler.  Multiple invocations only serve to 
	establish different interfaces and/or ports.
*/
func Start( port string, url string, usr_sheep *bleater.Bleater ) ( chan *ipc.Chmsg ) {
	if usr_sheep != nil {
		sheep = usr_sheep								// baa with user supplied bleater, else crate our own
	} else {
		sheep = bleater.Mk_bleater( 1, os.Stderr )		// if no bleater from user, only bleat critical things
		sheep.Set_prefix( "msgrtr" )
	}

	if disp_ch == nil {
		disp_ch = make( chan *ipc.Chmsg, 1024 )
		go dispatcher( disp_ch )
	}

	go listen( url, port )

	return disp_ch
}

