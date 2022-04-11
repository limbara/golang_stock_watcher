#!/usr/bin/env bash

# https://devcenter.heroku.com/articles/exec#using-with-docker
# https://gist.github.com/wwerner/05a8e627e8f3ba18300db745511d3bcb
set +o posix

[ -z "$SSH_CLIENT" ] && source <(curl --fail --retry 3 -sSL "$HEROKU_EXEC_URL")