<!-- SPDX-License-Identifier: MIT -->

# garm-provider-k8s

<!-- toc -->
- [✨ What is the <code>garm-provider-k8s</code>?](#-what-is-the-garm-provider-k8s)
- [🔀 Versioning](#-versioning)
    - [Garm Version](#garm-version)
    - [Kubernetes Version](#kubernetes-version)
- [🚀 Installation](#-installation)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
- [💻 Development](#-development)
- [📋 ADRs](#-adrs)
- [Contributing](#contributing)
- [Code of Conduct](#code-of-conduct)
- [License](#license)
- [Provider Information](#provider-information)
<!-- /toc -->

## ✨ What is the `garm-provider-k8s`?

garm-provider-k8s` is a [Kubernetes®](https://kubernetes.io) a plugin for [garm](https://github.com/cloudbase/garm) which enables garm to spin up Pod based GitHub Action runners on a k8s cluster.

## 🔀 Versioning

### Kubernetes Version

garm-provider-k8s uses [`client-go`](https://github.com/kubernetes/client-go) to talk with
Kubernetes clusters. The supported Kubernetes cluster version is determined by `client-go`.
The compatibility matrix for client-go and Kubernetes cluster can be found
[here](https://github.com/kubernetes/client-go#compatibility-matrix).

## 🚀 Installation

### Prerequisites

1. A Kubernetes cluster to spin up runners with the `garm-provider-k8s` plugin.
2. A working `garm` installation.

### Installation

#### `garm-provider-k8s`

We are releasing the `garm-provider-k8s` as a simple go binary. You can find the latest release [here](https://github.com/mercedes-benz/garm-provider-k8s/releases).
Adjust your existing `garm` config, so garm can register to provider plugin at runtime. Make sure garm has access to the binary and provider specific config path like so:

```toml
[[provider]]
name = "kubernetes_external"
description = "kubernetes provider"
provider_type = "external"
[provider.external]
config_file = "/path/to/garm-provider-k8s-config.yaml"
provider_executable = "/path/to/provider/binary/garm-provider-k8s"
environment_variables = ["KUBERNETES_"]
```

## 💻 Development

For local development, please read the [development guide](DEVELOPMENT.md).

## Contributing

We welcome any contributions.
If you want to contribute to this project, please read the [contributing guide](CONTRIBUTING.md).

## Code of Conduct

Please read our [Code of Conduct](https://github.com/mercedes-benz/foss/blob/master/CODE_OF_CONDUCT.md) as it is our base for interaction.

## License

This project is licensed under the [MIT LICENSE](LICENSE).

## Provider Information

Please visit <https://www.mercedes-benz-techinnovation.com/en/imprint/> for information on the provider.

Notice: Before you use the program in productive use, please take all necessary precautions,
e.g. testing and verifying the program with regard to your specific use.
The program was 
