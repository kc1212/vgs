#!/usr/bin/env bash
set -e
set -u
sed 's/<KEY>/string/g;s/<VAL>/int64/g' concurrent_map_template.txt > cmap_string_int64.go
