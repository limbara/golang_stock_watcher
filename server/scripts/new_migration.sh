#!/usr/bin/env bash

set -euo pipefail
# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=$(readlink -f "$0")
# Absolute path this script is in, thus /home/user/bin
SCRIPTPATH=$(dirname "$SCRIPT")

function create_migration() {
    local name="$1"
    echo "Executing to dir $SCRIPTPATH/migrations"
    migrate create -ext json -dir $SCRIPTPATH/../migrations -seq $name
}

if [ -z "${1-}" ]; then
    echo "Missing migration name argument"
    exit 1
fi

create_migration $1
exit 0