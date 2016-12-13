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
 Mnemonic:	tls.go
 Abstract:	Functions which support a TLS listener within a connman environment.

 Date:		30 October 2016
 Author: 	E. Scott Daniels

 Mods:
*/

package connman

import (
	"fmt"
	"strings"
	"os"
	"crypto/tls"

	"github.com/att/gopkgs/security"
)

/*
	Generate a self-singed certificate and key which are plced into the files
	with the given names passed in.
*/
func mk_cert( cert_name string, cfname *string, kfname *string ) ( err error ) {
	this_host, _ := os.Hostname( )
	tokens := strings.Split( this_host, "." )

	dns_list := make( []string, 3 )
	dns_list[0] = "localhost"
	dns_list[1] = this_host
	dns_list[2] = tokens[0]

	err = security.Mk_cert( 1024, &cert_name, dns_list, cfname, kfname )

	return err
}

/*
	Given a cert and key filename pair, load them and return a cert that can be
	used in a tls configuration struct.
*/
func load_cert( cfname *string, kfname *string ) ( cert tls.Certificate, err error ) {
	if cfname == nil || kfname == nil {
		return cert, fmt.Errorf( "one of both filename pointers were nil" )
	}

	cert, err = tls.LoadX509KeyPair( *cfname, *kfname )
	return cert, err
}

/*
	Create a base tls config to give to tls.Listen() etc.
*/
func mk_tls_config( certs []tls.Certificate ) ( *tls.Config ) {
	base_config := &tls.Config {
		Rand: nil,					// TLS will use crypto/rand
		Time: nil, 					// now is used
		NameToCertificate: nil, 	// cause 1st cert to always be used
		GetCertificate: nil,		// we don't supply a 'lookup' func
		RootCAs: nil,				// use host os's set
		NextProtos: nil,
		ClientAuth: tls.NoClientCert,
		InsecureSkipVerify: true,
		CipherSuites: nil,			// use implementation available suites
		PreferServerCipherSuites: false,
		SessionTicketsDisabled: true,	// don't allow session tickets
		//SessionTicketKey (allow to be zeros)
		MinVersion: 0,
		MaxVersion: 0,
		//CurvePreferences (allow to be zeros)	
		//Renegotiation: tls.RenegotiateNever,
	}
	
	
	base_config.ServerName, _ = os.Hostname()
	base_config.Certificates = certs
	return base_config
}

/*
	Starts a listener (TCP or UDP) for tls connections on the indicated port and interface. The
	data2usr channel is used to pass back connections when accepted in the same manner as is 
	done by the Listen() function in this package. The actual listener is started in a parallel
	goroutine which passes back accept information on the channel; this function returns and 
	the caller does not need to invoke as a goroutine.

	The cert_base parameter is the filename path (e.g. /usr2/foo/bar/progx) to which '.key' and '.cert'
	will be appended to create the cert and key filenames.  If these files do not exist, then a
	self-signed certificate and key will be created and written to them.

	The listener ID string is returned (allowing the listener to be cancelled) along with an error
	object if there was a failure.
*/
func (this *Cmgr) TLS_listen( kind string, port string,  iface string, data2usr chan *Sess_data, cert_base string ) ( lid string, err error ) {
	if this == nil  {
		return "", nil
	}

	if port == ""  || port == "0" {		// user probably called constructor not wanting a listener
		return "", nil
	}
	if cert_base == "" {
		return "", fmt.Errorf( "must supply a basename string that is non-empty" )
	}

	kfname := fmt.Sprintf( "%s.key", cert_base )
	cfname := fmt.Sprintf( "%s.cert", cert_base )

	have_count := 0
	_, err = os.Stat( kfname ) 				// check files; both or neither must exist
	if err == nil {
		have_count++
	}
	_, err = os.Stat( cfname )
	if err == nil {
		have_count++
	}

	switch have_count {
		case 0:											// both missing, create into the filenames given
			toks := strings.Split( cert_base, "/" )
			cert_name := toks[len(toks)-1]
			err = mk_cert( cert_name, &cfname, &kfname )
			if err != nil {
				return "", fmt.Errorf( "unable to make certificate: %s", err )
			}
	
		case 1:						// found one, but not both -- don't know what to do so abort
			return "", fmt.Errorf( "found one of the key/cert pair, but not both. must have both or neither" )
	}
		
	certs := make( []tls.Certificate, 1 )
	certs[0], err = load_cert( &cfname, &kfname )
	if err != nil {
		return "", err
	}

	config := mk_tls_config( certs )		// create a configuration
	
	lid = ""
	l, err := tls.Listen( kind, fmt.Sprintf( "%s:%s", iface, port ), config )
	if err != nil {
		err = fmt.Errorf( "unable to create a TLS listener on port: %s; %s", port, err )
		return
	}

	lid = fmt.Sprintf( "l%d", this.lcount )
	this.lcount += 1

	this.llist[lid] = l
	go this.listener(  this.llist[lid], data2usr )
	return
}
