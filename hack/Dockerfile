# SPDX-License-Identifier: MIT

FROM golang:1.22.4 AS build

ARG garm_repo=https://github.com/cloudbase/garm
ARG garm_repo_ref=v0.1.4

# build garm binary
# primarly used to get the binary build for
# the local CPU architecture
RUN git clone $garm_repo garm_repo && \
  cd garm_repo && \
  git checkout $garm_repo_ref && \
  CGO_ENABLED=1 go install -ldflags "-linkmode external -extldflags '-static' -X main.Version=$garm_repo_ref" ./cmd/garm && \
  CGO_ENABLED=1 go install -ldflags "-linkmode external -extldflags '-static' -X main.Version=$garm_repo_ref" ./cmd/garm-cli


FROM alpine:3.18

WORKDIR /opt/garm/

ENV PATH="${PATH}:/opt/garm/bin"

COPY --from=build /go/bin/garm ./bin/garm
COPY --from=build /go/bin/garm-cli ./bin/garm-cli

COPY ./scripts/entrypoint.sh ./bin/entrypoint.sh
COPY ./bin/garm-provider-k8s ./bin/garm-provider-k8s

RUN chmod +x ./bin/entrypoint.sh && \
  chmod +x ./bin/garm && \
  chmod +x ./bin/garm-cli && \
  chmod +x ./bin/garm-provider-k8s

USER 65543:65543

EXPOSE 9997

ENTRYPOINT ["/opt/garm/bin/entrypoint.sh"]
