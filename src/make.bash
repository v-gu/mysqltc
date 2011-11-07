#!/usr/bin/env bash

# default applications to install
APPS="\
mtc-cordump \
mtc-rplerr-monitor\
"
S_APPS=""

. "$(dirname $0)/env.sh"

# parse opts
while getopts "a:v" opt; do
    case $opt in
        a)
            S_APPS=${S_APPS}${OPTARG}
            ;;
        v)
            VERBOSE=true
            ;;
        \?)
            ;;
    esac
    shift
done

# clean up src folder
eval "$(dirname $0)/clean.sh" >/dev/null 2>&1

# build
if [ "${S_APPS}" != "" ];then
    for i in $S_APPS;
    do
        if [ "$VERBOSE" = "true" ]; then
            goinstall -v "$i"
        else
            goinstall "$i"
        fi
    done
else
    for i in $APPS;
    do
        if [ "$VERBOSE" = "true" ]; then
            goinstall -v "$i"
        else
            goinstall "$i"
        fi
    done
fi