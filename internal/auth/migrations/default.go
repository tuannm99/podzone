// TODO: using goose for migration tool for sql
// go install github.com/pressly/goose/v3/cmd/goose@latest
package migrations

import (
	sq "github.com/Masterminds/squirrel"
)

// Optional: enable extensions (Postgres)
var CreateExts = []sq.Sqlizer{
	// sq.Expr(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`),
	// sq.Expr(`CREATE EXTENSION IF NOT EXISTS "citext"`),
}

var CreateTableUsers = []sq.Sqlizer{
	sq.Expr(`
CREATE TABLE IF NOT EXISTS users (
  id           BIGSERIAL PRIMARY KEY,
  username     TEXT NOT NULL DEFAULT '' UNIQUE,
  email        TEXT NOT NULL DEFAULT '' UNIQUE,
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
	sq.Expr(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`),
	sq.Expr(`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`),
}
