#!/usr/bin/env ksh

# Abstract:	The go test environment will test some things, but cannot verify output of
# 			messages such as whether or not the baa_some functions supressed the correct
# 			messages, and similar.  This script will execute the test, and scan the output
# 			for things expected and report success or failure based on that and the fail/pass
# 			indication from the go test itself.
# Author:	E. Scott Daniels
# Date:		03 January 2017
# ----------------------------------------------------------------------------------------------

if ! go test >test.out 2>&1
then
	echo "test failed, secondary checks not made  [FAIL]"
	cat test.out
	rm test.out
	exit 1
else
	echo "go test checks passed, making secondary checks"
	awk '
		BEGIN {
			errors = 0
		}

		{ 
			print
		}

		/FAIL/ { next }

		/should show/ {
			show_count++
			next
		}

		/shoud NOT show/ {
			errors++
			printf( "[FAIL] previous message should not have appeareed in the output\n\n" )
			next
		}

		/different time format/ {
			if( split( $2, a, ":" ) < 2 ) {			# second field should just be a timestamp for this
				errors++
				printf( "[FAIL] expected just a timestamp in previous message, not date then time\n\n" )
			}	
			next
		}

		/switching/ {
			switched = 1
			next
		}

		$3 == "little" {			# after switch little timestamp is only time, so id is $3
			if( switched ) {
				printf( "[FAIL] little sheep message found in stderr stream after switch\n\n" )
				errors++
			}
			next
		}

		$4 == "big" {
			if( switched ) {
				printf( "[FAIL] big sheep message found in stderr stream after switch\n\n" )
				errors++
			}
			next
		}

		/baa_some.*:/ {			# the last field should be a multiple of the second value in NF-1
			split( $(NF-1), a, ":" )
			if( ($NF % a[2] ) != 0 ) {
				errors++
				printf( "[FAIL] expected %d to be a multiple of %d in line %d\n\n", $NF, a[2], NR )
			}
			next
		}

		END {
			if( show_count != 6 ) {
				printf( "[FAIL] did not see enough 'should' messages in the output\n\n" )
				errors++
			}

			exit( errors ? 1 : 0 )
		}
	' <test.out
	rc1=$?

	rc2=0
	if [[ ! -f /tmp/bleater_test.log ]]
	then
		echo "[FAIL] could not find the log file created by test: /tmp/bleater_test.log"
		rc2=1
	else
		if ! grep -q big /tmp/bleater_test.log
		then
			echo "[FAIL] expected big sheep messages in /tmp/bleater_test.log"
			rc2=1
		fi
		if ! grep -q little /tmp/bleater_test.log
		then
			echo "[FAIL] expected little sheep messages in /tmp/bleater_test.log"
			rc2=1
		fi
		if  grep -q black /tmp/bleater_test.log
		then
			echo "[FAIL] did NOT expect black sheep messages in /tmp/bleater_test.log"
			rc2=1
		fi
	fi

	if [[ ! -f /tmp/bleater_ls_test.log ]]			# separate little sheep log to ensure we can move it away if needed
	then
		rc2=1
		echo "[FAIL] did not find the second, little sheep only, log /tmp/bleater_ls_test.log"
	else
		if egrep -q "black|big" /tmp/bleater_ls_test.log
		then
			echo "[FAIL] black or big sheep messages found in little sheep only log /tmp/bleater_ls_test.log"
			cat /tmp/bleater_ls_test.log
			rc2=1
		else
			if ! grep -q "little" /tmp/bleater_ls_test.log
			then
				echo "[FAIL] did not find little sheep messages in little sheep only log: /tmp/bleater_ls_test.log"
				cat /tmp/bleater_ls_test.log
				rc2=1
			fi
		fi
	fi

	if (( (rc1 + rc2) > 0 ))
	then
		echo "[FAIL] one or more issues during secondary checks ($rc1, $rc2)"
	else
		echo "[PASS] all secondary checks good"
	fi

	rm -f test.out /tmp/bleater_test.log /tmp/bleater_ls_test.log
	exit $(( (rc1 + rc2) > 0 ))
