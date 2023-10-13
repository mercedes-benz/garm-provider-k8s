#!/bin/sh
# SPDX-License-Identifier: MIT

alias gcli=/home/mb/bin/garm/garm-cli
alias garm=/home/mb/bin/garm/garm

ls -la /home/mb/bin/garm/
echo $DEBUG

if [ "$DEBUG" = "true" ]; then
  # build garm as follows for dlv to work
  # go build -gcflags="all=-N -l" -mod vendor -o $BIN_DIR/garm -tags osusergo,netgo,sqlite_omit_load_extension -ldflags "-linkmode external -extldflags '-static' -X main.Version=$(git describe --always --dirty)" .
  /home/mb/bin/garm/dlv --listen=:40000 --headless=true --api-version=2 --accept-multiclient exec /home/mb/bin/garm/garm -- -config /home/mb/bin/garm/config.toml &
else
  /home/mb/bin/garm/garm -config /home/mb/bin/garm/config.toml &
fi

sleep 5

gcli init -e=garm@mercedes-benz.com -n admin -p'LmrBG1KcBOsDfNKq4cQTGpc0hJ0kejkk' -a http://127.0.0.1:9997 -u admin --debug

ORG=$(gcli organization create \
  --credentials=GitHub-Actions \
  --name=GitHub-Actions \
  --webhook-secret="mysecret" | grep -E "^| ID" | awk -F'|' 'NR==4 {gsub(/^[[:space:]]+|[[:space:]]+$/, "", $3); print $3}')

# should be set from ENV variable of deployment according to your local cpu arch
if [ -z "$ARCH" ]; then
  export ARCH="amd64"
fi

gcli pool add \
  --image="roadrunner-default-container:latest"  \
  --provider-name=kubernetes_external \
  --tags=kubernetes,ubuntu \
  --os-type=linux \
  --os-arch=arm64 \
  --flavor=small \
  --org="$ORG" \
  --runner-group=kubernetes \
  --enabled=true

while true; do
  sleep 10
done
