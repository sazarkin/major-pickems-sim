FROM ghcr.io/astral-sh/uv:bookworm-slim
ARG USERNAME=dev
ENV UV_LINK_MODE=copy

# install dependencies
ARG DEBIAN_FRONTEND=noninteractive
RUN set -eux; \
    apt-get update; \
    apt-get install -y wget git zsh; \
    rm -rf /var/lib/apt/lists/*

# create user
RUN set -eux; \
    useradd --create-home --user-group --no-log-init "$USERNAME"; \
    mkdir -p "/home/$USERNAME/src"; \
    chown -R "$USERNAME:$USERNAME" "/home/$USERNAME"
USER "$USERNAME"
WORKDIR "/home/$USERNAME/src"

# make venv
RUN set -eux; uv venv .venv

# entrypoint
CMD sleep infinity
