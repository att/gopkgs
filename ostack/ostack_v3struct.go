// vi: sw=4 ts=4:
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

type osv3_error struct {
	Code		int
	Message		string
	Title		string
}

/*
	Returned by v3/auth/tokens. This is a union of all possible top layer things.
*/
type osv3_token struct {
	Methods		[]string
	Roles		[]*osv3_role
	Projects	*osv3_proj
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
	Error		*osv3_error	
}

