#!/bin/sh
# SPDX-License-Identifier: MIT

alias gcli=/opt/garm/bin/garm-cli
alias garm=/opt/garm/bin/garm

if [ "$DEBUG" = "true" ]; then
  # build garm as follows for dlv to work
  # go build -gcflags="all=-N -l" -mod vendor -o $BIN_DIR/garm -tags osusergo,netgo,sqlite_omit_load_extension -ldflags "-linkmode external -extldflags '-static' -X main.Version=$(git describe --always --dirty)" .
  /home/mb/bin/garm/dlv --listen=:40000 --headless=true --api-version=2 --accept-multiclient exec /home/mb/bin/garm/garm -- -config /home/mb/bin/garm/config.toml &
else
  /opt/garm/bin/garm -config /opt/garm/config/config.toml &
fi

sleep 5

# TODO: remove this once https://github.com/mercedes-benz/garm-operator/pull/30 got merged and
# extend the local dev-setup script to do the initialization via the operator
gcli init -e=garm@mercedes-benz.com -n admin -p'LmrBG1KcBOsDfNKq4cQTGpc0hJ0kejkk' -a http://127.0.0.1:9997 -u admin --debug

while true; do
  sleep 10
done
