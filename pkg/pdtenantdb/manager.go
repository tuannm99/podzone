package pdtenantdb

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Manager interface {
	GetDB(ctx context.Context, tenantID string) (*sqlx.DB, error)
	CloseAll() error
}
