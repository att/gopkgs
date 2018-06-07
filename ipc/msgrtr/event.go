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

	Mnemonic:	event.go
	Abstract:	Event struct and related functions. The event struct is organised
				with publically visible fields to allow for json unmarshalling and
				to make it easier for the listener to process.  There are some 	
				private fields which are visable only to this package (e.g. the 
				related data block for replies).
	Date:		30 Oct 2015
	Author:		E. Scott Daniels

*/

package msgrtr

import (
	"fmt"
	"sync"
	"github.com/att/gopkgs/ipc"
)


/*
	An event expected on a POST request. Multiple events may be bundled
	int a single data block as an array.
*/
type Event struct {
	Event_type string				// dot separated 'path' or band (e.g. network.router.add)
	Ack		bool					// true an ack is expected by a listener
	Payload	map[string]interface{}	// event message content

								// ------ private fields ----------------
	db		*Data_block			// reference back for responses
	msg		string				// ack string to send
	ack_sent	bool			// only allow one ack message per event
	mu		*sync.Mutex			// replies must have the lock
}

/*
	Send the event to all who have registered in the given gallary.
*/
func ( e *Event ) bcast( gallery *audience ) ( acks_needed int, err error ) {
	acks_needed = 0

	sent := gallery.bcast( e, e.Event_type )
	if e.Ack {
		acks_needed++
		if ! sent {
			err = fmt.Errorf( "acks needed and no listener" )
		}
	}

	return acks_needed, err
}

/*
	Adds a reference to the controlling data_block so that we have ack and writer 
	info if we need to send some response.
*/
func ( e *Event ) add_db( db *Data_block ) {
	e.db = db
}

/*
	Because this class is instanced by unmarshall, it isn't initialised with the goodies we need
	so we provide the means for things like the mutex.	
*/
func ( e *Event )  add_mutex() {
	e.mu = &sync.Mutex{}
}

func ( e *Event ) get_db( ) ( *Data_block ) {
	return e.db
}

// ------------ event public ----------------------------------------------------------------------------

/*
	Returns the currently formatted message.
*/
func ( e *Event ) Get_msg( ) ( string ) {
	return e.msg
}

/*
	Path returns the path of this event.
*/
func ( e *Event ) Path() ( string ) {
	return e.Event_type
}

/*
	Reply sends a reply back to the http requestor. This is a wrapper that puts a 
	request on the dispatcher queue so that we serialise the access to the underlying
	data block. Status is presumed to be OK or ERROR or somesuch. msg is any
	string that is a 'commment' and data is json or other data (not quoted in the
	output).
*/
func ( e *Event ) Reply( state string, msg string, data string ) {
	e.mu.Lock()
	if e.ack_sent {
		e.mu.Unlock()
		return
	}
	e.ack_sent = true			// set now then release the lock; no need to hold others while we write
	e.mu.Unlock()

	if data != "" {
		e.msg = fmt.Sprintf( `{ "endstate": { "status": %q, "comment": %q, "data": %s} }`, state, msg, data )
	} else {
		e.msg = fmt.Sprintf( `{ "endstate": { "status": %q, "comment": %q } }`, state, msg )
	}

	cmsg := ipc.Mk_chmsg()
	cmsg.Send_req( disp_ch, nil, SEND_ACK, e, nil )            // queue the event for a reply
}

/*
	String returns a string describing the instance of the structure.
*/
func ( e *Event ) String( ) ( string ) {
	return fmt.Sprintf( "%s ack=%v", e.Event_type, e.Ack )
}


