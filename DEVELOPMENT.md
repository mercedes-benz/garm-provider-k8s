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
      GARM_GITHUB_NAME=comario-dev GARM_GITHUB_OAUTH_TOKEN=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 make template
      ```

1. Start the development environment by running `make tilt-up` in the root directory of this repository.

   This will start a local Kubernetes cluster using `kind` (`kind get clusters` will show you a `garm` cluster) and deploy a garm-server with an already registered `garm-provider-k8s`.
   The `make tilt-up` command will give you the URL to the local tilt environment.

1. Time to start developing. 🎉
