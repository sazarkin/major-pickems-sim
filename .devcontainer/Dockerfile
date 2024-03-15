FROM ndunnett/python:latest

# install dependencies
RUN set -eux; \
    DEBIAN_FRONTEND="noninteractive"; \
    apt-get update; \
    apt-get install -y wget git zsh; \
    rm -rf /var/lib/apt/lists/*
ARG USERNAME=dev

# create user
RUN set -eux; \
    useradd --create-home --user-group --no-log-init "$USERNAME"; \
    mkdir -p "/home/$USERNAME/src"; \
    chown -R "$USERNAME:$USERNAME" "/home/$USERNAME"
USER "$USERNAME"

# entrypoint
CMD ["sleep", "infinity"]
