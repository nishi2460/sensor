#!/bin/bash

TARGET=${WAIT_CLOCK_TARGET:-"/var/lib/systemd/timesync/clock"}
TIMEOUT=${WAIT_TIMEOUT:-60}

CHANGE=0
LAST_TS=`stat -c %Y $TARGET`

for i in `seq 1 $TIMEOUT` ; do
    TS=`stat -c %Y $TARGET`
    if [ $LAST_TS != $TS ]; then
        CHANGE=1
        echo "timesyncd clock updated [$i sec waited]"
        break
    fi
    sleep 1
done

if [ $CHANGE == 0 ]; then
    echo "timeout watch timesyncd clock update"
fi


while true
do
  ~/Z_Work/sensor/sensor
done


