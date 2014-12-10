/*
	The ipc pagage provides several objects which make managing inter
	processes communications easier.  The ipc package is built on top
	of Go's channels. Within the package is a generic message which 
	provides request and response support and a tickler object which 
	allows an application to schedule periodic messages to be delivered
	on one or more channels. 
*/
package ipc
