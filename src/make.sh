#!/usr/bin/env sh

# select which applications to install
APPS="mtc-cordump mtc-rpl-monitor"


. "$(dirname $0)/env.sh"

# parse opts
while getopts "v" opt; do
    case $opt in
        v)
            VERBOSE=true
            ;;
        \?)
            ;;
    esac
done

# clean up src folder
eval "$(dirname $0)/clean.sh" >/dev/null 2>&1

for i in $APPS;
do
    if [ "$VERBOSE" = "true" ]; then
        goinstall -v "$i"
    else
        goinstall "$i"
    fi
done