#!/usr/bin/env bash

set -e
set -u

function usage {
    echo "usage:"
    echo "./monitor.sh <gridsdr|resman> <id>"
}

if [ "$#" -ne 2 ]; then
    echo "incorrect number of arguments"
    usage
    exit 1
fi

p=$1
id=$2
tail -f "$HOME/tmp/$p.$id.log"
