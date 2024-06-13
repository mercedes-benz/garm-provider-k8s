# SPDX-License-Identifier: MIT

FROM ghcr.io/actions/actions-runner:2.317.0

USER root

RUN apt-get update && apt-get install -y curl && apt-get clean

COPY entrypoint.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint.sh

USER 1001

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
