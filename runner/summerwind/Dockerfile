# SPDX-License-Identifier: MIT

FROM summerwind/actions-runner:ubuntu-22.04

USER root

COPY my-entrypoint.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/my-entrypoint.sh

USER 1001

ENTRYPOINT ["/usr/local/bin/my-entrypoint.sh"]
