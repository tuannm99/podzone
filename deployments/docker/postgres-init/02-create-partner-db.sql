SELECT 'CREATE DATABASE partner'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'partner')\gexec
