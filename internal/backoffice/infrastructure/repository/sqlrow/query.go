package sqlrow

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

// QueryMany collapses the ToSql -> err -> WithTenantTx -> Select sandwich
// repeated at nearly every read call site in this repository layer. Only
// fits the plain "run this SELECT, scan every row" case - call sites with
// their own tx logic (multiple statements, writes, custom not-found errors)
// keep using mgr.WithTenantTx directly. See docs/07-problems/
// repository-query-layer-2026-07-24.md RQ4 step 2.
func QueryMany[Row any](ctx context.Context, mgr pdtenantdb.Manager, b sq.SelectBuilder) ([]Row, error) {
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}
	query, args, err := b.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []Row
	if err := mgr.WithTenantTx(ctx, tenantID, nil, func(tx *sqlx.Tx) error {
		return tx.SelectContext(ctx, &rows, query, args...)
	}); err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryOne is QueryMany's single-row counterpart. Returns (nil, nil) - not
// an error - when no row matches, matching sql.ErrNoRows-as-empty-result
// call sites. Call sites that need a not-found error instead keep their own
// mgr.WithTenantTx + errors.Is(err, sql.ErrNoRows) handling.
func QueryOne[Row any](ctx context.Context, mgr pdtenantdb.Manager, b sq.SelectBuilder) (*Row, error) {
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}
	query, args, err := b.ToSql()
	if err != nil {
		return nil, err
	}

	var row Row
	found := false
	if err := mgr.WithTenantTx(ctx, tenantID, nil, func(tx *sqlx.Tx) error {
		if err := tx.GetContext(ctx, &row, query, args...); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}
		found = true
		return nil
	}); err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &row, nil
}
