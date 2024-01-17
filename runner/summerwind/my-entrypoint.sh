#!/bin/bash
# SPDX-License-Identifier: MIT

set -x
set -e
set -o pipefail

if [ -z "$METADATA_URL" ];then
	echo "METADATA_URL is not set"
	exit 1
fi

if [ -z "$CALLBACK_URL" ];then
	echo "CALLBACK_URL is not set"
	exit 1
fi

export RUNNER_TOKEN=$(curl --retry 5 --retry-max-time 5 --fail -s -X GET -H 'Accept: application/json' -H "Authorization: Bearer ${BEARER_TOKEN}" "${METADATA_URL}/runner-registration-token/")

# this if statement is necessary for the startup.sh script from the actions-runner-controller to call the correct github url
# otherwise the runner creation on repository level would not work because the url is not set correctly
# see: https://github.com/actions/actions-runner-controller/blob/aa6dab5a9ac2a730ecd009a45ca49c7413fd8c42/runner/startup.sh#L35-L36
if [ -n "${RUNNER_ORG}" ] && [ -n "${RUNNER_REPO}" ] && [ -z "${RUNNER_ENTERPRISE}" ]; then
  export RUNNER_ENTERPRISE="dummy"
fi

function call() {
	PAYLOAD="$1"
  local cb_url=$CALLBACK_URL
  # strip status from the callback url
  [[ $cb_url =~ ^(.*)/status(/)?$ ]] || cb_url="${cb_url}/status" || true
	curl --fail -s -v -X POST -d "${PAYLOAD}" -H 'Accept: application/json' -H "Authorization: Bearer ${BEARER_TOKEN}" "${cb_url}" || { echo "failed to call home: exit code ($?)"; true; }
}
function sendStatus() {
	MSG="$1"
	call "{\"status\": \"installing\", \"message\": \"$MSG\"}"
}
function success() {
	MSG="$1"
	ID=$2
	call "{\"status\": \"idle\", \"message\": \"$MSG\", \"agent_id\": $ID}"
}
function fail() {
	MSG="$1"
	call "{\"status\": \"failed\", \"message\": \"$MSG\"}"
	exit 1
}

function systemInfo() {
  if [ -f "/etc/os-release" ];then
    . /etc/os-release
  fi

  local cb_url=$CALLBACK_URL
  OS_NAME=${NAME:-""}
  OS_VERSION=${VERSION_ID:-""}
  AGENT_ID=${1:-null}
  # strip status from the callback url
  [[ $cb_url =~ ^(.*)/status(/)?$ ]] && cb_url="${BASH_REMATCH[1]}" || true
  SYSINFO_URL="${cb_url}/system-info/"
  PAYLOAD="{\"os_name\": \"$OS_NAME\", \"os_version\": \"$OS_VERSION\", \"agent_id\": $AGENT_ID}"
  curl --retry 5 --retry-delay 5 --retry-connrefused --fail -s -X POST -d "${PAYLOAD}" -H 'Accept: application/json' -H "Authorization: Bearer ${BEARER_TOKEN}" "${SYSINFO_URL}" || true
}

function check_runner {
    echo "Checking runner health..."
    RETRIES=0
    MAX_RETRIES=15

    while true; do
        if [[ $RETRIES -eq $MAX_RETRIES ]]; then
          echo "failed to start runner"
          fail "failed to start runner"
        fi

        if [ -f /runner/.runner ]; then
          AGENT_ID=$(grep "agentId" /runner/.runner |  tr -d -c 0-9)
          echo "Calling $CALLBACK_URL with $AGENT_ID"
          systemInfo "$AGENT_ID"
          success "runner successfully installed" "$AGENT_ID"
          break
        fi

        RETRIES=$(expr $RETRIES + 1)
        echo "RETRIES: $RETRIES"
        sleep 10
    done
}

# Run the original entrypoint script in the background
/usr/bin/entrypoint.sh &

sleep 5

# Run check_runner function in the background
check_runner &

# Wait for all background jobs to finish
wait
