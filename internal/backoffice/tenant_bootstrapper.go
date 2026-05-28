package backoffice

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/internal/backoffice/migrations"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
)

type TenantBootstrapper interface {
	EnsureReady(ctx context.Context, tenantID string) error
}

type MigrationTenantBootstrapper struct {
	mgr pdtenantdb.Manager
}

func NewTenantBootstrapper(mgr pdtenantdb.Manager) *MigrationTenantBootstrapper {
	return &MigrationTenantBootstrapper{mgr: mgr}
}

func (b *MigrationTenantBootstrapper) EnsureReady(ctx context.Context, tenantID string) error {
	return b.mgr.WithTenantTx(ctx, tenantID, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return migrations.ApplyTx(ctx, tx)
	})
}
