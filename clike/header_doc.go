// this doc is picked up and applied at the package level, makes sense to have it
// here rather than in one of the other files. 


/*
The clike package provides some functions that behave more like the traditional Clib
functions than functions supported by the Go packages.  Currently these all revolve
around the atoX() family of Clib functions which do not error when they encounter a
non-digit, and return a 0 if the leading character in the string is not numeric or a
sign.  These Clike functions all stop parsing when the first non digit is encountered.

A small extension has been added which allows the string representation to contain a
trailing suffix. The suffix causes the function to expand the value. For example the
suffix M causes the value to be expanded to 1000000 times the string value.  The following
suffixes are supported:

	MiB or m	power of 2
	GiB or g	power of 2
	KiB or k	power of 2
	M			power of 10
	G			power of 10
	K			power of 10


The intger based functions all accept a leading '0x' or '0' to indicate that the string
should be interpreted as a hex or octal value respectively.

The functions all accept either a string or an array of bytes ([]byte). 

*/
package clike
