#!/bin/bash

#!/bin/bash

set -ex
set -o pipefail

if [ -z "$METADATA_URL" ];then
	echo "no token is available and METADATA_URL is not set"
	exit 1
fi

if [ -z "$CALLBACK_URL" ];then
	echo "CALLBACK_URL is not set"
	exit 1
fi

function call() {
	PAYLOAD="$1"
	[[ $CALLBACK_URL =~ ^(.*)/status(/)?$ ]] || CALLBACK_URL="${CALLBACK_URL}/status"
	curl --retry 5 --retry-delay 5 --retry-connrefused --fail -s -X POST -d "${PAYLOAD}" -H 'Accept: application/json' -H "Authorization: Bearer ${BEARER_TOKEN}" "${CALLBACK_URL}" || echo "failed to call home: exit code ($?)"
}

function sendStatus() {
	MSG="$1"
	call "{\"status\": \"installing\", \"message\": \"$MSG\"}"
}

function fail() {
	MSG="$1"
	call "{\"status\": \"failed\", \"message\": \"$MSG\"}"
	exit 1
}

function success() {
	MSG="$1"
	ID=$2
    if [ $JIT_CONFIG_ENABLED != "true" ] && [ -z "$ID" ]; then
        fail "agent ID is required when JIT_CONFIG_ENABLED is not true"
    fi

    if [ -z "$ID" ]; then
        ID="null"
    fi
	call "{\"status\": \"idle\", \"message\": \"$MSG\", \"agent_id\": $ID}"
}

function getRunnerFile() {
	curl --retry 5 --retry-delay 5 \
		--retry-connrefused --fail -s \
		-X GET -H 'Accept: application/json' \
		-H "Authorization: Bearer ${BEARER_TOKEN}" \
		"${METADATA_URL}/$1" -o "$2"
}

pushd /home/runner/

sendStatus "configuring runner"
if [ $JIT_CONFIG_ENABLED == "true" ]; then
    sendStatus "downloading JIT credentials"
    getRunnerFile "credentials/runner" "/home/runner/.runner" || fail "failed to get runner file"
    getRunnerFile "credentials/credentials" "/home/runner/.credentials" || fail "failed to get credentials file"
    getRunnerFile "credentials/credentials_rsaparams" "/home/runner/.credentials_rsaparams" || fail "failed to get credentials_rsaparams file"
else
    if [ -n "${RUNNER_ORG}" ] && [ -n "${RUNNER_REPO}" ]; then
        ATTACH="${RUNNER_ORG}/${RUNNER_REPO}"
    elif [ -n "${RUNNER_ORG}" ]; then
        ATTACH="${RUNNER_ORG}"
    elif [ -n "${RUNNER_REPO}" ]; then
        ATTACH="${RUNNER_REPO}"
    elif [ -n "${RUNNER_ENTERPRISE}" ]; then
        ATTACH="enterprises/${RUNNER_ENTERPRISE}"
    else
        fail 'At least one of RUNNER_ORG, RUNNER_REPO, or RUNNER_ENTERPRISE must be set'
        exit 1
    fi

    REPO_URL="${GITHUB_URL}/${ATTACH}"

    GITHUB_TOKEN=$(curl --retry 5 --retry-delay 5 --retry-connrefused --fail -s -X GET -H 'Accept: application/json' -H "Authorization: Bearer ${BEARER_TOKEN}" "${METADATA_URL}/runner-registration-token/")
    set +e
    attempt=1
    while true; do
        ERROUT=$(mktemp)
        if [ ! -z $RUNNER_GROUP ] && [ $RUNNER_GROUP != "default" ];then
            ./config.sh --unattended \
                --url "$REPO_URL" \
                --token "$GITHUB_TOKEN" \
                --runnergroup $RUNNER_GROUP \
                --name "$RUNNER_NAME" \
                --labels "$RUNNER_LABELS" \
                --ephemeral 2>$ERROUT
        else
            ./config.sh --unattended \
                --url "$REPO_URL" \
                --token "$GITHUB_TOKEN" \
                --name "$RUNNER_NAME" \
                --labels "$RUNNER_LABELS" \
                --ephemeral 2>$ERROUT
        fi
        if [ $? -eq 0 ]; then
            rm $ERROUT || true
            sendStatus "runner successfully configured after $attempt attempt(s)"
            break
        fi
        LAST_ERR=$(cat $ERROUT)
        echo "$LAST_ERR"

        # if the runner is already configured, remove it and try again. In the past configuring a runner
        # managed to register it but timed out later, resulting in an error.
        ./config.sh remove --token "$GITHUB_TOKEN" || true

        if [ $attempt -gt 5 ];then
            rm $ERROUT || true
            fail "failed to configure runner: $LAST_ERR"
        fi

        sendStatus "failed to configure runner (attempt $attempt): $LAST_ERR (retrying in 5 seconds)"
        attempt=$((attempt+1))
        rm $ERROUT || true
        sleep 5
    done
    set -e
fi

AGENT_ID=""

if [ $JIT_CONFIG_ENABLED != "true" ]; then
    set +e
    AGENT_ID=$(grep "agentId" /home/runner/.runner |  tr -d -c 0-9)
    if [ $? -ne 0 ];then
        fail "failed to get agent ID"
    fi
    set -e
fi

success "runner successfully installed" $AGENT_ID
./run.sh "$@"