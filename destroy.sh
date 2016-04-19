#!/usr/bin/env bash

set -e
set -u

ps -ef | egrep 'gridsdr|discosrv|resman'

pkill discosrv
pkill gridsdr
pkill resman
