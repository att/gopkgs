// vi: sw=4 ts=4:
/*
 ---------------------------------------------------------------------------
   Copyright (c) 2013-2017 AT&T Intellectual Property

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
	Mnemonic:	jcfg.go
	Abstract:	Manage a json config file providing the same Extract_* functions
				as a plain-text config as managed in the cfg_file module. 
				Allows the json config file to be read and parsed, then the user 
				can pull values giving desired sections in the config to search. 
				This is a layer on top of the att/jsontools.

				The basic format of the config is (annoying quotes omitted for clarity):
					{
						value1: 1,
						value2: 2,

						section_a: {
							cvalue: 0
							sa_value1: 11,
							sa_value2: 22
							sa_nested1: {
								san1_v1: 100,
								san1_v2: 200,
								san1_sect_1 {
									xray: true,
									gamaray: false
								}
								san1_sect_2 {
									alpha: "a",
									omega: "z"
								}
							}
						}

						section_b: {
							cvalue: 1
							sb_value1: 111,
							sb_value2: 222
						}
					}

					The values in the 'outer layer' are considered to be in section 'default'.
					When the Mk_jconfig() function is used to build a configuration, the list
					of desired secions is given. This allows those sections to be captured
					and easily referenced without reparsing for each call to extract data. 

					When extracting data, multiple sections can be given, and they will be
					searched in order given.  For example, giving a list of "section_b section_a"
					to an extract function for cvalue, would return 1 as it is in section_b which
					is searched first.   Section hierarchy makes sense when sections are 
					related (e.g. general defaults for all analyser agents, with a section 
					for each agent that overrides those defaults. The analyser sections 
					could be maintained in a nested subsection, but those are more complicated
					to deal with and should be used to help isolate common sections. 

					Nested subsections are extracted as a standalone config using the Get_section()
					function.  The list of 'parent' sections that a subsection is found in
					can be supplied. When extracting a subsection, the nested sections need to 
					be supplied in the same way as sections were supplied to Mk_jconfig().
					To 'extract' the sa_nested1 subsection from section_a, the parent wouold 
					be given as section_a, and the desired san1_sect_n would be given as well:

					sect = cfg.Get_section( "section_a", "sa_nested", "san1_sect_1 san1_sect_2" );

					If sa_nested exists in more than one parent section, and each of those parents
					should be searched, the parent (parm 1) can be a list and they will be searched
					in order until a matching section is found. In the returned struct sect, a 
					'default' section exists which contains the san1_v1 and san1_v2 values should
					values at that level exist. 

					Refer to the test cases for this module for examples of howt to reference 
					a config created here.

	Date:		20 July 2017
	Author:		E. Scott Daniels
*/

package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/att/gopkgs/jsontools"
)

type Jconfig struct {
	cfg_tree *jsontools.Jtree					// main tree
	sects map[string]*jsontools.Jtree			// sections (subobjects) that the user wants to search
}

/*
	Parse a file which we assume is json with a configuration. We generate
	a json tree which can be used to extract information. The sections is
	a space separated list of section names that any get requests will 
	search.

	We assume the conifig is small enough to read into core without issues.
*/
func Mk_jconfig( fname string, sections string ) ( jc *Jconfig, err error ) {
	var (
		rerr error = nil
		rec	string
	)

	f, err := os.Open( fname ) 			// pop the lid on the config
	if err != nil {
		return nil, err
	}

	defer f.Close()

	jblob := ""

	br := bufio.NewReader( f )
	for ; rerr == nil; {
		rec, rerr = br.ReadString( '\n' )	
		if len( rec ) > 0 {
			jblob += rec
		}
	}

	jc = &Jconfig{ }
	jc.cfg_tree, err = jsontools.Json2tree( []byte( jblob ) )
	if err != nil {
		return nil, err
	}

	stokens := strings.Split( sections, " " )
	jc.sects = make( map[string]*jsontools.Jtree, 17 )
	if len( stokens ) > 0 {
		jc.sects["default"] = jc.cfg_tree				// allow fetching of values from outer section
		for _, t := range stokens {
			sect, ok := jc.cfg_tree.Get_subtree( t )	// get the underlying section as requested
			if ok {
				jc.sects[t] = sect						// keep what we find
			}
		}
	} 

	return jc, nil
}

