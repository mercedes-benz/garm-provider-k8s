# SPDX-License-Identifier: MIT

FROM summerwind/actions-runner:ubuntu-22.04

USER root

RUN apt-get update && apt-get install -y curl && apt-get clean

COPY entrypoint.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint.sh

USER 1001

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
