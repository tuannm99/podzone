package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/model"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.uber.org/fx"
)

var _ outputport.UserRepository = (*UserRepositoryImpl)(nil)

type UserRepoParams struct {
	fx.In
	Logger pdlog.Logger
	DB     *sqlx.DB `name:"sql-auth"`
}

func NewUserRepositoryImpl(p UserRepoParams) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		logger: p.Logger,
		db:     p.DB,
	}
}

type UserRepositoryImpl struct {
	logger pdlog.Logger
	db     *sqlx.DB `name:"sql-auth"`
}

// -------------------- Squirrel setup --------------------

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var userColumns = []string{
	"id",
	"username",
	"email",
	"password",
	"full_name",
	"middle_name",
	"first_name",
	"last_name",
	"address",
	"initial_from",
	"age",
	"dob",
	"created_at",
	"updated_at",
}

// -------------------- helpers --------------------

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// lib/pq or pgx both include similar phrases
	return strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint")
}

func hashIfSet(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	return entity.GeneratePasswordHash(plain)
}

// -------------------- repository methods --------------------

// GetByUsernameOrEmail implements outputport.UserRepository.
func (u *UserRepositoryImpl) GetByUsernameOrEmail(identity string) (*entity.User, error) {
	ctx := context.Background()

	qb := psql.
		Select(userColumns...).
		From("users").
		Where(sq.Or{
			sq.Eq{"email": identity},
			sq.Eq{"username": identity},
		}).
		Limit(1)

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, err
	}

	var m model.User
	if err := u.db.GetContext(ctx, &m, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}
	return m.ToEntity()
}

// Create implements outputport.UserRepository.
func (u *UserRepositoryImpl) Create(e entity.User) (*entity.User, error) {
	ctx := context.Background()

	m, err := toolkit.MapStruct[entity.User, model.User](e)
	if err != nil {
		return nil, err
	}

	if m.Password != "" {
		hashed, err := hashIfSet(m.Password)
		if err != nil {
			return nil, err
		}
		m.Password = hashed
	}
	now := time.Now().UTC()
	m.CreatedAt = now
	m.UpdatedAt = now

	qb := psql.
		Insert("users").
		Columns(
			"username", "email", "password",
			"full_name", "middle_name", "first_name", "last_name",
			"address", "initial_from", "age", "dob",
			"created_at", "updated_at",
		).
		Values(
			m.Username, m.Email, m.Password,
			m.FullName, m.MiddleName, m.FirstName, m.LastName,
			m.Address, m.InitialFrom, m.Age, m.Dob,
			m.CreatedAt, m.UpdatedAt,
		).
		Suffix("RETURNING " + strings.Join(userColumns, ","))

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, err
	}

	var out model.User
	if err := u.db.QueryRowxContext(ctx, query, args...).StructScan(&out); err != nil {
		if isUniqueViolation(err) {
			return nil, entity.ErrUserAlreadyExists
		}
		return nil, err
	}
	return out.ToEntity()
}

// CreateByEmailIfNotExisted implements outputport.UserRepository.
// INSERT ... ON CONFLICT DO NOTHING RETURNING ... ; if no row returned, SELECT one.
func (u *UserRepositoryImpl) CreateByEmailIfNotExisted(email string) (*entity.User, error) {
	ctx := context.Background()

	qb := psql.
		Insert("users").
		Columns("email", "initial_from", "created_at", "updated_at").
		Values(email, "google", sq.Expr("NOW()"), sq.Expr("NOW()")).
		Suffix("ON CONFLICT (email) DO NOTHING RETURNING " + strings.Join(userColumns, ","))

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, err
	}

	var out model.User
	if err := u.db.QueryRowxContext(ctx, query, args...).StructScan(&out); err == nil {
		return out.ToEntity()
	}
	return u.GetByUsernameOrEmail(email)
}

// Update implements outputport.UserRepository.
// Strategy: SET only provided (non-zero) fields; hash password when provided.
func (u *UserRepositoryImpl) Update(usr entity.User) error {
	if usr.Id == 0 {
		return entity.ErrUserNotFound
	}

	ctx := context.Background()

	set := sq.Eq{"updated_at": time.Now().UTC()}

	if usr.Username != "" {
		set["username"] = usr.Username
	}
	if usr.Email != "" {
		set["email"] = usr.Email
	}
	if usr.FullName != "" {
		set["full_name"] = usr.FullName
	}
	if usr.MiddleName != "" {
		set["middle_name"] = usr.MiddleName
	}
	if usr.FirstName != "" {
		set["first_name"] = usr.FirstName
	}
	if usr.LastName != "" {
		set["last_name"] = usr.LastName
	}
	if usr.Address != "" {
		set["address"] = usr.Address
	}
	if usr.InitialFrom != "" {
		set["initial_from"] = usr.InitialFrom
	}
	if usr.Age != 0 {
		set["age"] = usr.Age
	}
	if !usr.Dob.IsZero() {
		set["dob"] = usr.Dob
	}
	if usr.Password != "" {
		hashed, err := hashIfSet(usr.Password)
		if err != nil {
			return err
		}
		set["password"] = hashed
	}

	// nothing to update?
	if len(set) == 1 { // only updated_at
		return nil
	}

	qb := psql.
		Update("users").
		SetMap(set).
		Where(sq.Eq{"id": usr.Id})

	query, args, err := qb.ToSql()
	if err != nil {
		return err
	}

	res, err := u.db.ExecContext(ctx, query, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return entity.ErrUserAlreadyExists
		}
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return entity.ErrUserNotFound
	}
	return nil
}

// GetByID implements outputport.UserRepository.
func (u *UserRepositoryImpl) GetByID(id string) (*entity.User, error) {
	ctx := context.Background()

	qb := psql.
		Select(userColumns...).
		From("users").
		Where(sq.Eq{"id": id}).
		Limit(1)

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, err
	}

	var m model.User
	if err := u.db.GetContext(ctx, &m, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}
	return m.ToEntity()
}

// UpdateById implements outputport.UserRepository.
func (u *UserRepositoryImpl) UpdateById(id uint, user entity.User) error {
	user.Id = id
	return u.Update(user)
}
