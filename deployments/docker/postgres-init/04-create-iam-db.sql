SELECT 'CREATE DATABASE iam'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'iam')\gexec
