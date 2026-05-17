SELECT 'CREATE DATABASE backoffice'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'backoffice')\gexec
