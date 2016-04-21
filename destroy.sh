#!/usr/bin/env bash

ps -ef | egrep 'gridsdr|discosrv|resman'

pkill discosrv
pkill gridsdr
pkill resman
