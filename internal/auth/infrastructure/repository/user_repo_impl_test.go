package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

var testLogger = &pdlog.NopLogger{}

func newSQLXMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	raw, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := sqlx.NewDb(raw, "sqlmock")
	cleanup := func() { _ = db.Close() }
	return db, mock, cleanup
}

func userCols() []string {
	return []string{
		"id", "username", "email", "password",
		"full_name", "middle_name", "first_name", "last_name",
		"address", "initial_from", "age", "dob",
		"created_at", "updated_at",
	}
}

func makeUserRow() *sqlmock.Rows {
	now := time.Now().UTC()
	return sqlmock.NewRows(userCols()).
		AddRow(
			int64(1),
			"jdoe",
			"jdoe@example.com",
			"hashed-pass",
			"John Doe",
			"", "John", "Doe",
			"Somewhere", "google",
			int64(30), now, now, now,
		)
}

// ---------------- GetByUsernameOrEmail ----------------

func TestGetByUsernameOrEmail_Found(t *testing.T) {
	sqlxDB, mock, cleanup := newSQLXMock(t)
	defer cleanup()

	rawQ := "SELECT id, username, email, password, full_name, middle_name, first_name, last_name, address, initial_from, age, dob, created_at, updated_at FROM users WHERE (email = $1 OR username = $2) LIMIT 1"
	mock.ExpectQuery(regexp.QuoteMeta(rawQ)).
		WithArgs("jdoe@example.com", "jdoe@example.com").
		WillReturnRows(makeUserRow())

	repo := NewUserRepositoryImpl(UserRepoParams{Logger: testLogger, DB: sqlxDB})
	u, err := repo.GetByUsernameOrEmail("jdoe@example.com")
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "jdoe@example.com", u.Email)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByUsernameOrEmail_NotFound(t *testing.T) {
	sqlxDB, mock, cleanup := newSQLXMock(t) // matcher mặc định
	defer cleanup()

	identity := "nope@example.com"

	query, _, err := psql.
		Select(userColumns...).
		From("users").
		Where(sq.Or{
			sq.Eq{"email": identity},
			sq.Eq{"username": identity},
		}).
		Limit(1).
		ToSql()
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(identity, identity).
		WillReturnError(sql.ErrNoRows)

	repo := NewUserRepositoryImpl(UserRepoParams{Logger: testLogger, DB: sqlxDB})
	u, err := repo.GetByUsernameOrEmail(identity)
	require.Error(t, err)
	assert.Nil(t, u)

	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------------- Create ----------------

// func TestCreate_Success(t *testing.T) {
// 	sqlxDB, mock, cleanup := newSQLXMock(t)
// 	defer cleanup()
//
// 	re := regexp.MustCompile(`INSERT INTO users .* RETURNING .*`)
// 	// ExpectQuery vì dùng QueryRowx (RETURNING)
// 	mock.ExpectQuery(re.String()).
// 		WithArgs(
// 			"jdoe", "jdoe@example.com", // username, email
// 			sqlmock.AnyArg(),              // password (hashed)
// 			"John Doe", "", "John", "Doe", // names
// 			"Somewhere", "google",
// 			int64(30),
// 			sqlmock.AnyArg(),                   // dob time.Time
// 			sqlmock.AnyArg(), sqlmock.AnyArg(), // created_at, updated_at
// 		).
// 		WillReturnRows(makeUserRow())
//
// 	repo := NewUserRepositoryImpl(UserRepoParams{Logger: testLogger, DB: sqlxDB})
// 	u, err := repo.Create(entity.User{
// 		Username:    "jdoe",
// 		Email:       "jdoe@example.com",
// 		Password:    "plain-pass",
// 		FullName:    "John Doe",
// 		FirstName:   "John",
// 		LastName:    "Doe",
// 		Address:     "Somewhere",
// 		InitialFrom: "google",
// 		Age:         30,
// 		Dob:         time.Now().UTC(),
// 	})
// 	require.NoError(t, err)
// 	require.NotNil(t, u)
//
// 	require.NoError(t, mock.ExpectationsWereMet())
// }
//
// func TestCreate_UniqueViolation(t *testing.T) {
// 	sqlxDB, mock, cleanup := newSQLXMock(t)
// 	defer cleanup()
//
// 	re := regexp.MustCompile(`INSERT INTO users .* RETURNING .*`)
// 	mock.ExpectQuery(re.String()).
// 		WithArgs(
// 			"jdoe", "jdoe@example.com",
// 			sqlmock.AnyArg(),
// 			"John Doe", "", "John", "Doe",
// 			"Somewhere", "google",
// 			int64(30),
// 			sqlmock.AnyArg(),
// 			sqlmock.AnyArg(), sqlmock.AnyArg(),
// 		).
// 		WillReturnError(sqlmock.ErrCancelled) // dùng 1 error có message; hoặc errors.New("duplicate key …")
//
// 	repo := NewUserRepositoryImpl(UserRepoParams{Logger: testLogger, DB: sqlxDB})
//
// 	_, err := repo.Create(entity.User{
// 		Username:    "jdoe",
// 		Email:       "jdoe@example.com",
// 		Password:    "plain-pass",
// 		FullName:    "John Doe",
// 		FirstName:   "John",
// 		LastName:    "Doe",
// 		Address:     "Somewhere",
// 		InitialFrom: "google",
// 		Age:         30,
// 		Dob:         time.Now().UTC(),
// 	})
// 	require.Error(t, err)
// 	require.NoError(t, mock.ExpectationsWereMet())
// }
//
// // ---------------- CreateByEmailIfNotExisted ----------------
//
// func TestCreateByEmailIfNotExisted_Inserted(t *testing.T) {
// 	sqlxDB, mock, cleanup := newSQLXMock(t)
// 	defer cleanup()
//
// 	reInsert := regexp.MustCompile(`INSERT INTO users .* ON CONFLICT .* DO NOTHING RETURNING .*`)
// 	mock.ExpectQuery(reInsert.String()).
// 		WithArgs("new@example.com").
// 		WillReturnRows(makeUserRow())
//
// 	repo := NewUserRepositoryImpl(UserRepoParams{Logger: testLogger, DB: sqlxDB})
// 	u, err := repo.CreateByEmailIfNotExisted("new@example.com")
// 	require.NoError(t, err)
// 	require.NotNil(t, u)
//
// 	require.NoError(t, mock.ExpectationsWereMet())
// }
//
// func TestCreateByEmailIfNotExisted_AlreadyExists(t *testing.T) {
// 	sqlxDB, mock, cleanup := newSQLXMock(t)
// 	defer cleanup()
//
// 	reInsert := regexp.MustCompile(`INSERT INTO users .* ON CONFLICT .* DO NOTHING RETURNING .*`)
// 	mock.ExpectQuery(reInsert.String()).
// 		WithArgs("exist@example.com").
// 		WillReturnError(sql.ErrNoRows) // không có RETURNING row
//
// 	reSelect := regexp.MustCompile(`SELECT .* FROM users .* WHERE .*email .* OR .*username .* LIMIT 1`)
// 	mock.ExpectQuery(reSelect.String()).
// 		WithArgs("exist@example.com", "exist@example.com").
// 		WillReturnRows(makeUserRow())
//
// 	repo := NewUserRepositoryImpl(UserRepoParams{Logger: testLogger, DB: sqlxDB})
// 	u, err := repo.CreateByEmailIfNotExisted("exist@example.com")
// 	require.NoError(t, err)
// 	require.NotNil(t, u)
//
// 	require.NoError(t, mock.ExpectationsWereMet())
// }
//
// // ---------------- GetByID ----------------
//
// func TestGetByID_Found(t *testing.T) {
// 	sqlxDB, mock, cleanup := newSQLXMock(t)
// 	defer cleanup()
//
// 	re := regexp.MustCompile(`SELECT .* FROM users .* WHERE .*id .* LIMIT 1`)
// 	mock.ExpectQuery(re.String()).
// 		WithArgs("1").
// 		WillReturnRows(makeUserRow())
//
// 	repo := NewUserRepositoryImpl(UserRepoParams{Logger: testLogger, DB: sqlxDB})
// 	u, err := repo.GetByID("1")
// 	require.NoError(t, err)
// 	require.NotNil(t, u)
//
// 	require.NoError(t, mock.ExpectationsWereMet())
// }
//
// func TestGetByID_NotFound(t *testing.T) {
// 	sqlxDB, mock, cleanup := newSQLXMock(t)
// 	defer cleanup()
//
// 	re := regexp.MustCompile(`SELECT .* FROM users .* WHERE .*id .* LIMIT 1`)
// 	mock.ExpectQuery(re.String()).
// 		WithArgs("404").
// 		WillReturnError(sql.ErrNoRows)
//
// 	repo := NewUserRepositoryImpl(UserRepoParams{Logger: testLogger, DB: sqlxDB})
// 	u, err := repo.GetByID("404")
// 	require.Error(t, err)
// 	assert.Nil(t, u)
//
// 	require.NoError(t, mock.ExpectationsWereMet())
// }

// -------------- optional: context compile sanity --------------

func TestContextAvailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	assert.NotNil(t, ctx)
}
