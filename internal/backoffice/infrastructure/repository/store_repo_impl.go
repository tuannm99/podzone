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

func (r *StoreRepositoryImpl) FindAll(ctx context.Context) ([]entity.Store, error) {
	query, args, err := psql.
		Select("id", "name", "description", "owner_id", "status", "created_at", "updated_at").
		From(model.Store{}.TableName()).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var stores []model.Store
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		return tx.SelectContext(ctx, &stores, query, args...)
	}); err != nil {
		return nil, err
	}

	res := make([]entity.Store, 0, len(stores))
	for _, s := range stores {
		res = append(res, entity.Store{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			IsActive:    s.Status == model.StoreStatusActive,
		})
	}
	return res, nil
}

func (r *StoreRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.Store, error) {
	query, args, err := psql.
		Select("id", "name", "description", "owner_id", "status", "created_at", "updated_at").
		From(model.Store{}.TableName()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var s model.Store
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
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

	return &entity.Store{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		IsActive:    s.Status == model.StoreStatusActive,
	}, nil
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
		_, err = tx.ExecContext(ctx, query, args...)
		return err
	})
}

func (r *StoreRepositoryImpl) UpdateStatus(ctx context.Context, id string, status model.StoreStatus) error {
	query, args, err := psql.
		Update(model.Store{}.TableName()).
		Set("status", status).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	return r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
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
