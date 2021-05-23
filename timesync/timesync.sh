#!/bin/bash

function log() {
	echo "$(date) $1"
}

function init() {
	stty -F /dev/ttyS0 sane
	stty -F /dev/ttyS0 115200 -parenb -parodd cs8 -echo

	while true; do
		2>/dev/null chat -t 1 -V '' 'AT' 'OK' > /dev/ttyS0 < /dev/ttyS0 && log "modem replied" && break
		log "no response from modem"
		echo "" > /dev/ttyS0
		sleep 5
	done

	# disable echo, otherwise it will mess up parsing +CCLK? response
	2>/dev/null chat -t 1 -V '' 'ATE0' 'OK' > /dev/ttyS0 < /dev/ttyS0
}

init

while true; do
	log "asking time from modem"
	2>/tmp/clk.file chat -t 3 -V '' 'AT+CCLK?' 'OK' > /dev/ttyS0 < /dev/ttyS0
	timevar="$(cat /tmp/clk.file | grep +CCLK | cut -d' ' -f2)"
	if [ -z "$timevar" ]; then
		log "did not get result"
		sleep 15
		continue
	fi
	log "got result: $timevar"
	# looks like this, including quotation marks "21/04/16,19:36:21+12" # already in local time
	year="20$(echo $timevar | cut -c2-3)"
	month="$(echo $timevar | cut -c5-6)"
	day="$(echo $timevar | cut -c8-9)"
	hour="$(echo $timevar | cut -c11-12)"
	min="$(echo $timevar | cut -c14-15)"
	second="$(echo $timevar | cut -c17-18)"
#	echo "year is $year"
#	echo "month is $month"
#	echo "day is $day"
#	echo "hour is $hour"
#	echo "minute is $min"
#	echo "second is $second"

	date -s "$year-$month-$day $hour:$min:$second" && log "time synchronized: \"$(date)\"" && exit
	sleep 10
done
