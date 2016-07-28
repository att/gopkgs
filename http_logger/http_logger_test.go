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

package http_logger_test

import (
	"fmt"
	"github.com/att/gopkgs/http_logger"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestHttp_logger(t *testing.T) {
	// --- Building a sample http request to send to log file ---
	request := http.Request{RemoteAddr: "192.168.0.3",
		Method: "GET",
		URL: &url.URL{
			Path:     "/tegu/mirrors",
			RawQuery: "",
		},
		Proto: "HTTP/1.1",
	}
	msg := ""          // parameter to send to log request
	userid := "pg754h" // Userid to send to log request
	code := 404        // Test parameter for log request

	// --- Specifying a log directory for logs testing ---
	log_dir := "/tmp"

	logger := http_logger.Mk_Http_Logger(&log_dir) // Creating http logger object

	// Sending request and other details to log request to store in logs
	logger.LogRequest(&request, userid, code, len(msg))
	if err := logger.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to close the log file : %s\n", err)
		t.Fail()
	}

	fname := logger.Get_fname() // get filename of log
	now := time.Now()           // get latest time details to form logname

	// Forming a filename format that the log uses
	fname = fmt.Sprintf("%s/%s.%4d%02d%02d", log_dir, fname, now.Year(), now.Month(), now.Day())

	// Reading the contents of the file
	_, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read the file contents : %s\n", err)
		t.Fail()
	}

	// Change the filename of the log to the user specified name
	chg_fname := "changed_access_log"
	if err = logger.Set_fname(chg_fname); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to rename the log file : %s\n", err)
		t.Fail()
	}
}
