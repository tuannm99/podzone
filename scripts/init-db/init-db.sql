SELECT 'CREATE DATABASE podzone' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'podzone');
