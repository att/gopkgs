package uuid_test

import (
	"fmt"
	"testing"

	"github.com/att/gopkgs/uuid"
)


func TestUuid4( t *testing.T ) {
	u, err := uuid.Mk_v4()

	if u == nil {
		t.Fail()
		fmt.Printf( "u was nil: %s\n", err )
	}
	//u2, _ := uuid.Mk_v4()
	u2 := uuid.NewRandom()
	
	fmt.Printf( "uuid =  %s plain = %s\n", u, u.Plain_string() )
	fmt.Printf( "uuid2 = %s\n", u2 )
	fmt.Printf( "uuid == uuid2? %v (epect true)\n", u.Equals( u ) )
	fmt.Printf( "uuid == uuid2? %v (epect false)\n", u.Equals( u2 ) )
}
