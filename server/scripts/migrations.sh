#!/usr/bin/env bash

set -euo pipefail
# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=$(readlink -f "$0")
# Absolute path this script is in, thus /home/user/bin
SCRIPTPATH=$(dirname "$SCRIPT")

if [ -z "${MONGODB_URI-}" ]; then
    echo "Missing database DB_AUTH_SOURCE env"
    exit 1
fi

function run_migrations() {
    echo "Executing from dir $SCRIPTPATH/migrations"
    migrate -database "${MONGODB_URI}" -path $SCRIPTPATH/../migrations "$@"
}

run_migrations "$@"
exit 0