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

	Mnemonic:	chmsg
	Abstract:	Channel message object and related methods.
	Date:		02 December 2013
	Author:		E. Scott Daniels

*/

/*
	The chmsg provides a data structure that helps to manage interprocess
	communications between Goroutines via channels.
*/
package ipc

import (
)


// --------------------------------------------------------------------------------------
/*
	Defines a message. Request and response data pointers are used so that in the case
	where the requesting thread is not blocking on the response the request data will
	be available for response processing if needed. A third data field, requestor_data
	is another hook to assist with asynch response processing. This is assumed to point
	to a data object that is of no significance to the receiver of the message when
	the message is a request.  It should be nil if the response channel is also nil.

	The struct allows the application to define how it is used, and does not really imply
	any meaning on the fields.
	
*/
type Chmsg struct {					// all fields are externally available as a convenence
	Msg_type	int;				// action requested
	Response_ch	chan *Chmsg;		// channel to write this request back onto when done
	Response_data	interface{};	// what might be useful in addition to the state
	Req_data	interface{};		// necessary information to processes the request
	Requestor_data	interface{};	// private data meaningful only to the requestor and
									// necessary to process an asynch response to the message.
	State		error;				// response state, nil == no error
}

/*
	constructor
*/
func Mk_chmsg( ) (r *Chmsg) {

	r = &Chmsg { };
	return;
}


// ---- these are convenience functions that might make the code a bit easier to read ------------------

/*
	Send the message as a request oriented packet.
	Data is mapped to request data, while response data and state are set to nil.
	Pdata is the private requestor data.
*/
func (r *Chmsg) Send_req( dest_ch chan *Chmsg, resp_ch chan *Chmsg, mtype int, data interface{}, pdata interface{} ) {
	r.Msg_type = mtype;
	r.Req_data = data;
	r.Response_ch = resp_ch;
	r.Requestor_data = pdata;
	r.Response_data = nil;
	r.State = nil;

	dest_ch <- r;				// this may block until the receiver reads it if the channel is not buffered
}

/*
	Send the request as a response oriented packet.
	The response channel is used to send the block.
	Data is mapped to response data and state is set. All other fields are left unchanged.
*/
func (r *Chmsg) Send_resp( data interface{}, state error ) {
	r.Response_data = data;
	r.State = state;

	r.Response_ch <- r;				// this may block until the receiver reads it if the channel is not buffered
}
