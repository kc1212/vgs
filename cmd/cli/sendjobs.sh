#!/usr/bin/env bash

set -e
set -u

for i in 3000 3001 3002; do
    ./client -addr localhost:$i -count 5 &
done
