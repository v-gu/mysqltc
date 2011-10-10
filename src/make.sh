ROOT="$PWD/$(dirname $0)/../"
GOPATH="$ROOT"
export GOPATH
MAKE=make
VERBOSE=false

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

if [ "$VERBOSE" == "true" ]; then
    goinstall -v mtc-cordump
    goinstall -v mtc-rpl-monitor
else
    goinstall mtc-cordump
    goinstall mtc-rpl-monitor
fi
