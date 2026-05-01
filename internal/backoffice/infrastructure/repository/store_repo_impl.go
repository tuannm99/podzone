package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/outputport"
	"github.com/tuannm99/podzone/internal/backoffice/infrastructure/model"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type StoreRepositoryImpl struct {
	mgr pdtenantdb.Manager
}

func NewStoreRepository(mgr pdtenantdb.Manager) outputport.StoreRepository {
	return &StoreRepositoryImpl{mgr: mgr}
}

const ensureStoresTableSQL = `
CREATE TABLE IF NOT EXISTS stores (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	owner_id TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
)`

func (r *StoreRepositoryImpl) FindAll(ctx context.Context) ([]entity.Store, error) {
	ownerID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	query, args, err := psql.
		Select("id", "name", "description", "owner_id", "status", "created_at", "updated_at").
		From(model.Store{}.TableName()).
		Where(sq.Eq{"owner_id": ownerID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var stores []model.Store
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureStoreTables(ctx, tx); err != nil {
			return err
		}
		return tx.SelectContext(ctx, &stores, query, args...)
	}); err != nil {
		return nil, err
	}

	res := make([]entity.Store, 0, len(stores))
	for _, s := range stores {
		res = append(res, toEntity(s))
	}
	return res, nil
}

func (r *StoreRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.Store, error) {
	ownerID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	query, args, err := psql.
		Select("id", "name", "description", "owner_id", "status", "created_at", "updated_at").
		From(model.Store{}.TableName()).
		Where(sq.Eq{"id": id, "owner_id": ownerID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var s model.Store
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureStoreTables(ctx, tx); err != nil {
			return err
		}
		if err := tx.GetContext(ctx, &s, query, args...); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("store not found")
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	out := toEntity(s)
	return &out, nil
}

func (r *StoreRepositoryImpl) Create(ctx context.Context, s *model.Store) error {
	query, args, err := psql.
		Insert(s.TableName()).
		Columns("id", "name", "description", "owner_id", "status", "created_at", "updated_at").
		Values(s.ID, s.Name, s.Description, s.OwnerID, s.Status, s.CreatedAt, s.UpdatedAt).
		ToSql()
	if err != nil {
		return err
	}

	return r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureStoreTables(ctx, tx); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, query, args...)
		return err
	})
}

func (r *StoreRepositoryImpl) UpdateStatus(ctx context.Context, id string, status model.StoreStatus) error {
	ownerID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return err
	}

	query, args, err := psql.
		Update(model.Store{}.TableName()).
		Set("status", status).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id, "owner_id": ownerID}).
		ToSql()
	if err != nil {
		return err
	}

	return r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureStoreTables(ctx, tx); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			return errors.New("store not found")
		}
		return nil
	})
}

func (r *StoreRepositoryImpl) withTenantTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return err
	}
	return r.mgr.WithTenantTx(ctx, tenantID, nil, fn)
}

func ensureStoreTables(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, ensureStoresTableSQL)
	return err
}

func toEntity(s model.Store) entity.Store {
	return entity.Store{
		ID:          s.ID,
		Name:        s.Name,
		OwnerID:     s.OwnerID,
		IsActive:    s.Status == model.StoreStatusActive,
		Description: s.Description,
		Status:      string(s.Status),
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}
