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

	Mnemonic:	audience.go
	Abstract:	An audience is a tree of groups of listeners. Each node represents a 
				specific set of messages (e.g. network link add) and contains a list of 
				channels to broadcast matching messages to.  A node's children are referenced
				by a map suc that the network node might have 'link' and 'router' children
				while the endpoint node might just have 'add', 'del', and 'mod' children.

	Date:		30 Oct 2015
	Author:		E. Scott Daniels

*/

package msgrtr

import (
	"fmt"
	"strings"
)


type audience struct {				// describes a group of listeners
	chlist	[]chan *Event			// direct listeners for this type (all)
	subsect	map[string]*audience	// subsections (children)
}


/*
	Create an audience structure. subsect is the list of nodes in the tree
	below this node (e.g. network link add, or network link del). Subsects can be
	empty ("") if this is the leaf.
*/
func mk_audience( subsect string, listener chan *Event ) ( a *audience ) {
	
	a = &audience { }

	a.chlist = make( []chan *Event, 0, 16 )			// this will automatically extend if needed
	a.subsect = make( map[string]*audience, 16 )	// 16 is a hint, not a hard limit

	if subsect != "" {
		tokens := strings.SplitN( subsect, ".", 2 )
		rest := ""
		if len( tokens ) > 1 {
			rest = tokens[1]
		}

		a.subsect[tokens[0]] = mk_audience( rest, listener )
	} else {
		if listener != nil {
			a.chlist = append( a.chlist, listener )					// add the first listener
		}
	}

	return a
}

/*
	Add a listener to an audience. Creates the child sections if needed.
	Subect is the rest of the path below this node where the listener is 
	to be placed.
*/
func ( a *audience ) add_listener( subsect string, listener chan *Event ) {
	if subsect != "" {
		tokens := strings.SplitN( subsect, ".", 2 )
		rest := ""
		if len( tokens ) > 1 {
			rest = tokens[1]
		}
		if a.subsect[tokens[0]] == nil {								// new child; make it
			a.subsect[tokens[0]] = mk_audience( rest, listener )
		} else {
			a.subsect[tokens[0]].add_listener( rest, listener )							// just add if existing
		}
	} else {												// at the leaf, add the listener to this list
		if listener != nil {
			a.chlist = append( a.chlist, listener )			// this will grow the slice if needed so must reassign
		}
	}
}

/*
	Remove a listener to an audience. Creates the child sections if needed.
	Subect is the rest of the path below this node where the listener is 
	to be placed.
*/
func ( a *audience ) drop_listener( subsect string, listener chan *Event ) {
	if listener == nil {
		return
	}

	for k, v := range a.chlist {					// it could be here if full event path was longer, so must check
		if v == listener {
			new_len := len( a.chlist ) - 1
			cl := make( []chan *Event, new_len, new_len + (new_len/2) )		// some room to grow
			if k > 0 {
				copy( cl, a.chlist[0:k-1] )
			}
			if k < len( a.chlist ) -1 {
				cl = append( a.chlist[k+1:] )
			}

			a.chlist = cl
			return
		}
	}

	if subsect != "" {													// just pass further down
		tokens := strings.SplitN( subsect, ".", 2 )
		rest := ""
		if len( tokens ) > 1 {
			rest = tokens[1]
		}
		if a.subsect[tokens[0]] != nil {								// existing child; send it on, no action if child doesn't exist
			a.subsect[tokens[0]].drop_listener( rest, listener )
		} 
	} 
}

/*
	Send the event to all audience members.  Path is the current point in the 
	message band which is of the form a/b/c.  We use 'a' to find the audience 
	for the next level down assuming that our level string has been removed.
	We pass b/c (maybe more) to the next lower structure in the tree where 
	the process repeats.
	
	Returns boolean indicating whether there was actually a listener which matched.  
	If there wasn't, and the message has the ack flag on, the top level must send 
	it to the dev/null path so that it can be acked with a 'missed' message.
*/
func ( a *audience ) bcast( event *Event, path string ) ( bool ){

	sent := false

	for i := range a.chlist {				// for each listener registered here
		if a.chlist[i] != nil {
			sent = true
			sheep.Baa( 2, "send message to next channel" )
			a.chlist[i] <- event
		}
	}

	if path == "" {
		return sent
	}

	child_sent := false
	tokens := strings.SplitN( path, ".", 2 )
	if sa := a.subsect[tokens[0]]; sa != nil {
		sheep.Baa( 1, "send to child: %s", tokens[0] )
		if len( tokens ) > 1 {
			child_sent = sa.bcast( event, tokens[1] )
		} else {
			child_sent = sa.bcast( event, "" )
		}
	}

	return sent || child_sent
}

func ( a *audience ) String( ) ( string ) {
	s := fmt.Sprintf( "%d listeners, %d children", len( a.chlist ), len( a.subsect ) )
	for  k, v := range a.subsect {
		s += fmt.Sprintf( " <%s: %s> ", k, v )
	}

	return s
}
