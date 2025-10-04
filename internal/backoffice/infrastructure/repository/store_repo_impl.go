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
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type StoreRepositoryImpl struct {
	db *sqlx.DB
}

func NewStoreRepository(db *sqlx.DB) outputport.StoreRepository {
	return &StoreRepositoryImpl{db: db}
}

func (r *StoreRepositoryImpl) FindAll() ([]entity.Store, error) {
	query, args, err := psql.
		Select("id", "name", "description", "owner_id", "status", "created_at", "updated_at").
		From(model.Store{}.TableName()).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var stores []model.Store
	if err := r.db.Select(&stores, query, args...); err != nil {
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

func (r *StoreRepositoryImpl) FindByID(id string) (*entity.Store, error) {
	query, args, err := psql.
		Select("id", "name", "description", "owner_id", "status", "created_at", "updated_at").
		From(model.Store{}.TableName()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var s model.Store
	if err := r.db.Get(&s, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("store not found")
		}
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

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
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

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("store not found")
	}
	return nil
}
