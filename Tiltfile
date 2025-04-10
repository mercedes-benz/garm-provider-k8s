# SPDX-License-Identifier: MIT

allow_k8s_contexts('kind-garm')

# we need cert-manager for the garm-operator
load('ext://cert_manager', 'deploy_cert_manager')

# we use the `cert_manager` extension to deploy cert-manager into the kind-cluster
# as the plugin has already well written readiness checks we can use it to wait for
deploy_cert_manager(
    kind_cluster_name='kind-garm', # just for security reasons ;-)
    version='v1.15.3' # the version of cert-manager to deploy
)

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

k8s_resource(
    "garm-server",
    objects=[
        'garm-server:Namespace:default',
        'runner:Namespace:default',
        'garm-server:ServiceAccount:garm-server',
        'garm-provider-k8s:ClusterRole:default',
        'garm-provider-k8s:RoleBinding:runner',
        'garm-configuration:ConfigMap:garm-server',
        'garm-kubernetes-provider-config:ConfigMap:garm-server',
        'garm-data:PersistentVolumeClaim:garm-server',
        'garm-home:PersistentVolumeClaim:garm-server'

    ],
    labels=["garm-server"],
)

# deploy the garm-operator and CRs
k8s_yaml('hack/local-development/kubernetes/garm-operator-all.yaml')

k8s_yaml('hack/local-development/kubernetes/garm-operator-crs.yaml')

k8s_resource(
    "garm-operator-controller-manager",
    objects=[
        'garm-operator-system:Namespace:default',
        'enterprises.garm-operator.mercedes-benz.com:CustomResourceDefinition:default',
        'garmserverconfigs.garm-operator.mercedes-benz.com:CustomResourceDefinition:default',
        'githubcredentials.garm-operator.mercedes-benz.com:CustomResourceDefinition:default',
        'githubendpoints.garm-operator.mercedes-benz.com:CustomResourceDefinition:default',
        'images.garm-operator.mercedes-benz.com:CustomResourceDefinition:default',
        'organizations.garm-operator.mercedes-benz.com:CustomResourceDefinition:default',
        'pools.garm-operator.mercedes-benz.com:CustomResourceDefinition:default',
        'repositories.garm-operator.mercedes-benz.com:CustomResourceDefinition:default',
        'runners.garm-operator.mercedes-benz.com:CustomResourceDefinition:default',
        'garm-operator-controller-manager:ServiceAccount:garm-operator-system',
        'garm-operator-leader-election-role:Role:garm-operator-system',
        'garm-operator-manager-role:Role:garm-operator-system',
        'garm-operator-manager-role:ClusterRole:default',
        'garm-operator-leader-election-rolebinding:RoleBinding:garm-operator-system',
        'garm-operator-manager-rolebinding:RoleBinding:garm-operator-system',
        'garm-operator-kube-state-metrics-config:ConfigMap:garm-operator-system',
        'garm-operator-serving-cert:Certificate:garm-operator-system',
        'garm-operator-selfsigned-issuer:Issuer:garm-operator-system',
        'garm-operator-validating-webhook-configuration:ValidatingWebhookConfiguration:default',
        'garm-server-config:GarmServerConfig:garm-operator-system',
        'github:GitHubEndpoint:garm-operator-system',
        'github-pat:GitHubCredential:garm-operator-system',
        'github-pat:Secret:garm-operator-system',
        'test-workflows:Repository:garm-operator-system',
        'repo-webhook-secret:Secret:garm-operator-system',
        'runner-default:Image:garm-operator-system',
        'kubernetes-pool-repo:Pool:garm-operator-system'
    ],
    labels=["operator"],
)
