// vi: sw=4 ts=4:

/*

	Mnemonic:	chkpt

	Date:		4 December 2013
	Author:		E. Scott Daniels

	Mods:		03 Feb 2015 - Fix straggling % in sprintf format, remove
					unneeded ;'s.
*/

/*
				The chkpt package implements a checkpoint file manager.
				The checkpoint scheme is based on a two tumbler (a and b) system
				where the b tumber cyles with each call, and the a tumber cycles 
				once with each roll-over of b.  If the roll point of b is 24 and 
				the roll point of a is three, checkpoint files are created in this
				order:
					a1, b1, b2, b3... b24, a2, b1... b24, a2...

				If the using programme is creating a checkpoint every hour, then 
				an 'a' file is created once per day with a coverage of three days; 
				at the end of the third cycle there will exist a1 (2 days old), 
				a2 (one day old) and a3 new, along with 24 'b' files. 

				As the checkpoint is written, the mdd5sum is computed and if desired
				upon close the value is added to the name such that final names have 
				the form:
						<path>_Xnn-<md5>.ckpt

				This is enabled by invoking the Add_md5() function after the checkpoint
				object is created.

				The file <path>.ckpt (no tumbler) is used to record the tumbler
				values such that the next time the programme is started it will
				create the next file in a continued sequence rather than resetting
				or requiring the user programme to manage the tumbler data. 

				chkpt implements the io.Writer interface such that it is possible
				to use the pointer to the object in an fmt.Fprintf() call:
					c = Mk_chkpt( "/usr2/ckpts/agama", 5, 25 )
					c.Create()
					fmt.Fprintf( c, "%s\n", data )
					c.Close()

				The method Write_string( string ) can also be used to write to 
				an open checkpoint file.
*/
package chkpt

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"os"

	"codecloud.web.att.com/gopkgs/clike"
	"codecloud.web.att.com/gopkgs/token"
)

// --------------------------------------------------------------------------------------

/*
	Defines a checkpoint environment.
*/
type Chkpt struct {
	path	*string			// path where chkpt files are saved: [/path/path/]fname-prefix
	aval	int				// current tumbler values
	bval	int
	amax	int				// rollover points for each tumbler
	bmax	int
	errors	bool				// errors during write -- reported on close
	output	*os.File			// open ckpt file
	bw		*bufio.Writer		// wrapping writer to file
	br		*bufio.Reader		// wrapping reader
	md5hash	hash.Hash			// we compute the md5 on write and add to the name
	open_name	*string		// name of the file that is being written to
	addmd5	bool				// true if we should add the md5 value to the file name (prevents overlay)
}

// --------------- private -----------------------------------------------------------------

/*
	Look into the tumbler counter file and get the two counters out.
*/
func (c *Chkpt) read_tumblers( ) (aval int, bval int, err error) {
	var (
		buffer	[]byte
	)

	fname := fmt.Sprintf( "%s.ckpt", *c.path )
	f, err := os.Open( fname )

	aval = c.amax					// default to starting such that we roll both over on first write
	bval = c.bmax
	if err != nil {
		return
	}
	defer f.Close()

	buffer = make( []byte, 1024 )
	nread, err := f.Read( buffer )
	if nread <= 0 {
		err = fmt.Errorf( "empty tumbler info file: %s", fname )
		return
	}

	ntokens, tokens := token.Tokenise_populated(  string( buffer ), ", " )

	if ntokens < 2 {
		err = fmt.Errorf( "missing info from tumbler file: %s", fname )
		return
	}

	aval = clike.Atoi( tokens[0] )
	bval = clike.Atoi( tokens[1] )
	err = nil

	return
}

/*
	Save the current tumbler values into the counter file.
	We write the current values to a temp file and close it. If successful
	then we move it over the existing file. 
*/
func (c *Chkpt) save_tumblers( ) (error) {

	old_fname := fmt.Sprintf( "%s.ckpt", *c.path )
	new_fname := fmt.Sprintf( "%s", old_fname )

	f, err := os.Create( new_fname )
	if err != nil {
		return err
	}
	// do NOT defer f.close() because we must close it before exit to allow for rename

	s := fmt.Sprintf( "%d %d\n", c.aval, c.bval )
	b := []byte( s )
	_, err = f.Write( b )
	if err != nil {
		f.Close( )
		return err
	}
	f.Close( )

	err = os.Rename( new_fname, old_fname )
	return err
}