fi

exit


Expected output on stderr:
483468602 2017/01/03 18:36Z     little [1] should  show (1)
1483468602 2017/01/03 18:36Z     little [2] should  show (2)
1483468602 2017/01/03 18:36Z     little [2] should  show after inc (2)
1483468602 2017/01/03 18:36Z     little [2] should  show after little sheep inc (2)
1483468602 18:36     little [2] should  show with different time format
1483468602 2017/01/03 18:36Z        big [0] switching log file to /tmp/bleater_test.log; all messages should go there now
1483468602 2017/01/03 18:36Z            [0] black sheep should still be writing to stderr
1483468602 2017/01/03 18:36Z            [0] Testing baa_some now
1483468602 2017/01/03 18:36Z            [1] foo baa_some message 1:15 0
1483468602 2017/01/03 18:36Z            [1] bar baa_some message 1:5 0
1483468602 2017/01/03 18:36Z            [1] bar baa_some message 1:5 5
1483468602 2017/01/03 18:36Z            [1] bar baa_some message 1:5 10
1483468602 2017/01/03 18:36Z            [1] foo baa_some message 1:15 15
1483468602 2017/01/03 18:36Z            [1] bar baa_some message 1:5 15
1483468602 2017/01/03 18:36Z            [1] bar baa_some message 1:5 20
1483468602 2017/01/03 18:36Z            [1] bar baa_some message 1:5 25
1483468602 2017/01/03 18:36Z            [1] foo baa_some message 1:15 30
1483468602 2017/01/03 18:36Z            [1] bar baa_some message 1:5 30
1483468602 2017/01/03 18:36Z            [1] bar baa_some message 1:5 35
1483468602 2017/01/03 18:36Z            [1] bar baa_some message 1:5 40
1483468602 2017/01/03 18:36Z            [1] foo baa_some message 1:15 45
1483468602 2017/01/03 18:36Z            [1] bar baa_some message 1:5 45
1483468602 2017/01/03 18:36Z            [1] after reset: foo baa_some message 1:15 0
1483468602 2017/01/03 18:36Z            [1] after reset: bar baa_some message 1:5 0
1483468602 2017/01/03 18:36Z            [1] after reset: bar baa_some message 1:5 5
1483468602 2017/01/03 18:36Z            [1] after reset: bar baa_some message 1:5 10
1483468602 2017/01/03 18:36Z            [1] after reset: foo baa_some message 1:15 15
1483468602 2017/01/03 18:36Z            [1] after reset: bar baa_some message 1:5 15
1483468602 2017/01/03 18:36Z            [1] after reset: bar baa_some message 1:5 20
1483468602 2017/01/03 18:36Z            [1] after reset: bar baa_some message 1:5 25
1483468602 2017/01/03 18:36Z            [1] after reset: foo baa_some message 1:15 30
1483468602 2017/01/03 18:36Z            [1] after reset: bar baa_some message 1:5 30
1483468602 2017/01/03 18:36Z            [1] after reset: bar baa_some message 1:5 35
1483468602 2017/01/03 18:36Z            [1] after reset: bar baa_some message 1:5 40
1483468602 2017/01/03 18:36Z            [1] after reset: foo baa_some message 1:15 45
1483468602 2017/01/03 18:36Z            [1] after reset: bar baa_some message 1:5 45
1483468602 2017/01/03 18:36Z            [0] testing baa some reset
1483468602 2017/01/03 18:36Z            [0] end suppression test (no lines should say 'fail' above.
1483468602 2017/01/03 18:36Z            [1] after reset: foo baa_some message 1:15  (after level reset)
1483468602 2017/01/03 18:36Z            [1] after reset: bar baa_some message 1:5  (after level reset)
1483468602 2017/01/03 18:36Z            [0] two lines should have been written between the end suppression message and this

Expected output from the little sheep only log:
1483472633 19:43     little [0] should go into little sheep log file

