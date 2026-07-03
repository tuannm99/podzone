package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
	"github.com/tuannm99/podzone/internal/backoffice/migrations"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

const storesTable = "stores"

type storeRow struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	OwnerID     string    `db:"owner_id"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Repository struct {
	mgr pdtenantdb.Manager
}

var _ storectx.StoreRepository = (*Repository)(nil)

func New(mgr pdtenantdb.Manager) storectx.StoreRepository {
	return &Repository{mgr: mgr}
}

func (r *Repository) FindPage(
	ctx context.Context,
	collectionQuery collection.Query,
) (collection.Page[storectx.Store], error) {
	ownerID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return collection.Page[storectx.Store]{}, err
	}
	normalized, predicates, orderBy, err := buildStoreCollectionQuery(collectionQuery)
	if err != nil {
		return collection.Page[storectx.Store]{}, err
	}
	predicates = append([]sq.Sqlizer{sq.Eq{"owner_id": ownerID}}, predicates...)
	var stores []storeRow
	var total int64
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureStoreTables(ctx, tx); err != nil {
			return err
		}
		countBuilder := psql.Select("COUNT(*)").From(storesTable)
		listBuilder := psql.
			Select("id", "name", "description", "owner_id", "status", "created_at", "updated_at").
			From(storesTable)
		for _, predicate := range predicates {
			countBuilder = countBuilder.Where(predicate)
			listBuilder = listBuilder.Where(predicate)
		}
		countSQL, countArgs, err := countBuilder.ToSql()
		if err != nil {
			return err
		}
		if err := tx.GetContext(ctx, &total, countSQL, countArgs...); err != nil {
			return err
		}
		listSQL, listArgs, err := listBuilder.
			OrderBy(orderBy, "id ASC").
			Limit(uint64(normalized.PageSize)).
			Offset(uint64(normalized.Offset())).
			ToSql()
		if err != nil {
			return err
		}
		return tx.SelectContext(ctx, &stores, listSQL, listArgs...)
	}); err != nil {
		return collection.Page[storectx.Store]{}, err
	}

	res := make([]storectx.Store, 0, len(stores))
	for _, s := range stores {
		res = append(res, toEntity(s))
	}
	return collection.NewPage(res, total, normalized), nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*storectx.Store, error) {
	ownerID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	query, args, err := psql.
		Select("id", "name", "description", "owner_id", "status", "created_at", "updated_at").
		From(storesTable).
		Where(sq.Eq{"id": id, "owner_id": ownerID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var s storeRow
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

func (r *Repository) Create(ctx context.Context, store storectx.Store) error {
	query, args, err := psql.
		Insert(storesTable).
		Columns("id", "name", "description", "owner_id", "status", "created_at", "updated_at").
		Values(store.ID, store.Name, store.Description, store.OwnerID, store.Status, store.CreatedAt, store.UpdatedAt).
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

func (r *Repository) Bootstrap(ctx context.Context, store storectx.Store) error {
	query, args, err := psql.
		Insert(storesTable).
		Columns("id", "name", "description", "owner_id", "status", "created_at", "updated_at").
		Values(store.ID, store.Name, store.Description, store.OwnerID, store.Status, store.CreatedAt, store.UpdatedAt).
		Suffix(`
ON CONFLICT (id) DO UPDATE SET
	name = EXCLUDED.name,
	owner_id = EXCLUDED.owner_id,
	status = EXCLUDED.status,
	updated_at = EXCLUDED.updated_at`).
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

func (r *Repository) UpdateStatus(ctx context.Context, id string, status string) error {
	ownerID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return err
	}

	query, args, err := psql.
		Update(storesTable).
		Set("status", status).
		Set("updated_at", time.Now().UTC()).
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

func (r *Repository) withTenantTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return err
	}
	return r.mgr.WithTenantTx(ctx, tenantID, nil, fn)
}

func ensureStoreTables(ctx context.Context, tx *sqlx.Tx) error {
	return migrations.ApplyTx(ctx, tx)
}

func toEntity(s storeRow) storectx.Store {
	return storectx.Store{
		ID:          s.ID,
		Name:        s.Name,
		OwnerID:     s.OwnerID,
		IsActive:    s.Status == storectx.StoreStatusActive,
		Description: s.Description,
		Status:      s.Status,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}