/*
	Given a parent section, find the underlying section and return that as 
	a Jconfig object that can be used on its own. Err is nil on success.
	Parent may be a space separated list searched in order given. Subsections
	are the list of sections under sname that are to be searched. If there 
	are none, the 'default' section can be used to extract values from sname.


	{
		"v1": 1,
		"v2": 2,
		"foo": {
				"fv1": 1,
				"fv2": 2,
				"fs1": {
					"something": 1,
					"else":		2
					"deeper_sect": {
						"dsv1":	10,
						"dsv2": 20
					}
				}
		}
	}

	To get a Jconfig for foo/fs1, assuming that cfg points at the whole tree, call:
		fss := cfg.Get_section( "foo", fs1, "" )  // no subsections

	Then fss can be used to get values from the default section:
		fss.Get_int( "default", "something", 0 )

	If fs1 has subobjects, they can be defined in the sections list on the Get_section()
	call and referenced instead of default.
	
*/
func ( cfg *Jconfig ) Extract_section( parent string, sname string, subsections string ) ( sect *Jconfig, err error ) {
	if cfg != nil && parent != "" {
		ptokens := strings.Split( parent, " " )
	
		for _, p := range ptokens {
			sj := cfg.sects[p]
			if sj != nil {								// sections might not have been defined
				stree, ok := sj.Get_subtree( sname )		// get the name
				if ok {
					sect = &Jconfig{
						cfg_tree: stree,
					}

					sect.sects = make( map[string]*jsontools.Jtree, 17 )
					sect.sects["default"] = stree

					stokens := strings.Split( subsections, " " )
					for _, s := range stokens {
						ss, ok := stree.Get_subtree( s )
						if ok {
							sect.sects[s] = ss
						}
					}

					return sect, nil
				}
			}
		}

		return nil, fmt.Errorf( "unable to find section %s in parent(s): %s", sname, parent );
	}

	return nil, fmt.Errorf( "invalid parameters passed" );
}

/*
	Look up a string in the config and return its value or the 
	default that is provided. Sects is a space separated list of
	sections to search, searched in the order given (e.g. 
	"rico anolis default").
*/
func ( cfg *Jconfig ) Extract_string( sects string, name string, def string ) ( string ) {
	if cfg != nil && sects != "" {
		stokens := strings.Split(sects, " " )
	
		for _, s := range stokens {
			sj := cfg.sects[s]
			if sj != nil {								// sections might not have been defined
				sp := sj.Get_string( name )				// get the name
				if sp != nil {
					return *sp
				}
			}
		}
	}

	return def
}

/*
	Look up a string in the config and return a pointer to the string
	or the default. The default may be either a string or a pointer; 
	if a string is supplied as the default a pointer to that string is
	returned (allows constant string to be supplied without having to 
	have a separate dummy variable created before calling this function.)
*/
func ( cfg *Jconfig ) Extract_stringptr( sects string, name string, def interface{} ) ( *string ) {
	if cfg != nil && sects != "" {
		stokens := strings.Split(sects, " " )
	
		for _, s := range stokens {
			sj := cfg.sects[s]
			if sj != nil {								// sections might not have been defined
				sp := sj.Get_string( name )				// get the name
				if sp != nil {
					return sp
				}
			}
		}
	}

	switch defval := def.(type) {
		case *string:
			return defval

		case string:
			return &defval
		
	}

	return nil
}

/*
	Suss out the value for name and return if it is positive (>= 0) value
	if missing or the first found value in sects is < 0, then the default 
	is returned (def).
*/
func ( cfg *Jconfig ) Extract_posint( sects string, name string, def int ) ( int ) {
	if cfg != nil && sects != "" {
		stokens := strings.Split( sects, " " )
	
		for _, s := range stokens {
			sj := cfg.sects[s]
			if sj != nil {								// sections might not have been defined
				v, ok := sj.Get_int( name )
				if ok {
					if int( v ) > 0 {
						return int( v )
					} else {
						return def			// first found value was <0 so return default
					}
				}
			}
		}
	}

	return def
}

/*
	Suss out the named value as a 64bit integer and return if there; not in the
	config, returns the default.
*/
func ( cfg *Jconfig ) Extract_int64( sects string, name string, def int64 ) ( int64 ) {
	if cfg != nil && sects != "" {
		stokens := strings.Split(sects, " " )
	
		for _, s := range stokens {
			sj := cfg.sects[s]
			if sj != nil {								// sections might not have been defined
				v, ok := sj.Get_int( name )
				if ok {
					return v
				}
			}
		}
	}

	return def
}

/*
	Various int wrapper functions which rely on the previous function.
*/
func ( cfg *Jconfig ) Extract_int( sects string, name string, def int ) ( int ) {
	return int( cfg.Extract_int64( sects, name, int64( def ) ) )
}

func ( cfg *Jconfig ) Extract_int32( sects string, name string, def int32 ) ( int32 ) {
	return int32( cfg.Extract_int64( sects, name, int64( def ) ) )
}

/*
	Suss out the named value as an integer and return if there; not in the
	config, returns the default.
*/
func ( cfg *Jconfig ) Extract_bool( sects string, name string, def bool ) ( bool ) {
	if cfg != nil && sects != "" {
		stokens := strings.Split( sects, " " )
	
		for _, s := range stokens {
			sj := cfg.sects[s]
			if sj != nil {								// sections might not have been defined
				v, ok := sj.Get_bool( name )
				if ok {
					return v
				}
			}
		}
	}

	return def
}

/*
	Write to stderr a dump of the config.
*/
func ( cfg *Jconfig ) Dump( ) {
	if cfg == nil {
		return
	}

	for k, v := range cfg.sects {
		fmt.Fprintf( os.Stderr, "\nSection %q:\n", k )
		v.Dump();
	}

}
