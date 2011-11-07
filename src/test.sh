#!/usr/bin/env sh

# default applications to test
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

# test
SRCDIR="${PWD}/$(dirname $0)"
if [ "${S_APPS}" != "" ];then
    for i in $S_APPS;
    do
        if [ "$VERBOSE" = "true" ]; then
            cd "${SRCDIR}/${i}"
            gotest -x -v "$@"
            cd ${SRCDIR}
        else
            cd "${SRCDIR}/${i}"
            gotest "$@"
            cd ${SRCDIR}
        fi
    done
else
    for i in $APPS;
    do
        if [ "$VERBOSE" = "true" ]; then
            cd "${SRCDIR}/${i}"
            gotest -x -v "$@"
            cd ${SRCDIR}
        else
            cd "${SRCDIR}/${i}"
            gotest "$@"
            cd ${SRCDIR}
        fi
    done
fi
