FROM kong:3.9.0

# Switch to root user to install dependencies
USER root

# Install build dependencies
RUN apt-get update && \
    apt-get install -y build-essential && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*;

# Create directory for our custom plugins
RUN mkdir -p /usr/local/kong/plugins/tenant-identifier

COPY ./plugins/tenant-identifier/ /usr/local/kong/plugins/tenant-identifier/

RUN chmod -R 755 /usr/local/kong/plugins/tenant-identifier/ && \
    chown -R kong:kong /usr/local/kong/plugins/tenant-identifier/

RUN mkdir -p /usr/local/share/lua/5.1/kong/plugins/tenant-identifier && \
    cp /usr/local/kong/plugins/tenant-identifier/handler.lua /usr/local/share/lua/5.1/kong/plugins/tenant-identifier/ && \
    cp /usr/local/kong/plugins/tenant-identifier/schema.lua /usr/local/share/lua/5.1/kong/plugins/tenant-identifier/

# Create rockspec file for the custom plugin
RUN echo 'package = "kong-plugin-tenant-identifier"' > /tmp/kong-plugin-tenant-identifier-1.0.0-1.rockspec && \
    echo 'version = "1.0.0-1"' >> /tmp/kong-plugin-tenant-identifier-1.0.0-1.rockspec && \
    echo 'source = { url = "git://github.com/tuannm99/podzone" }' >> /tmp/kong-plugin-tenant-identifier-1.0.0-1.rockspec && \
    echo 'description = { summary = "Kong tenant identifier plugin" }' >> /tmp/kong-plugin-tenant-identifier-1.0.0-1.rockspec && \
    echo 'dependencies = { "lua >= 5.1" }' >> /tmp/kong-plugin-tenant-identifier-1.0.0-1.rockspec && \
    echo 'build = { type = "builtin", modules = { ["kong.plugins.tenant-identifier.handler"] = "/usr/local/kong/plugins/tenant-identifier/handler.lua", ["kong.plugins.tenant-identifier.schema"] = "/usr/local/kong/plugins/tenant-identifier/schema.lua" } }' >> '/tmp/kong-plugin-tenant-identifier-1.0.0-1.rockspec'

# Install the custom plugin via LuaRocks
RUN luarocks install '/tmp/kong-plugin-tenant-identifier-1.0.0-1.rockspec'

RUN luarocks install kong-plugin-grpc-gateway


# Set the environment variables
ENV KONG_PLUGINS="tenant-identifier,bundled,grpc-gateway"
ENV KONG_LUA_PACKAGE_PATH="/usr/local/kong/?.lua;/usr/local/kong/plugins/?.lua;/usr/local/kong/plugins/tenant-identifier/?.lua;;"

# Switch back to kong user
USER kong

HEALTHCHECK --interval=10s --timeout=10s --retries=3 CMD kong health
