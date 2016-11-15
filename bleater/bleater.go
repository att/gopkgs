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

	Mnemonic:	bleater

	Date:		27 December 2013
	Author:		E. Scott Daniels

	Mods:		30 Apr 2014 (sd) : Added ability to change the target.
				08 Nov 2016 (sd) : Add ability to push existing target down.
				15 Nov 2016 (sd) : Add nil checks to functions without.
*/


/*
	Bleater: Something that goes baaa in the night.

	Supports a a multi-tiered (parent with multiple children) bleat object
	which allows for a 'master level' control through the parent which affects
	all children, and individual control which affects only the child.  See
	the test module for an example.

	When the Baa() function is called, the message is written only if the
	indicaed level is <= the current level in the bleater, or <= than the
	parent level.  When a parent's level is changed, it is "broadcast" to
	all children in an attempt to minimise the cycles needed to check for
	each bleat (the assumption is that the parent level seldomly changes
	and pushing it is less expensive than constantly checking the parent
	object's value).

	Each bleat message written is prefixed with the current unix timestamp,
	a human readable timestamp, the bleater prefix (if given), the level
	number in square brackets, and the formatted user message passed in printf()
	style on the Baa() call.  The default human readable timestamp is of
	the form YYYY/MM/DD HH:MM:SSZ; use the Set_tsformat() function to supply
	a 'mask' if a different layout is desired.  Masks are as described in
	the Golang time package.  Bleat messages are automatically terminated
	with a newline, so including one in the message is not needed.
*/
package bleater

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)


type Bleater struct {
	mtx		sync.Mutex
	target	io.Writer;		// where stuff is written (e.g. os.Stderr)
	tfile	*os.File;		// if we opened the file, we can close it
	level	uint;			// current 'volume'
	plevel	uint;			// parent level -- should make things faster
	children []*Bleater;	// if this bleater has children, we'll push our level when it changes
	cidx	int
	pfx		string
	tsfmt	string
	bleat_some	map[string]int	// counter for each bleat_some class
}

// --------------- private -------------------------------------------------------------------------------------

/*
	Set our copy of the parent's level. Called only by parent.
*/
func ( b *Bleater ) set_plevel( l uint ) {
	if b == nil {
		return
	}

	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.plevel = l
}

// --------------- public -------------------------------------------------------------------------------------

/*
	Mk_bleater creates a new bleater object with a few initial settings.
*/
func Mk_bleater( level uint, target io.Writer ) ( b *Bleater ) {
	b  = &Bleater {
		target: target,
		level: 	level,
		plevel:	0,
		cidx: 0,
		tsfmt: "2006/01/02 15:04Z",
		tfile: nil,
	}

	b.bleat_some = make( map[string]int, 10 )
	b.children = make( []*Bleater, 10 )
	return
}

/*
	Baa causes the message to be written to the output device if the current bleating level is
	greater or equal to the message level (when). If the when value is greater than the
	current level, the message is suppressed.
	
*/
func ( b *Bleater ) Baa( when uint, uformat string, va ...interface{} ) {
	if b == nil {
		return
	}

	if when <= b.level || when <= b.plevel {	// bleat when our level is set high or when parent (master) level is set high
		b.mtx.Lock(); 							// yes we check before the lock so the call isn't expensive when level is low
		defer b.mtx.Unlock()

		n := time.Now()
		//fmt.Fprintf( b.target, "%d %s %10s [%d] %s\n", n.Unix(), n.Format( b.tsfmt ), b.pfx, when, fmt.Sprintf( uformat, va... ) )
		fmt.Fprintf( b.target, "%d %s %10s [%d] %s\n", n.Unix(), n.UTC().Format( b.tsfmt ), b.pfx, when, fmt.Sprintf( uformat, va... ) )
	}
}

/*
	Allows a caller to bleat messages belonging to a 'class' less often than every time called.
	The Baa_some function accepts additional parameters (class name, and frequency) and will
	bleat the message on the first call, and then every frequency of calls there after.
	Frequency is not saved with the class, so it is possible to change the frequency at any time.

	The class counter is incremented only if the message would otherwise be written with respect
	to the value of when.  Thus, a class poised to write a message on the next invocation
	will write that message as soon as the level is appropriate, and does not run the risk of
	always skipping if levels fluxuate.
*/
func ( b *Bleater ) Baa_some( class string, freq int, when uint, uformat string, va ...interface{} ) {
	if b == nil {
		return
	}
	if !( when <= b.level || when <= b.plevel) {		// wouldn't bleat, don't bump the counter
		return
	}

	c, ok := b.bleat_some[class]
	if ! ok || c >= freq {						// c could be > if freq was lowered
		b.bleat_some[class] = 1
		b.Baa( when, uformat, va... )			// write if level is ok
	} else {
		b.bleat_some[class]++
	}
}

