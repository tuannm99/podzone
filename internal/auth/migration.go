package auth

import (
	sq "github.com/Masterminds/squirrel"
)

// Optional: enable extensions (Postgres)
var CreateExts = []sq.Sqlizer{
	// sq.Expr(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`),
	// sq.Expr(`CREATE EXTENSION IF NOT EXISTS "citext"`),
}

var CreateUsers = []sq.Sqlizer{
	sq.Expr(`
CREATE TABLE IF NOT EXISTS users (
  id           BIGSERIAL PRIMARY KEY,
  username     TEXT UNIQUE,
  email        TEXT UNIQUE,
  password     TEXT NOT NULL DEFAULT '',
  full_name    TEXT NOT NULL DEFAULT '',
  middle_name  TEXT NOT NULL DEFAULT '',
  first_name   TEXT NOT NULL DEFAULT '',
  last_name    TEXT NOT NULL DEFAULT '',
  address      TEXT NOT NULL DEFAULT '',
  initial_from TEXT NOT NULL DEFAULT '',
  age          SMALLINT NOT NULL DEFAULT 0,
  dob          TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
)`),
	// Useful indexes (idempotent)
	sq.Expr(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`),
	sq.Expr(`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`),
}
