local TenantIdentifier = {
    PRIORITY = 1000,
    VERSION = "1.0.0",
}

-- Kong requires these standard phase handlers, even if they do nothing
function TenantIdentifier:init_worker()
end

function TenantIdentifier:preread()
end

function TenantIdentifier:certificate()
end

function TenantIdentifier:rewrite()
end

function TenantIdentifier:access(conf)
    local tenant_id
    -- Try to extract from subdomain
    local host = kong.request.get_host()
    local subdomain = string.match(host, "^([^.]+)%.")
    if subdomain and subdomain ~= "www" and subdomain ~= "api" then
        tenant_id = subdomain
    end
    -- If not found, try from path
    if not tenant_id then
        local path = kong.request.get_path()
        local tenant_from_path = string.match(path, "^/tenants/([^/]+)")
        if tenant_from_path then
            tenant_id = tenant_from_path
            -- Optionally rewrite the path to remove tenant prefix
            local new_path = string.gsub(path, "^/tenants/[^/]+", "")
            if new_path == "" then new_path = "/" end
            kong.service.request.set_path(new_path)
        end
    end
    -- If still not found, try from header
    if not tenant_id then
        tenant_id = kong.request.get_header("X-Tenant-ID")
    end
    if tenant_id then
        -- Set the tenant ID as a header for downstream services
        kong.service.request.set_header("X-Tenant-ID", tenant_id)
        -- Add to request context
        kong.ctx.shared.tenant_id = tenant_id
        -- Log the tenant ID (optional)
        kong.log.debug("Tenant ID: ", tenant_id)
    else
        if conf.required then
            return kong.response.exit(400, { message = "Tenant ID is required" })
        end
    end
end

function TenantIdentifier:header_filter(conf)
end

function TenantIdentifier:body_filter(conf)
end

function TenantIdentifier:log(conf)
end

return TenantIdentifier
