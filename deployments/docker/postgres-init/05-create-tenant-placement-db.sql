SELECT 'CREATE DATABASE podzone_tenants'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'podzone_tenants')\gexec
