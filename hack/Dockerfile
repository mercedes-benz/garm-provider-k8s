# SPDX-License-Identifier: MIT

FROM golang:1.20.4 AS build

ARG garm_repo=https://github.com/cloudbase/garm
ARG garm_repo_ref=fdfcc619adbb6f79814471294ff145bb609a3714

# build garm binary
RUN git clone $garm_repo garm_repo && \
  cd garm_repo && \
  git checkout $garm_repo_ref && \
  CGO_ENABLED=1 go install -ldflags "-linkmode external -extldflags '-static' -X main.Version=$garm_repo_ref" ./cmd/garm && \
  CGO_ENABLED=1 go install -ldflags "-linkmode external -extldflags '-static' -X main.Version=$garm_repo_ref" ./cmd/garm-cli


FROM alpine:3.18

RUN addgroup -S mb && \
  adduser -u 1001 -S mb -G mb

WORKDIR /home/mb/bin/garm

COPY --from=build --chown=1001:mb /go/bin/garm ./garm
COPY --from=build --chown=1001:mb /go/bin/garm-cli ./garm-cli

COPY --chown=1001:mb ./config/config.toml .
COPY --chown=1001:mb ./config/garm-provider-k8s-config.yaml .
COPY --chown=1001:mb ./scripts/entrypoint.sh .
COPY --chown=1001:mb ./bin/garm-provider-k8s .

RUN chmod +x ./entrypoint.sh && \
    chmod +x ./garm && \
    chmod +x ./garm-cli && \
    chmod +x ./garm-provider-k8s

RUN chown 1001:mb /home/mb/bin/garm && \
    chmod u+w /home/mb/bin/garm

USER 1001

EXPOSE 9997

ENTRYPOINT ["/home/mb/bin/garm/entrypoint.sh"]
