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

/*
	Mnemonic:	cert
	Abstract: 	Support for generating self-signed certificates
	Date:		06 June 2014
	Author: 	E. Scott Daniels
*/

package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

/*
	Create a certificate (self signed) and write to the output file.
*/
func Mk_cert( key_bits int, cert_name *string, dns_list []string, cfname *string, kfname *string ) ( err error ) {

	rsa_key, err := rsa.GenerateKey( rand.Reader, key_bits )
	if err != nil {
		return
	}

	now := time.Now()

	// per golang doc, only these fields from the cert are used as the template:
	//	SerialNumber, Subject, NotBefore, NotAfter, KeyUsage, ExtKeyUsage, UnknownExtKeyUsage, BasicConstraintsValid, IsCA, MaxPathLen,
	//	SubjectKeyId, DNSNames, PermittedDNSDomainsCritical, PermittedDNSDomains.
	template := x509.Certificate {
		SerialNumber:	new( big.Int ).SetInt64( 1 ),
		Subject: pkix.Name {
			CommonName:   *cert_name,
			Organization: []string{"ATT_Labs_Research"},
		},
		NotBefore:    now,
		NotAfter:     now.Add( 86400 * 365 * 2 * time.Second ),
		IsCA:         false,
		DNSNames:	dns_list,
		MaxPathLen:	0,
		SubjectKeyId: []byte{ 1, 2, 3, 4 },
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	cert_bytes, err := x509.CreateCertificate( rand.Reader, &template, &template, &rsa_key.PublicKey, rsa_key )
	if err != nil {
		return
	}

	fd, err := os.Create( *cfname )
	if err != nil {
		return
	}

	pem.Encode( fd, &pem.Block{ Type: "CERTIFICATE", Bytes: cert_bytes} )
	fd.Close( )

	fd, err = os.OpenFile( *kfname, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0600 )
	if err != nil {
		return
	}

	pem.Encode( fd, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey( rsa_key )})
	fd.Close()


	return
}
