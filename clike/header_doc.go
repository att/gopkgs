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


The integer based functions all accept a leading '0x' or '0' to indicate that the string
should be interpreted as a hex or octal value respectively.

The functions all accept either a string or an array of bytes ([]byte).

*/
package clike
