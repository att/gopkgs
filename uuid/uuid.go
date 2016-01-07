// vi: ts=4 sw=4:

/*
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
*/

/*
	Mnemonic:	uuid.go
	Abstract:	UUID generation support. Currently supports generating
				version 4 (random) UUIDs.

				Based on the description in section 4.4 of:
				https://tools.ietf.org/html/rfc4122

	Author:		E. Scott Daniels
	Date:		12 Aug 2015
*/

package uuid

import (
	"crypto/rand"
	"fmt"
)

/*
	    Timelow       med     hi/v    clock   node
                                      hi lo
        0             4       6       8       a
		xx xx xx xx - xx xx - xx xx - xx xx - xx xx xx xx xx xx
                              |       |
                              |       |--- 01.. ....
                              |
                              |--- 0100 ....

	From the RFC:
	The algorithm is as follows:

		o  Set the two most significant bits (bits 6 and 7) of the
			clock_seq_hi_and_reserved to zero and one, respectively.
	
		o  Set the four most significant bits (bits 12 through 15) of the
			time_hi_and_version field to the 4-bit version number from
			Section 4.1.3.
				From 4.1.3: 0 1 0 0		(version 4)

		o  Set all the other bits to randomly (or pseudo-randomly) chosen
			values.
*/

type Uuid struct  {
	buf	[]byte
}
	
/*
	Take a buffer of bytes and return a uuid struct.
*/
func bytes2uuid( buf []byte ) ( *Uuid, error ) {
	
	if len( buf ) != 16 { return nil, fmt.Errorf( "invalid buf size (expected 16)" )	}
	u := &Uuid{ 
		buf: buf[:16],
	}

	return u, nil
}


/*
	Returns a string constructed from the uuid struct.
	Implements stringer.
*/
func (u *Uuid) String( ) ( string ) {
	if u == nil {
		return ""
	}

	return fmt.Sprintf( "%x-%x-%x-%x-%x", u.buf[0:4], u.buf[4:6], u.buf[6:8], u.buf[8:10],  u.buf[10:16] )
}

/*
	Returns a string representing the UUID without any separating dashes.
*/
func (u *Uuid) Plain_string( ) ( string ) {
	return fmt.Sprintf( "%x", u.buf[:16] )
}


/*
	Compares two uuids and returns true if they either represent the same struct
	(both point to same), or if the contents of each are identical.
*/
func (u *Uuid) Equals( u2 *Uuid ) ( bool ) {
	if u == nil && u2 == nil {
		return true
	}


	if u == nil || u2 == nil {
		return false
	}

	if u == u2 {
		return true
	}

	for i := range u.buf {
		if u.buf[i] != u2.buf[i] {
			return false
		}
	}

	return true
}


// ----- specific version functions ------

/*
	Generate a byte buffer with a random uuid.
*/
func Mk_v4( ) ( *Uuid, error ) {
	buf := make( []byte, 16 )

	_, err := rand.Read( buf )			// this is crypto secure
	if err != nil {
		return nil, err
	}

	buf[8] &= 0x3f
	buf[8] |= 0x40	// flip 01 on in byte 8

	buf[6] &= 0x0f
	buf[6] |= 0x40  // flip 0100 on in byte 6

	return bytes2uuid( buf )
}

/*
	Same as calling Mk_v4(); mirrors the name in the old (deprecated)
	code.google uuid code.
*/
func NewRandom() ( u *Uuid ) {
	u, _ =  Mk_v4()
	return u
}

