// vi: sw=4 ts=4:

package chkpt_test

import (
	"fmt"
	"time"
	"testing"

	"forge.research.att.com/gopkgs/chkpt"
)

func TestChkpt( t *testing.T ) {
	c := chkpt.Mk_chkpt(  "/tmp/foo", 3, 5 );
	c.Add_md5()

	if c == nil {
		t.Errorf( "creation of chkpt obj failed" );
		return;
	}

	err := c.Create( );

	if err != nil {
		t.Errorf( "open file for write failed: %s", err );
		return
	}

	fmt.Fprintf( c, "this is a test! %d\n", time.Now().Unix() );		// Chkpt can be used directly in a format statement if open
	fname, err := c.Close( );

	if err != nil {
		fmt.Printf( "error reported by close: %s\n", err )
		t.Fail()
	}
	fmt.Printf( "output chkpt to: %s\n", fname )
}
