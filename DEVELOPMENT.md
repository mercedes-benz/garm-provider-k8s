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

1. To spin up GitHub Action Runners with `garm`, the `garm-operator` needs some CRs which can be created by the 
   following command.
   You can use the `make template` target in the root directory of this repository to generate a new 
   `garm-operator-crs.yaml` which contains all CRs required for `garm-operator`.
   
      ```bash
      GARM_GITHUB_ORGANIZATION=my-github-org \
      GARM_GITHUB_REPOSITORY=my-github-repo \
      GARM_GITHUB_TOKEN=gha_testtoken \
      GARM_GITHUB_WEBHOOK_SECRET=supersecret \
      make template
      ```

1. Start the development environment by running
   ```bash
   $ make tilt-up
   ```
   in the root directory of this repository.

   This will start a local Kubernetes cluster using `kind` (`kind get clusters` will show you a `garm` cluster) and deploy a garm-server with an already registered `garm-provider-k8s`.
   The `make tilt-up` command will give you the URL to the local tilt environment.

2. As we also written the [garm-operator](https://github.com/mercedes-benz/garm-operator) to manage
   `garm` resources (like `Pools`, `Organizations` and so on) we are using the `garm-operator` to bootstrap the local development setup. 
   
   You will notice that in the `garm-operator-system` namespace the `garm-operator` is running, and your previously 
   created CRs are applied.
   ```bash
   $ kubectl get garm -n garm-operator-system
   ```
3. Also in the `runner` namespace, you should see created runner pods by the `garm-provider-k8s`.
   ```bash
   $ kubectl get pods -n runner
   ```

4. Time to start developing. 🎉
