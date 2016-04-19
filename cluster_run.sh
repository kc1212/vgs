#!/usr/bin/env bash

set -e
set -u

./bin/discosrv 1>&2 2>"$HOME/tmp/discosrv.log" &

sleep 5

for i in $(seq 5); do
    ./bin/gridsdr -addr "localhost:300$i" -id "$i" 1>&2 2>"$HOME/tmp/gridsdr.$i.log" &
done
