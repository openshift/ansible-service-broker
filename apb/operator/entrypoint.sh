#!/bin/bash -e
# Patch /etc/passwd file with the current user info.
# The current user's entry must be correctly defined in this file in order for
# the `ssh` command to work within the created container.
if ! whoami &>/dev/null; then
  if [ -w /etc/passwd ]; then
    echo "${USER_NAME:-operator}:x:$(id -u):$(id -g):${USER_NAME:-operator} user:${HOME}:/sbin/nologin" >> /etc/passwd
  fi
fi
exec "${OPERATOR:-/usr/local/bin/ansible-operator}"
