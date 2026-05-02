SELECT 'CREATE DATABASE supplier'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'supplier')\gexec

