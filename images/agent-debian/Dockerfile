FROM node:22-bookworm-slim@sha256:de170b39439eb0e5a3444ad7d60b015614a381e10fecae2e30ec9e735122f38d

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates curl git \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --uid 10001 --create-home --home-dir /home/agent agent

USER agent
ENV HOME=/home/agent
ENV PATH=/home/agent/.local/bin:$PATH

RUN curl -fsS https://cursor.com/install | bash \
 && curl -fsSL https://antigravity.google/cli/install.sh | bash

RUN agent --version && agy --version
