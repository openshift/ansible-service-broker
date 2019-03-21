#!/bin/bash

if ! whoami &>/dev/null; then
  if [ -w /etc/passwd ]; then
    echo "${USER_NAME:-molecule}:x:$(id -u):$(id -g):${USER_NAME:-molecule} user:${HOME}:/sbin/nologin" >> /etc/passwd
  fi
fi

exec $@
