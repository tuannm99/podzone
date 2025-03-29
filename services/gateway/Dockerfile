FROM kong:3.3.0

# Switch to root user to install dependencies
USER root

# Install build dependencies
RUN apt-get update && \
    apt-get install -y build-essential && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install grpc-gateway plugin
RUN luarocks install kong-plugin-grpc-gateway

# Create directory for custom plugins
RUN mkdir -p /usr/local/kong/plugins/tenant-identifier
COPY ./plugins/tenant-identifier/ /usr/local/kong/plugins/tenant-identifier/

# Set the environment variables
ENV KONG_PLUGINS=bundled,grpc-gateway,tenant-identifier
ENV KONG_LUA_PACKAGE_PATH=/usr/local/kong/plugins/?.lua;;

# Switch back to kong user
USER kong

HEALTHCHECK --interval=10s --timeout=10s --retries=3 CMD kong health
