#!/usr/bin/env bash

set -e
set -u

ps -e -o pid,cmd | egrep './bin/(discosrv|resman|gridsdr)' | egrep -v 'grep'