/*
	Allow a bleat_some class to be reset such that the next bleat_some() will
	cause the message to be written and the counter cycled.
*/
func ( b *Bleater ) Baa_some_reset( class string ) {
	if b == nil || b.bleat_some == nil {
		return
	}
	
	delete( b.bleat_some, class )
}

/*
	Add_child allows a bleater object to be added to the given object as a child.
	Managing bleaters in a parent child organisation allows a 'master bleating volume'
	to be managed in the parent, while allowing the volume for a particular subsystem
	(child) to be set differently (louder) than the rest.
*/
func ( b *Bleater ) Add_child( cb *Bleater ) {
	if b == nil {
		return
	}

	b.mtx.Lock()
	defer b.mtx.Unlock()

	if b.cidx >= len( b.children ) {
		nc := make( []*Bleater, b.cidx + 10 )
		copy( nc, b.children )
		b.children = nc;			
	}

	b.children[b.cidx] = cb
	b.cidx++;	
	cb.set_plevel( b.level )
	cb.Set_target( b.target, true )
}

/*
	Set_tsformat allows the timestamp format that is written on a bleat message to be changed.
	Any format string that is supported by the Go time package may be used.
*/
func ( b *Bleater ) Set_tsformat( fmt string ) {
	if b == nil {
		return
	}

	b.tsfmt = fmt
}

/*
	Set_target changes the output target and pushes the target to children.
	If close_old is set, the old target is closed (if possible), otherwise it is left
	alone.  (it is not possible to close standard error, or any target that was
	not opened by al call to bleater.Create_target(). If new_target is nil, this 
	just pushes the current target down.
*/
func ( b *Bleater ) Set_target( new_target io.Writer, close_old bool ) {
	if b == nil {
		return
	}


	b.mtx.Lock()
	defer b.mtx.Unlock()

	if new_target != nil {
		if close_old  && b.tfile != nil {
			b.tfile.Close( )
		}
		b.tfile = nil

		b.target = new_target
	}

	for i := 0; i < b.cidx; i++ {
		b.children[i].Set_target( new_target, false )			// propigate the target, but we might've closed it so they shouldn't
	}
}

/*
	Creates a new file and truncates it. All subsequent Baa() calls will write to this file.
*/
func ( b *Bleater ) Create_target( target_fname string, close_old bool ) ( err error ) {
	if b == nil {
		return
	}

	f, err := os.Create( target_fname )
	if err != nil {
		return
	}

	b.Set_target( f, close_old )			// push to all child bleaters

	b.mtx.Lock()
	b.tfile = f								// now safe to capture this
	b.mtx.Unlock()

	return
}

/*
	Opens the target file and appends to it. If the file doesn't exist, it creates it.
*/
func ( b *Bleater ) Append_target( target_fname string, close_old bool ) ( err error ) {
	if b == nil {
		return
	}

	f, err := os.OpenFile( target_fname, os.O_CREATE|os.O_WRONLY, 0664 )
	if err != nil {
		return
	}

	f.Seek( 0, os.SEEK_END )				// to end of file before we write anything
	b.Set_target( f, close_old )			// push to all children

	b.mtx.Lock()
	b.tfile = f								// now safe to capture this
	b.mtx.Unlock()

	return
}

/*
	Set_level changes the volume for the object, and pushes it to any child bleaters.
*/
func ( b *Bleater ) Set_level( l uint ) {
	if b == nil {
		return
	}

	if l < 0 {
		l = 0
	}

	b.mtx.Lock()
	defer b.mtx.Unlock()

	b.level = l
	for i := 0; i < b.cidx; i++ {
		b.children[i].set_plevel( l )
	}
}
	
/*
	Get_level returns the bleater's current level.
*/
func ( b *Bleater ) Get_level( ) ( uint ) {
	if b == nil {
		return 0
	}

	return b.level;					// yes, we'll risk not locking here too
}

/*
	Would_baa will return true if the Baa method were invokde for the given level.
	This might be advantageous if the information that is to be bleated is fairly
	expensive to compute, and the application wishes to avoid the computation unless
	it is sure that it will be written.
*/
func ( b *Bleater ) Would_baa( lvl uint ) ( bool ) {
	if b == nil {
		return false
	}

	return lvl <= b.level || lvl <= b.plevel
}

/*
	Set_prefix establishes the prefix for this bleater. The prefix is the portion
	of the message that is written after the timestamp.
*/
func ( b *Bleater ) Set_prefix( pfx string ) {
	if b == nil {
		return
	}

	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.pfx = pfx
}

/*
	Inc_level is a convenient way to increase the volume by one.
*/
func ( b *Bleater ) Inc_level( ) {
	if b == nil {
		return
	}

	b.level++
}

/*
	Dec_level is a convenient way to decrease the volume by one.
*/
func ( b *Bleater ) Dec_level( ) {
	if b == nil {
		return
	}

	if b.level > 0 {
		b.level--
	}
}

