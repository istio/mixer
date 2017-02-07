#!/bin/bash
SCRIPTPATH=$( cd "$(dirname "$0")" ; pwd -P )
ROOTDIR=$SCRIPTPATH/..
cd $ROOTDIR

ret=0
TMPFILE=$(mktemp)

function cleanup() {
	rm -f $TMPFILE
}

trap cleanup EXIT

grep -n commit WORKSPACE  | grep -v "#" > $TMPFILE
ret=$?

# found a commit line with no comment
if [[ $ret -eq 0 ]];then
	cat $TMPFILE
	echo "Missing comment"
	exit 1
fi

exit 0
