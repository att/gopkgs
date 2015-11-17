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
	Mnemonic:	http_logger
	Author:		Robert Eby
	Mods:		10 Aug 2015 - Created.
				17 Nov 2015 - Log query string as well
*/

/*
	This package provides a basic logger to log HTTP requests in the format that will
	be familiar to anyone who has ever used Apache. The logfiles are placed in the
	directory specified when the Http_Logger object is created.  For now only Apache
	"Common Log Format" is supported, although this could be extended in the future.
	The logfiles themselves will always be named access.log.YYYYMMDD, and will be rolled
	daily.
 */
package http_logger

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

/*
	An object that knows how to log requests in Apache format to a logfile.
	For the time being, only Apache "Common Log Format"
	(http://httpd.apache.org/docs/1.3/logs.html#common) is supported.
 */
type Http_Logger struct {
	fmt		string			// for right now, always set to the "common" format
	basenm	string			// for now, always "access.log"
	dir		string			// directory for logfiles
	lastday	time.Time		// day for last log entry
	logfile	*os.File		// if we opened the file, we can close it
}

/*
 Make a Http_Logger object. Specify the directory where logfiles should be placed.
 */
func Mk_Http_Logger(dir *string) ( p *Http_Logger ) {
	if dir == nil {
		s := "/tmp"
		dir = &s
	}
	p = &Http_Logger {
		fmt:	 `%h %l %u %t "%r" %s %b`,
		basenm:  "access.log",
		dir:	 *dir,
		lastday: time.Unix(0, 0),
		logfile: nil,
	}
	return
}

func ( p *Http_Logger ) doRollover() bool {
	now := time.Now()
	a := (now.YearDay() != p.lastday.YearDay())
	b := (now.Year()    != p.lastday.Year())
	return a || b 
}

/*
	Log the HTTP request represented by in to the logfile.  user should be the user who made
	the request, and code and length are the HTTP status code, and length of the response body.
 */
func ( p *Http_Logger ) LogRequest( in *http.Request, user string, code int, length int ) {
// TODO add synchronization
	if p.logfile == nil || p.doRollover() {
		// Time to roll - close old file and open new one.
		now := time.Now()
		fname := fmt.Sprintf( "%s/%s.%4d%02d%02d", p.dir, p.basenm, now.Year(), now.Month(), now.Day() )
		if p.logfile != nil {
			p.logfile.Close()
			p.logfile = nil
		}
		f, err := os.OpenFile( fname, os.O_CREATE|os.O_WRONLY, 0664 )
		if err != nil {
			return
		}
		p.lastday = now
		p.logfile = f
	}

	msg := bytes.NewBufferString("")
	ch  := strings.Split(p.fmt, "")
	for ix := 0; ix < len(ch); ix++ {
		if ch[ix] == "%" && (ix+1) < len(ch) {
			ix++
			switch ch[ix] {
			case "b":
				msg.WriteString(fmt.Sprintf("%d", length))

			case "h":
				addr := in.RemoteAddr
				if addr[0:1] == "[" {
					// Strip port from IPv6 address
					ix := strings.Index( addr, "]" )
					if ix > 0 {
						addr = addr[0:ix+1]
					}
				} else {
					// Strip port from IPv4 address
					ix := strings.Index( addr, ":" )
					if ix > 0 {
						addr = addr[0:ix]
					}
				}
				msg.WriteString(addr)

			case "l":
				msg.WriteString("-")

			case "r":
				url := in.URL.Path
				q := in.URL.RawQuery
				if q != "" {
					url = url + "?" + q
				}
				msg.WriteString(fmt.Sprintf("%s %s %s", in.Method, url, in.Proto))

			case "s":
				msg.WriteString(fmt.Sprintf("%d", code))

			case "t":
				t    := time.Now().UTC()
				mon  := t.Month().String()[0:3]
				date := fmt.Sprintf(`[%02d/%s/%4d:%02d:%02d:%02d -0000]`, t.Day(), mon, t.Year(), t.Hour(), t.Minute(), t.Second())
				msg.WriteString(date)

			case "u":
				msg.WriteString(user)

			default:
				msg.WriteString("%")
				msg.WriteString(ch[ix])
			}
		} else {
			msg.WriteString(ch[ix])
		}
	}
	msg.WriteString("\n")
	if p.logfile != nil {
        p.logfile.Seek( 0, os.SEEK_END )
		fmt.Fprint(p.logfile,  msg.String() )
	} else {
		fmt.Print( msg.String() )
	}
}