// --------------- public -----------------------------------------------------------------

/*
	Mk_chkpt creates a checkpointing environmnent. Path is any leading directory path, with 
	a basename (prefix) mask (e.g. /usr2/backups/foo).  The tumbler information is added to 
	path in the form of either a_xxxx or b_xxxx where xxx is the tumbler number.  
*/
func Mk_chkpt( path string, amax int, bmax int ) (c *Chkpt) {

	c = &Chkpt { 
		amax:	amax,
		bmax:	bmax,
		path:	&path,
	}

	c.aval, c.bval, _ = c.read_tumblers( )

	if c.aval > c.amax { 
		c.aval = 0
	}
	if c.bval > c.bmax { 
		c.bval = 0
	}

	return
}

/*
	Sets the flag that will cause the final file name to include the MD5 value computed for
	the file.  Adding the md5 breaks the 'overlay' nature of the tumbler system.
*/
func (c *Chkpt) Add_md5( ) {
	c.addmd5 = true
}

/*
	Write_string writes the string to the open chkpt file. 
*/
func (c *Chkpt) Write_string( s string ) (n int, err error ) {
	if c.md5hash != nil {
		io.WriteString( c.md5hash, s )
	}

	n, err = io.WriteString( c.output, s )
	if err != nil {
		c.errors = true
	}
	return
}

/*
	Write allows the Go fmt package to be used to easily write a formatted string to the 
	open checkpoint file. (see the example in the package description for how the fmt
	package can be used.
*/
func (c *Chkpt) Write( b []byte ) (n int, err error ) {

	if c.output == nil {			// silently ignore attempt to write before it's open
		return
	}

	if c.md5hash != nil {
		io.WriteString( c.md5hash, string( b ) )
	}

	n, err = c.output.Write( b )
	if err != nil {
		c.errors = true
	}
	return
}

/*
	Closes the open checkpoint file. The final file is renamed with the md5sum (if no write errors).
	The final filename and any error status are reported to the caller. 
	Success is reported (err == nil) only when no write errors and a successful close/rename.
*/
func (c *Chkpt) Close( ) (final_name string, err error) {


	if c.output == nil {
		return "", nil
	}

	if c.bw != nil {		// flush if writing
		c.bw.Flush( )
	} 

	err = c.output.Close( )

	c.bw = nil
	c.br = nil
	c.output = nil

	final_name = ""
	if err == nil && c.md5hash != nil {
		md5 := c.md5hash.Sum( nil )

		c.md5hash = nil

		if c.open_name != nil {
			if c.addmd5 {
				final_name = fmt.Sprintf( "%s-%x.ckpt", *c.open_name, md5 )
				err = os.Rename( *c.open_name + ".ckpt", final_name )
			} else {
				final_name = *c.open_name
			}
		}
	}

	c.open_name = nil

	if err == nil && c.errors {			// no close error, but write errors, fail the close
		err = fmt.Errorf( "close successful, but write errors detected; file may be corrupt: %s", final_name )
	}

	return
}

/*
	The create method caues a new checkpoint file to be created and opened. Once
	opened, thw Write or Write_string methods can be called to write to the the file.
	The Close method should be invoked when all writing has occurred.
*/
func (c *Chkpt) Create( )  ( error ) {
	var (
		err		error
		tval	int				// tumbler value for filename
		tch		string = "b"	// tumbler letter for filename; likely it will be 'b'
	)
	
	if c.output != nil {		// user didn't close last one
		c.Close( )				// we'll ignore errors -- little we can do if they didn't drive the close
	}

	c.errors = false
	c.bval++					// inc the tumbler(s) and set the ch/value pair for name
	if c.bval > c.bmax {
		c.bval = 0
		tch = "a"
		c.aval++
		if c.aval > c.amax {
			c.aval = 1
		}

		tval = c.aval
	} else {
		tval = c.bval
	}

	fname := fmt.Sprintf( "%s_%s%d.ckpt", *c.path, tch, tval )
	c.output, err = os.Create( fname )
	if err != nil {
		return err
	}

	on := fmt.Sprintf( "%s_%s%d", *c.path, tch, tval )
	c.open_name = &on

	c.save_tumblers( )

	c.bw = bufio.NewWriter( c.output )
	c.md5hash = md5.New()

	return nil
}

