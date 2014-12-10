// vi: sw=4 ts=4:

/*

	Mnemonic:	cfg_file	
	Abstract:	functions that make processing a configfile easier.
	Date:		27 December 2013
	Author:		E. Scott Daniels

	Mods:		30 Nov 2014 - allows for comments on key = value lines.
*/

/*
	Provides the means to open and parse a sectioned configuration file. 
*/
package config

import (
	"bufio"
	//"fmt"
	"io"
	"os"
	"strings"
	//"time"

	"forge.research.att.com/gopkgs/clike"
	"forge.research.att.com/gopkgs/token"
)


// --------------------------------------------------------------------------------------

/*
	Parses a configuration file containing sections and key/value pairs within the 
	sections.  Returns a map of sections (by name) with each entry in the map being
	a map[string]interface{}.  Key/values are converted and stored by the key name as 
	either string pointers or float64s.  If the value is quoted, the quotes are removed. 

	Section names may be duplicated which causes values appearing later in subsequent 
	sections to be added to the previously encountered values.  Keys within each section
	must be uniqueue.  If a duplicate key is encountered, the last one read will be the
	one that ends up in the map.

	If all_str is true, then all values are returned as strings; no attempt is made
	to convert values that seem to be numeric into actual values as it  might make logic
	in the user programme a bit easier (no interface dreferences). 
*/
func Parse( sectmap map[string]map[string]interface{}, fname string, all_str bool ) ( m map[string]map[string]interface{}, err error ) {
	var (
		rec		string;							// record read from input file
		sect	map[string]interface{};			// current section
		sname	string;							// current section name
		rerr	error = nil;					// read error
	)

	if sectmap != nil {
		m = sectmap;
		if m["default"] == nil {				// don't count on user creating a default section
			m["default"] = make( map[string]interface{} )
		}
	} else {
		m = make( map[string]map[string]interface{} );
		m["default"] = make( map[string]interface{} )
	}

	sname = "default";
	sect = m[sname];				// always start in default section

	f, err := os.Open( fname );
	if err != nil {
		return;
	}
	defer f.Close( );

	br := bufio.NewReader( f );
	for ; rerr == nil ; {
		rec, rerr = br.ReadString( '\n' );
		if rerr == nil {
			rec = strings.Trim( rec, " \t\n" );				// ditch lead/trail whitespace

			if len( rec ) == 0 { 					// blank line
				continue;
			}

			switch rec[0] {
				case ':':							// section 
					sname = rec[1:];
					if m[sname]  == nil {
						sect = make( map[string]interface{} );
						m[sname] = sect;
					} else {
						sect = m[sname];
					}

				case '#':							// comment
					// nop

				case '<':
					m, err = Parse( m, rec[1:], all_str );
					if err != nil {
						return;
					}

				default:							// assume key value pair
					ntokens, tokens := token.Tokenise_qpopulated( rec, " \t=" )
					if ntokens >= 2 {				// if key = "value" # [comment],  n will be 3 or more
						key := tokens[0]
						if tokens[1] == "" {		// key = (missing value) given
							tokens[1] = " "
						}
						fc := tokens[1][0:1]
						if ! all_str && ((fc >= "0"  && fc <= "9") || fc == "+" || fc == "-") {		// allowed to convert numbers to float
							sect[key] = clike.Atof( tokens[1] );
						} else {
							dup := ""
							sep := ""
							for i := 1; i < ntokens && tokens[i][0:1] != "#"; i++ {
								dup += sep + tokens[i]									// snarf tokens up to comment reducing white space to 1 blank
								sep = " "
							}
							sect[key] = &dup
						}		
					}								// silently discard token that is just a key, allowing the default to override
			}
		} 
	}

	if rerr != io.EOF {
		err = rerr;
	}

	return;
}

/*
	Parsees the named file keeping the values as strings, and then we convert each 
	section map into a map of strings, allowing for a simpler interface if the 
	user doesn't want us to do the numeric conversions.

	This may be easier for the caller to use as the returned values can be referenced
	with this syntax (assuming m is the returned map):
		m[sect][key]
*/
func Parse2strs( sectmap map[string]map[string]*string, fname string ) ( m map[string]map[string]*string, err error ) {
	var (
		im map[string]map[string]interface{};		// interface map returned by parse
	)

	im, err = Parse( nil, fname, true );
	m = nil;

	if err != nil {
		return;
	}

	if sectmap == nil {									// create a new one if user did not supply, else use theirs and overlay dups from the config
		m = make( map[string]map[string]*string );		// run each section and save them in a string map to return
	} else {
		m = sectmap;
	}

	for sect, smap := range im {
		if m[sect] == nil {							// could have existed in user supplied map
			m[sect] = make( map[string]*string ); 	// make it if not there
		}

		for k, v := range smap {
			m[sect][k] = v.(*string);
		}
	}

	return;
}
