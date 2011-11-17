#!/usr/bin/env bash

. "$(dirname $0)/env.bash"

find ${ROOT} -name Makefile -a -type f -execdir make clean \;
