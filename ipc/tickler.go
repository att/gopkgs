// vi: sw=4 ts=4:

/*

	Mnemonic:	tickler
	Abstract:	manage a tickler that writes a request onto a channel every so often
				to tickle the receiver.
	Date:		17 December 2013
	Author:		E. Scott Daniels

*/

package ipc

import (
	"fmt"
	//"os"
	"sync"
	"time"
)

const (
	FOREVER	int = 0;
)

/*
	Manages the user's list of tickle spots.
*/
type Tickler struct {
	tlist	[]*tickle_spot;
	tidx	int;				// current index
	ok2run	bool;
	isrunning	bool;			// allows the first add to start the goroutine
	mu sync.Mutex;
}

/*
	Manages one particular tickle spot.
*/
type tickle_spot struct {
	ch			chan *Chmsg;
	req_type	int;
	req_data	interface{};
	delay		int64;
	count		int;			// number of times to tickle before stoping; 0 == forever
	nextgo		int64;			// unix time for the next tickling of this spot
	req			*Chmsg;		// each will have it's own so we don't thrash mem alloc
}

/*
	While there are active tickle blocks, loop and tickle each spot when it is time.
	If a block uses all of its count up then we mark it off and never tickle it again.
	If we get a stop we'll return, but that won't happen until after the current sleep 
	ends, but we will guarentee not to tickle anything after stop has been set. 

	We will block on a tickle notification if the user's channel isn't buffered. This 
	might delay other tickles, but that's completely in the user control. 

	If a tickle_spot is added with a shorter delay than the current time remaining in the 
	sleep, the first tickling will happen at the time of wakeup which will be after the 
	perceived first time it should be tickled.  Yes, this might be considered a bug, but
	if tickles are added shortest delay to longest, or the tickler is stopped before the 
	tickles are added, then started, it can be avoided without having to wake at constants
	settings and doing nothing if nothing was ready to fire. 
*/
func (t *Tickler) tickle_loop( ) {
	var (
		sleep_len	time.Duration;
		now			int64;
		delay		int64;
	)

	t.isrunning = true;

	for ; ; {
		if !t.ok2run {
			break;
		}

		delay = 0;
		for i := 0; i < t.tidx; i++ {
			if t.tlist[i] != nil {
				if t.tlist[i].ch != nil	{					// nil channel indicates that it's been stopped
					now = time.Now().Unix();				// we'll compute this for each spot to prevent odd things if sending a request blocks
	
					if t.tlist[i].nextgo <= now {				// time to drive this one
						//fmt.Fprintf( os.Stderr, "sending tickle: %d\n",  t.tlist[i].req_type )
						t.tlist[i].req.Send_req( t.tlist[i].ch, nil, t.tlist[i].req_type, t.tlist[i].req_data, nil );	// no response expected so return channel is nil
						t.tlist[i].nextgo = now + t.tlist[i].delay;
	
						if t.tlist[i].count > 0 {				// a counter, we dec it and if it reaches 0 then we set ch to nil to stop this spot
							t.tlist[i].count--;
							if t.tlist[i].count == 0 {
								t.tlist[i].ch = nil;
								t.tlist[i].nextgo = 0 //delay;		// now dropped, prevent it from causing an early wakeup
							}
						}
					}
		
					if (delay == 0 || t.tlist[i].nextgo < delay)  && t.tlist[i].nextgo > now {
						//fmt.Fprintf( os.Stderr, "tickle delay: ty=%d old=%d new=%d\n",  t.tlist[i].req_type, delay,  t.tlist[i].nextgo )
						delay = t.tlist[i].nextgo
					}
				} 
			}
		}

		if delay <= 0 {				// nothing left in the list, might as well stop
			break;
		}

		sleep_len = time.Duration( delay - now ) * time.Second;		// compute the next wakeup time -- seconds from now and nap
//fmt.Fprintf( os.Stderr, "tickle sleeping: %d \n",  delay - now )
		if sleep_len < 0 { 
			sleep_len = 1 * time.Second
		}
		time.Sleep( sleep_len );
	}

	t.isrunning = false;
}

// ------------- public ---------------------------------------

/*
	Create a tickle manager that can handle up to max concurrent tickles.
	We'll cap tickles at 1024 and default to 100 if 0 passed in as max.
*/
func Mk_tickler( max int ) ( t *Tickler ) {
	t = &Tickler { ok2run: true }

	if max > 1024 {			// sliently enforce sanity
		max = 1024;
	} else {
		if max <= 0 {
			max = 100;
		}
	}

	//fmt.Fprintf( os.Stderr, "tickler: object created: %d entries\n", max );
	t.tlist = make( []*tickle_spot, max );
	return;
}

/*
	Adds something to the tickle list.  Delay is the number of seconds between tickles
	and will be set to 1 if it is less than that. Data is any object (best if it's a pointer to 
	something) that will be sent on each tickle request.  The return is the 'id' of the tickle that
	can be used to drop it, and an error if we could not add the tickle spot.

	Add is synchronous so concurrent goroutines which share a common tickler can safely add 
	their tickle spots without worry of corruption.

	Tickle spots should be added in increasing delay order, or the tickler should be stopped until 
	all tickle spots have been added.  This prevents a long tickle spot from becoming active and 
	blocking shorter tickles until the first long tickle 'pops'.
*/
func (t *Tickler) Add_spot( delay int64, ch chan *Chmsg, ttype int, data interface{}, count int ) (id int, err error) {
	var (
		ip	int;			// insert point
		ts	*tickle_spot;
	)

	t.mu.Lock();				// we must be synchronous through the add
	defer t.mu.Unlock();		// unlock on return

	//fmt.Fprintf( os.Stderr, "adding tickle spot delay=%d type=%d count=%d\n", delay, ttype, count )
	id = -1;
	ip = t.tidx;
	if ip < len( t.tlist ) {				// just insert if there is free space
		t.tidx++;
	} else {
		for ip = 0; ip < len( t.tlist ) ; ip++ {
			if t.tlist[ip].ch == nil {
				break;
			}
		}

		if ip >= len( t.tlist ) {
			err = fmt.Errorf( "tickler/Add_spot: no space in the tickle list, cannot add reqest type: %d (%d/%d)\n", ttype, ip, len( t.tlist)  );
			return;
		}
	}

	id = ip;
	if delay < 1 {
		delay = 1;
	}

	ts = &tickle_spot{
		delay: delay,
		ch:	ch,
		req_type:	ttype,
		req_data:	data,
		count: count,
	}
	ts.req = Mk_chmsg( );
	t.tlist[ip] = ts;

	ts.nextgo = time.Now().Unix() + ts.delay;

	if t.ok2run && !t.isrunning {
		t.isrunning = true;				// MUST  bump this here else multiple calls might execute before the go routine initialises and we'll start many 
		go t.tickle_loop( );
	}

	return;
}

/*
	Drop the tickle_spot from the active list.  
*/
func (t *Tickler) Drop_spot( id int ) {
	if id >= 0 && id <= t.tidx {
		t.tlist[id].ch = nil;
	}
}

/*
	Stops the tickler. 
*/
func (t *Tickler) Stop() {
	//fmt.Fprintf( os.Stderr, "stopping tickler\n" );
	t.ok2run = false;
}

/*
	Restarts the tickler.
*/
func (t *Tickler) Start() {

	// TODO: need to searialise this???
	if !t.isrunning {
		t.ok2run = true;
		go t.tickle_loop( );
	}
}
