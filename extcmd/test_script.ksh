#!/usr/bin/env ksh
#
# test script called by the test go function(s) 
# generate multiple lines of output with silly stuff on them.


echo " $SHELL"
echo "test script output"
i=0
for x in "$@"
do
	echo "[$i] $x"
	(( i++ ))
done

echo "this is written to stderr" >&2
echo "as is this is written to stderr" >&2
echo "and finally this is written to stderr" >&2
