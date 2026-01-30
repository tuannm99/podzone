package repository

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/migrations"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/testkit"
)

var (
	migrateOnce sync.Once
	errMigrate  error
)

type nopLogger = pdlog.NopLogger

func setupRepo(t *testing.T) (*UserRepositoryImpl, *sqlx.DB) {
	t.Helper()

	db := testkit.PostgresDB(t)

	migrateOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		errMigrate = migrations.Apply(ctx, db.DB, "postgres")
	})
	require.NoError(t, errMigrate)

	truncateUsers(t, db)

	repo := NewUserRepositoryImpl(UserRepoParams{
		Logger: nopLogger{},
		DB:     db,
	})
	return repo, db
}

func truncateUsers(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE TABLE users RESTART IDENTITY`)
	require.NoError(t, err)
}

func insertUser(t *testing.T, db *sqlx.DB, u entity.User) uint {
	t.Helper()
	if u.Username == "" {
		u.Username = "user"
	}
	if u.Email == "" {
		u.Email = "user@example.com"
	}
	if u.Password == "" {
		u.Password = "pw"
	}

	var id uint
	err := db.Get(
		&id,
		`INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id`,
		u.Username,
		u.Email,
		u.Password,
	)
	require.NoError(t, err)
	return id
}

func Test_isUniqueViolation_Table(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"duplicate key", errors.New("duplicate key value violates unique constraint"), true},
		{"unique constraint", errors.New("unique constraint"), true},
		{"other", errors.New("other"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isUniqueViolation(tc.err))
		})
	}
}

func TestUserRepository_GetByUsernameOrEmail(t *testing.T) {
	repo, db := setupRepo(t)

	id := insertUser(t, db, entity.User{Username: "jdoe", Email: "a@b.com"})

	u, err := repo.GetByUsernameOrEmail("a@b.com")
	require.NoError(t, err)
	require.NotNil(t, u)
	require.Equal(t, id, u.Id)
}

func TestUserRepository_GetByUsernameOrEmail_NotFound(t *testing.T) {
	repo, _ := setupRepo(t)

	u, err := repo.GetByUsernameOrEmail("missing")
	require.ErrorIs(t, err, entity.ErrUserNotFound)
	require.Nil(t, u)
}

func TestUserRepository_GetByID(t *testing.T) {
	repo, db := setupRepo(t)

	id := insertUser(t, db, entity.User{Username: "jdoe", Email: "jdoe@example.com"})

	u, err := repo.GetByID(strconv.FormatUint(uint64(id), 10))
	require.NoError(t, err)
	require.NotNil(t, u)
	require.Equal(t, id, u.Id)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	repo, _ := setupRepo(t)

	u, err := repo.GetByID("9999")
	require.ErrorIs(t, err, entity.ErrUserNotFound)
	require.Nil(t, u)
}

func TestUserRepository_Create(t *testing.T) {
	repo, _ := setupRepo(t)

	out, err := repo.Create(entity.User{
		Username: "jdoe",
		Email:    "jdoe@example.com",
		Password: "secret",
		FullName: "John Doe",
		Age:      30,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotZero(t, out.Id)
}

func TestUserRepository_Create_UniqueViolation(t *testing.T) {
	repo, db := setupRepo(t)

	_ = insertUser(t, db, entity.User{Username: "dup", Email: "dup@example.com"})

	out, err := repo.Create(entity.User{Username: "dup2", Email: "dup@example.com"})
	require.ErrorIs(t, err, entity.ErrUserAlreadyExists)
	require.Nil(t, out)
}

func TestUserRepository_Update_EarlyReturn(t *testing.T) {
	repo, _ := setupRepo(t)

	require.ErrorIs(t, repo.Update(entity.User{Id: 0}), entity.ErrUserNotFound)
	require.NoError(t, repo.Update(entity.User{Id: 1}))
}

func TestUserRepository_Update(t *testing.T) {
	repo, db := setupRepo(t)

	id := insertUser(t, db, entity.User{Username: "jdoe", Email: "jdoe@example.com"})

	err := repo.Update(entity.User{
		Id:       id,
		FullName: "John X",
		Age:      31,
	})
	require.NoError(t, err)

	var fullName string
	var age int
	err = db.Get(&fullName, `SELECT full_name FROM users WHERE id=$1`, id)
	require.NoError(t, err)
	require.Equal(t, "John X", fullName)

	err = db.Get(&age, `SELECT age FROM users WHERE id=$1`, id)
	require.NoError(t, err)
	require.Equal(t, 31, age)
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	repo, _ := setupRepo(t)

	err := repo.Update(entity.User{Id: 9999, Email: "x@example.com"})
	require.ErrorIs(t, err, entity.ErrUserNotFound)
}

func TestUserRepository_Update_UniqueViolation(t *testing.T) {
	repo, db := setupRepo(t)

	id1 := insertUser(t, db, entity.User{Username: "u1", Email: "u1@example.com"})
	_ = insertUser(t, db, entity.User{Username: "u2", Email: "u2@example.com"})

	err := repo.Update(entity.User{Id: id1, Email: "u2@example.com"})
	require.ErrorIs(t, err, entity.ErrUserAlreadyExists)
}

func TestUserRepository_UpdateById(t *testing.T) {
	repo, db := setupRepo(t)

	id := insertUser(t, db, entity.User{Username: "u", Email: "u@example.com"})

	err := repo.UpdateById(id, entity.User{FullName: "A"})
	require.NoError(t, err)

	var fullName string
	err = db.Get(&fullName, `SELECT full_name FROM users WHERE id=$1`, id)
	require.NoError(t, err)
	require.Equal(t, "A", fullName)
}
