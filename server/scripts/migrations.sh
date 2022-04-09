#!/usr/bin/env bash

set -euo pipefail
# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=$(readlink -f "$0")
# Absolute path this script is in, thus /home/user/bin
SCRIPTPATH=$(dirname "$SCRIPT")

if [ -z "${DB_USER-}" ]; then
    echo "Missing database DB_USER env"
    exit 1
fi
if [ -z "${DB_PASSWORD-}" ]; then
    echo "Missing database DB_PASSWORD env"
    exit 1
fi
if [ -z "${DB_HOST-}" ]; then
    echo "Missing database DB_HOST env"
    exit 1
fi
if [ -z "${DB_PORT-}" ]; then
    echo "Missing database DB_PORT env"
    exit 1
fi
if [ -z "${DB_DATABASE-}" ]; then
    echo "Missing database DB_DATABASE env"
    exit 1
fi
if [ -z "${DB_AUTH_SOURCE-}" ]; then
    echo "Missing database DB_AUTH_SOURCE env"
    exit 1
fi

function run_migrations() {
    echo "Executing from dir $SCRIPTPATH/migrations"
    migrate -database "mongodb://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE}?authSource=${DB_AUTH_SOURCE}" -path $SCRIPTPATH/../migrations "$@"
}

run_migrations "$@"
exit 0