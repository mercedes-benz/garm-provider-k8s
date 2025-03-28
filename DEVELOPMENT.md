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

1. Get yourself a GitHub PAT for development purposes with access to a Repository where the runners will be 
   registered. Your PAT needs the following permissions:
   If you'll use a PAT (classic), you'll have to grant access for the following scopes. See official [cloudbase/garm](https://github.com/cloudbase/garm/blob/main/doc/github_credentials.md) docs for more information.

   * ```public_repo``` - for access to a repository
   * ```repo``` - for access to a private repository
   * ```admin:org``` - if you plan on using this with an organization to which you have access
   * ```manage_runners:enterprise``` - if you plan to use garm at the enterprise level
   * ```admin:repo_hook``` - if you want to allow GARM to install webhooks on repositories (optional)
   * ```admin:org_hook``` - if you want to allow GARM to install webhooks on organizations (optional)

Fine grained PATs are also supported as long as you grant the required privileges:

   * **Repository permissions**:
     * `Administration: Read & write` - needed to generate JIT config/registration token, remove runners, etc.
     * `Metadata: Read-only` - automatically enabled by above
     * `Webhooks: Read & write` - needed to install webhooks on repositories
  * **Organization permissions**:
     * `Self-hosted runners: Read & write` - needed to manage runners in an organization
     * `Webhooks: Read & write` - needed to install webhooks on organizations

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
   ``` 
   NAME                                                                  ID                                     VERSION   AGE
   garmserverconfig.garm-operator.mercedes-benz.com/garm-server-config   65cc57ab-3a6c-48b9-9565-af7332e65f32   v0.1.5    10m
   
   NAME                                                          ID    READY   AUTHTYPE   GITHUBENDPOINT   AGE
   githubcredential.garm-operator.mercedes-benz.com/github-pat   3     True    pat        github           28m
   
   NAME                                                    URL                      READY   AGE
   githubendpoint.garm-operator.mercedes-benz.com/github   https://api.github.com   True    28m
   
   NAME                                                   TAG                                              AGE
   image.garm-operator.mercedes-benz.com/runner-default   localhost:5000/runner:linux-ubuntu-22.04-arm64   10m
   
   NAME                                                        ID                                     MINIDLERUNNERS   MAXRUNNERS   READY   AGE
   pool.garm-operator.mercedes-benz.com/kubernetes-pool-repo   8d8699c3-0acb-4fd5-9f5a-8b60b7504eca   1                2            True    10m
   
   NAME                                                        ID                                     READY   AGE
   repository.garm-operator.mercedes-benz.com/test-workflows   3cd28551-18f8-430d-a0c8-49cf24dd7355   True    10m
   
   NAME                                                           ID                                     POOL                   GARM RUNNER STATUS   PROVIDER RUNNER STATUS   AGE
   runner.garm-operator.mercedes-benz.com/garm-k8s-vi5mdto8lnav   9e475cad-0e22-4db2-b53b-3d8cb184439e   kubernetes-pool-repo   running              idle                     5m46s

   ```
3. Also in the `runner` namespace, you should see created runner pods by the `garm-provider-k8s`.
   ```bash
   $ kubectl get pods -n runner
   ```
   ```
   NAME                    READY   STATUS    RESTARTS   AGE
   garm-k8s-vi5mdto8lnav   1/1     Running   0          6m14s
   ```

4. Everytime you now change some code, tilt will automatically build a new container image with the `garm-server` and
   `garm-provider-k8s` binary and update the image tag in the k8s Deployment manifest. Time to start developing! 🎉
