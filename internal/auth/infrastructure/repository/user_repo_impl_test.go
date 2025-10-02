package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/migrations"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

type nopLogger = pdlog.NopLogger

func execDDL(ctx context.Context, db *sqlx.DB, steps ...sq.Sqlizer) error {
	for _, s := range steps {
		sqlStr, args, err := s.ToSql()
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, sqlStr, args...); err != nil {
			return err
		}
	}
	return nil
}

func setupPostgres(t require.TestingT) (*sqlx.DB, func()) {
	ctx := context.Background()

	pgC, err := tcpg.RunContainer(ctx,
		tcpg.WithDatabase("testdb"),
		tcpg.WithUsername("testuser"),
		tcpg.WithPassword("testpass"),
	)
	require.NoError(t, err)

	dsn, err := pgC.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	var db *sqlx.DB
	const (
		maxAttempts    = 30
		initialBackoff = 200 * time.Millisecond
		maxBackoff     = 2 * time.Second
	)
	backoff := initialBackoff
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err = sqlx.Connect("postgres", dsn)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			pErr := db.PingContext(pingCtx)
			cancel()
			if pErr == nil {
				break
			}
			err = pErr
		}
		if attempt == maxAttempts {
			require.NoError(t, err, "postgres not ready after retries")
		}
		time.Sleep(backoff)
		if backoff *= 2; backoff > maxBackoff {
			backoff = maxBackoff
		}
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, execDDL(ctx, db, append(migrations.CreateExts, migrations.CreateTableUsers...)...))
	cleanup := func() {
		_ = db.Close()
		_ = pgC.Terminate(context.Background())
	}
	return db, cleanup
}

func newUserRepo(t *testing.T, db *sqlx.DB) *UserRepositoryImpl {
	t.Helper()
	return NewUserRepositoryImpl(UserRepoParams{
		Logger: nopLogger{},
		DB:     db,
	})
}

func TestUserRepository_Postgres_Integration_CRUD(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	repo := newUserRepo(t, db)

	u, err := repo.Create(entity.User{
		Username: "jdoe",
		Email:    "jdoe@example.com",
		Password: "secret",
		FullName: "John Doe",
		Age:      30,
	})
	require.NoError(t, err)
	require.NotNil(t, u)
	require.NotZero(t, u.Id)
	require.Equal(t, "jdoe@example.com", u.Email)
	require.NotEmpty(t, u.Password) // hashed
	require.NotEqual(t, "secret", u.Password)

	g1, err := repo.GetByUsernameOrEmail("jdoe@example.com")
	require.NoError(t, err)
	require.Equal(t, u.Id, g1.Id)

	g2, err := repo.GetByUsernameOrEmail("jdoe")
	require.NoError(t, err)
	require.Equal(t, u.Id, g2.Id)

	_, err = repo.Create(entity.User{
		Username: "another",
		Email:    "jdoe@example.com",
	})
	require.Error(t, err)

	u2, err := repo.CreateByEmailIfNotExisted("new@example.com")
	require.NoError(t, err)
	require.NotNil(t, u2)
	require.Equal(t, "new@example.com", u2.Email)
	require.Equal(t, "google", u2.InitialFrom)

	u3, err := repo.CreateByEmailIfNotExisted("new@example.com")
	require.NoError(t, err)
	require.NotNil(t, u3)
	require.Equal(t, u2.Id, u3.Id)

	err = repo.Update(entity.User{
		Id:          u.Id,
		FullName:    "John X. Doe",
		Password:    "changed",
		Age:         31,
		InitialFrom: "podzone",
	})
	require.NoError(t, err)

	got, err := repo.GetByID(toStr(u.Id))
	require.NoError(t, err)
	require.Equal(t, "John X. Doe", got.FullName)
	require.Equal(t, uint8(31), got.Age)
	require.Equal(t, "podzone", got.InitialFrom)
	require.NotEqual(t, "changed", got.Password)
	require.NotEmpty(t, got.Password)

	_, err = repo.GetByID("999999999")
	require.Error(t, err)
}

func toStr(id uint) string { return fmt.Sprintf("%d", id) }
