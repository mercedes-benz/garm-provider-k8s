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

# build image localhost:5000/garm-provider-k8s:latest in current context
docker_build(
    "localhost:5000/garm-with-k8s",
    ".",
    dockerfile="./hack/Dockerfile"
)

# build gh action runner  image in ./runner context
image_tag="localhost:5000/runner:linux-ubuntu-22.04"

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

# deploy the garm-operator
k8s_yaml('hack/local-development/kubernetes/garm-operator-all.yaml')

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
    ],
    labels=["operator"],
)

# Extract all resource names and kinds dynamically from the garm-operator-crs.yaml
k8s_yaml('hack/local-development/kubernetes/garm-operator-crs.yaml')

def get_operator_objects():
    result = []
    yaml_content = read_yaml_stream('hack/local-development/kubernetes/garm-operator-crs.yaml')
    for resource in yaml_content:
        if 'kind' in resource and 'metadata' in resource:
            kind = resource['kind']
            metadata = resource['metadata']
            if 'name' in metadata:
                name = metadata['name']
                namespace = metadata.get('namespace', 'garm-operator-system') if 'namespace' in metadata else 'garm-operator-system'
                result.append(name + ':' + kind + ':' + namespace)
    print('Total objects returned: %s' % len(result))
    return result

operator_objects = get_operator_objects()

k8s_resource(
    new_name='garm-operator-crs',
    objects=operator_objects,
    labels=["operator"],
    resource_deps=['garm-operator-controller-manager']
)
