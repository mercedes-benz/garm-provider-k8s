# SPDX-License-Identifier: MIT

allow_k8s_contexts('garm')

# we need cert-manager for the garm-operator
load('ext://cert_manager', 'deploy_cert_manager')

# we use the `cert_manager` extension to deploy cert-manager into the kind-cluster
# as the plugin has already well written readiness checks we can use it to wait for
deploy_cert_manager(
    kind_cluster_name='garm', # just for security reasons ;-)
    version='v1.12.0' # the version of cert-manager to deploy
)

# prepare garm-operator
local_resource(
    "prepare garm-operator",
    cmd="make prepare-operator",
    auto_init=True,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=["garm-operator"],
    deps=["."]
)

k8s_yaml('hack/local-development/kubernetes/garm-operator-all.yaml')


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
    cmd="RUNNER_IMAGE=" + image_tag + " make docker-build-summerwind-runner",
    auto_init=True,
    trigger_mode=TRIGGER_MODE_AUTO,
    labels=["runner"],
    deps=["./runner/summerwind"]
)

# take care of the kubernetes manifests where garm with the provider binary is deployed
templated_yaml = kustomize('hack/local-development/kubernetes')
k8s_yaml(templated_yaml)
