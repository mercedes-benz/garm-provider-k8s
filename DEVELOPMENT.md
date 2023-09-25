<!-- SPDX-License-Identifier: MIT -->

# Development

<!-- toc -->
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
<!-- /toc -->

## Prerequisites

To start developing, you need the following tools installed:

- [Go](https://golang.org/doc/install)
- [Docker](https://docs.docker.com/get-docker/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [Tilt](https://docs.tilt.dev/install.html)

## Getting Started

1. Get yourself a GitHub PAT for development purposes with access to an Organization where the runners will be registered.

1. To configure `garm` you have to generate a new `config.toml` with your previous created PAT.
   You can use the `make template` target in the root directory of this repository to generate a new `configMap` which contains a valid `config.toml` for `garm`
   
      ```bash
      GARM_GITHUB_OAUTH_TOKEN=ghp_QVK50pUo25QsJBy5DfA95EyUCzAbG20Q1NP2 make template
      ```

1. Start the development environment by running `make tilt-up` in the root directory of this repository.

   This will start a local Kubernetes cluster using `kind` (`kind get clusters` will show you a `garm` cluster) and deploy a garm-server with an already registered `garm-provider-k8s`.
   The `make tilt-up` command will give you the URL to the local tilt environment.

1. As we also written the [garm-operator](https://github.com/mercedes-benz/garm-operator) to manage
   `garm` resources (like `Pools`, `Organizations` and so on) we are using the `garm-operator` to bootstrap the local development setup. 
   
   You will notice that in the `garm-operator-system` namespace the `garm-operator` is running.

1. You are now able to create the `garm` resources by either creating them via `garm-cli` in the `garm-server` container (in `garm-server` namespace) or via the `garm-operator` by applying the following manifest files to your local cluster. 
   
      **organization:**
      ```yaml
      apiVersion: garm-operator.mercedes-benz.com/v1alpha1
      kind: Organization
      metadata:
        labels:
          app.kubernetes.io/name: organization
          app.kubernetes.io/instance: organization-sample
          app.kubernetes.io/part-of: garm-operator
        name: your-org # this organization must exist and the PAT must have access to it
        namespace: garm-operator-system
      spec:
        webhookSecretRef:
          key: "webhookSecret"
          name: "org-webhook-secret"
        credentialsName: "github-pat"
      ---
      apiVersion: v1
      kind: Secret
      metadata:
        name: org-webhook-secret
        namespace: garm-operator-system
      data:
        webhookSecret: bXlzZWNyZXQ=
      ```

      **image:**
      ```yaml
      apiVersion: garm-operator.mercedes-benz.com/v1alpha1
      kind: Image
      metadata:
        labels:
          app.kubernetes.io/name: image
          app.kubernetes.io/instance: image-sample
          app.kubernetes.io/part-of: garm-operator
        name: runner-default
        namespace: garm-operator-system
      spec:
        tag: runner:linux-ubuntu-22.04-x86_64
      ```

      **pool:**
      ```yaml
      apiVersion: garm-operator.mercedes-benz.com/v1alpha1
      kind: Pool
      metadata:
        labels:
          app.kubernetes.io/instance: pool-sample
          app.kubernetes.io/name: pool
          app.kubernetes.io/part-of: garm-operator
        name: k8s-pool
        namespace: garm-operator-system
      spec:
        githubScopeRef:
          apiGroup: garm-operator.mercedes-benz.com
          kind: Organization
          name: your-org
        enabled: true
        extraSpecs: "{}"
        flavor: medium
        githubRunnerGroup: ""
        imageName: runner-default
        maxRunners: 4
        minIdleRunners: 2
        osArch: amd64
        osType: linux
        providerName: kubernetes_external # this is the name defined in your garm server
        runnerBootstrapTimeout: 20
        runnerPrefix: ""
        tags:
          - linux
          - kubernetes
      ```

      Now you should be able to see two pods in the `runner` namespace:
      ```bash
      $ kubectl get pods -n runner
      NAME                READY   STATUS    RESTARTS   AGE
      garm-hoyjldsegfal   1/1     Running   0          7s
      garm-hpohj7g2b4df   1/1     Running   0          7s
      ```

1. Time to start developing. 🎉
