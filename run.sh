#!/usr/bin/env bash

set -e

function usage {
    echo "usage:"
    echo "./run.sh <gs count> <rm count> <worker count> [<my addr>] [<discosrv addr:port>]"
}

if [ "$#" -lt 3 ]; then
    echo "incorrect number of arguments"
    usage
    exit 1
fi

gs_start=3001
rm_start=3101

# set the number of GS and RM and Worker nodes
gs_end=$(printf "30%02d" "$1")
rm_end=$(printf "31%02d" "$2")
worker_count="$3"

# set my address
# alternatively we can set it automatically by using the following
# dig +short myip.opendns.com @resolver1.opendns.com
my_addr=localhost
if [ -n "$4" ]; then
    my_addr="$4"
fi

# set discovery server address
discosrv_addr="localhost:3333"
if [ -n "$5" ]; then
    discosrv_addr="$5"
else
    # start discosrv if we're using the default address
    ./bin/discosrv 1>&2 2>"$HOME/tmp/discosrv.log" &
    usleep 500000
fi

# start the GSs
for i in $(seq "$gs_start" "$gs_end"); do
    ./bin/gridsdr -addr "$my_addr:$i" -id "$RANDOM" -discosrv "$discosrv_addr" 1>&2 2>"$HOME/tmp/gridsdr.$i.log" &
    usleep 200000
done

# start the RMs
for i in $(seq "$rm_start" "$rm_end"); do
    ./bin/resman -addr "$my_addr:$i" -id "$RANDOM" -discosrv "$discosrv_addr" -nodes "$worker_count" 1>&2 2>"$HOME/tmp/resman.$i.log" &
    usleep 200000
done
