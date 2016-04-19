#!/usr/bin/env bash

set -e
set -u

function usage {
    echo "usage:"
    echo "./run.sh <gs count> <rm count>"
}

if [ "$#" -ne 2 ]; then
    echo "incorrect number of arguments"
    usage
    exit 1
fi

gs_start=3001
rm_start=3101

gs_end=$(printf "30%02d" "$1")
rm_end=$(printf "31%02d" "$2")

# first kill everything else
./destroy.sh

# start discosrv and wait a bit
./bin/discosrv 1>&2 2>"$HOME/tmp/discosrv.log" &
sleep 1

# start the GSs
for i in $(seq "$gs_start" "$gs_end"); do
    ./bin/gridsdr -addr "localhost:$i" -id "$i" 1>&2 2>"$HOME/tmp/gridsdr.$i.log" &
done

# start the RMs
for i in $(seq "$rm_start" "$rm_end"); do
    ./bin/resman -addr "localhost:$i" -id "$i" 1>&2 2>"$HOME/tmp/resman.$i.log" &
done
