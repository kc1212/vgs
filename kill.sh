#!/usr/bin/env bash

set -e
set -u

function usage {
    echo "usage:"
    echo "./kill.sh (gridsdr|resman) <id>"
}

if [ "$#" -ne 2 ]; then
    echo "incorrect number of arguments"
    usage
    exit 1
fi

p=$1
id=$2
pid=$(ps -e -o pid,cmd | egrep './bin/'"$p"'.*-id\s'"$id" | sed 's/\s\.\/bin\/'"$p"'.*//g')

kill "$pid"


