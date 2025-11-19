# SPDX-License-Identifier: MIT

ARG garm_repo_ref=v0.1.5

FROM golang:1.24.2 AS build

WORKDIR /app

# Get default value from ARG assigned before FROM
ARG garm_repo_ref

RUN ARCH=$(uname -m) && \
    case ${ARCH} in \
      x86_64) ARCH=amd64 ;; \
      aarch64) ARCH=arm64 ;; \
      *) echo "Unsupported architecture: ${ARCH}" ;; \
    esac && \
    wget -q https://github.com/cloudbase/garm/releases/download/${garm_repo_ref}/garm-cli-linux-${ARCH}.tgz && \
    tar -xzf garm-cli-linux-${ARCH}.tgz && \
    mv garm-cli /go/bin/garm-cli && \
    chmod +x /go/bin/garm-cli && \
    rm garm-cli-linux-${ARCH}.tgz


COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o ./garm-provider-k8s ./cmd/garm-provider-k8s

RUN chmod +x ./garm-provider-k8s && \
    chmod +x /go/bin/garm-cli


FROM ghcr.io/cloudbase/garm:${garm_repo_ref}

COPY --from=build /go/bin/garm-cli ./bin/garm-cli

COPY --from=build /app/garm-provider-k8s /opt/garm/providers.d/garm-provider-k8s
