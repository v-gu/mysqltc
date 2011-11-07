#!/usr/bin/env bash

. "$(dirname $0)/env.sh"

find ${ROOT} -name Makefile -a -type f -execdir make clean \;
