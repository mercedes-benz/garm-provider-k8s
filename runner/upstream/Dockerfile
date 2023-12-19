# SPDX-License-Identifier: MIT

FROM ghcr.io/actions/actions-runner:2.311.0

USER root

COPY entrypoint.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint.sh

USER 1001

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
