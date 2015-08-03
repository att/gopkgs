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

package ostack

/*
	Mnemonic:	ostack_v3struct
	Abstract:	Structures related to v3 api calls
	Date:		03 April 2015
	Author:		E. Scott Daniels
*/


type osv3_domain struct {
	Id		string
	Name	string
}

type  osv3_proj struct {
	Domain	*osv3_domain
	Id		string
	Name	string
}

type osv3_role struct {
	Id		string
	Name	string
}

type osv3_endpoint struct {
	Url		string
	Region	string
	Legacy_endpoint_id	string
	Interface string
	Id		string
}

type osv3_catentry struct {
	Endpoints	[]*osv3_endpoint
	Type		string					// network compute ec2 etc.
	Id			string
}

type osv3_user struct {
	//Domain							// who knows what this is
	Id			string
	Name		string
}

/*
	Returned by v3/auth/tokens. This is a union of all possible top layer things.
*/
type osv3_token struct {
	Methods		[]string
	Roles		[]*osv3_role
	Project		*osv3_proj
	Expires_at	string					// human readable expiry
	//Extras	unknown
	User		*osv3_user
	Issued_at	string					// token issue date
		
}

/*
	A generic list of things that might come back from ostack
*/
type osv3_generic struct {
	Token		*osv3_token
	Project		*osv3_proj
	Error		*error_obj				// we can use the generic error handling things here
}

