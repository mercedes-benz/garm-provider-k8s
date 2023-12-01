# SPDX-License-Identifier: MIT

allow_k8s_contexts('garm-operator')

# build garm-provider-k8s binary with 'make build copy'
local_resource(
    "build provider",
    cmd="make build copy",
    auto_init=True,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=["garm-k8s-provider"],
    deps=["."]
)

# build garm image garm-with-k8s and push to localhost:5000/garm-with-k8s in ./hack context
docker_build(
    'localhost:5000/garm-with-k8s',
    './hack'
)

# build gh action runner  image in ./runner context
cpu_arch = str(local('uname -m')).strip()
image_tag="localhost:5000/runner:linux-ubuntu-22.04-" + cpu_arch

local_resource(
    "build runner",
    cmd="RUNNER_IMAGE=" + image_tag + " make docker-build-runner",
    auto_init=True,
    trigger_mode=TRIGGER_MODE_AUTO,
    labels=["runner"],
    deps=["./runner"]
)

templated_yaml = kustomize('hack/local-development/kubernetes')
k8s_yaml(templated_yaml)
