package repository

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

type nopLogger = pdlog.NopLogger

func setupRepo(t *testing.T) (*UserRepositoryImpl, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	raw, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	db := sqlx.NewDb(raw, "sqlmock")
	repo := NewUserRepositoryImpl(UserRepoParams{
		Logger: nopLogger{},
		DB:     db,
	})
	return repo, db, mock
}

func userRow(
	now time.Time,
	id int64,
	username, email, password, fullName, initialFrom string,
	age int64,
) *sqlmock.Rows {
	// Avoid NULL for non-nullable Go fields (string/time.Time) in StructScan.
	return sqlmock.NewRows(userColumns).AddRow(
		id,
		username,
		email,
		password,
		fullName,
		"",          // middle_name
		"",          // first_name
		"",          // last_name
		"",          // address
		initialFrom, // initial_from
		age,
		time.Time{}, // dob
		now,
		now,
	)
}

func Test_isUniqueViolation_Table(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			require.Equal(t, tc.want, isUniqueViolation(tc.err))
		})
	}
}

func TestUserRepository_GetByUsernameOrEmail_Table(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC)
	selectRe := `SELECT .* FROM users WHERE .* LIMIT 1`

	cases := []struct {
		name     string
		identity string
		setup    func(mock sqlmock.Sqlmock)
		wantErr  error
		wantID   uint
	}{
		{
			name:     "ok",
			identity: "a@b.com",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectRe).
					WithArgs("a@b.com", "a@b.com").
					WillReturnRows(userRow(now, 1, "jdoe", "a@b.com", "hashed", "John", "podzone", 30))
			},
			wantErr: nil,
			wantID:  1,
		},
		{
			name:     "not_found",
			identity: "missing",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectRe).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: entity.ErrUserNotFound,
		},
		{
			name:     "db_error",
			identity: "x",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectRe).
					WillReturnError(errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, db, mock := setupRepo(t)
			defer db.Close()

			tc.setup(mock)

			u, err := repo.GetByUsernameOrEmail(tc.identity)
			if tc.wantErr == nil {
				require.NoError(t, err)
				require.NotNil(t, u)
				require.Equal(t, tc.wantID, u.Id)
			} else {
				require.Error(t, err)
				// For sentinel errors, use ErrorIs; for raw errors, check message.
				if errors.Is(tc.wantErr, entity.ErrUserNotFound) {
					require.ErrorIs(t, err, tc.wantErr)
				} else {
					require.Contains(t, err.Error(), tc.wantErr.Error())
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByID_Table(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC)
	selectRe := `SELECT .* FROM users WHERE id = \$1 LIMIT 1`

	cases := []struct {
		name    string
		id      string
		setup   func(mock sqlmock.Sqlmock)
		wantErr error
		wantID  uint
	}{
		{
			name: "ok",
			id:   "1",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectRe).
					WithArgs("1").
					WillReturnRows(userRow(now, 1, "jdoe", "jdoe@example.com", "hashed", "John", "podzone", 30))
			},
			wantID: 1,
		},
		{
			name: "not_found",
			id:   "999",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectRe).
					WithArgs("999").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: entity.ErrUserNotFound,
		},
		{
			name: "db_error",
			id:   "1",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectRe).
					WithArgs("1").
					WillReturnError(errors.New("select failed"))
			},
			wantErr: errors.New("select failed"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, db, mock := setupRepo(t)
			defer db.Close()

			tc.setup(mock)

			u, err := repo.GetByID(tc.id)
			if tc.wantErr == nil {
				require.NoError(t, err)
				require.NotNil(t, u)
				require.Equal(t, tc.wantID, u.Id)
			} else {
				require.Error(t, err)
				if errors.Is(tc.wantErr, entity.ErrUserNotFound) {
					require.ErrorIs(t, err, tc.wantErr)
				} else {
					require.Contains(t, err.Error(), tc.wantErr.Error())
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Create_Table(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC)
	insertRe := regexp.QuoteMeta("INSERT INTO users")

	cases := []struct {
		name    string
		in      entity.User
		setup   func(mock sqlmock.Sqlmock)
		wantErr error
		wantID  uint
	}{
		{
			name: "ok",
			in: entity.User{
				Username: "jdoe",
				Email:    "jdoe@example.com",
				Password: "secret",
				FullName: "John Doe",
				Age:      30,
			},
			setup: func(mock sqlmock.Sqlmock) {
				// Keep expectations loose; hashing and timestamps are not deterministic.
				mock.ExpectQuery(insertRe).
					WillReturnRows(userRow(now, 10, "jdoe", "jdoe@example.com", "hashed-secret", "John Doe", "", 30))
			},
			wantID: 10,
		},
		{
			name: "unique_violation",
			in: entity.User{
				Username: "another",
				Email:    "dup@example.com",
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(insertRe).
					WillReturnError(errors.New(`duplicate key value violates unique constraint "users_email_key"`))
			},
			wantErr: entity.ErrUserAlreadyExists,
		},
		{
			name: "db_error",
			in: entity.User{
				Username: "x",
				Email:    "x@example.com",
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(insertRe).
					WillReturnError(errors.New("insert failed"))
			},
			wantErr: errors.New("insert failed"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, db, mock := setupRepo(t)
			defer db.Close()

			tc.setup(mock)

			out, err := repo.Create(tc.in)
			if tc.wantErr == nil {
				require.NoError(t, err)
				require.NotNil(t, out)
				require.Equal(t, tc.wantID, out.Id)
			} else {
				require.Error(t, err)
				if errors.Is(tc.wantErr, entity.ErrUserAlreadyExists) {
					require.ErrorIs(t, err, tc.wantErr)
				} else {
					require.Contains(t, err.Error(), tc.wantErr.Error())
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Update_EarlyReturn_Table(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		in      entity.User
		wantErr error
	}{
		{
			name:    "id_zero",
			in:      entity.User{Id: 0},
			wantErr: entity.ErrUserNotFound,
		},
		{
			name:    "no_fields",
			in:      entity.User{Id: 1}, // nothing set => only updated_at => return nil without hitting DB
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, db, mock := setupRepo(t)
			defer db.Close()

			err := repo.Update(tc.in)
			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			// No SQL expectations should be pending in early-return paths.
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Update_Exec_Table(t *testing.T) {
	t.Parallel()

	updateRe := `UPDATE users SET .* WHERE id = \$[0-9]+`

	cases := []struct {
		name    string
		in      entity.User
		setup   func(mock sqlmock.Sqlmock)
		wantErr error
	}{
		{
			name: "ok_rows_1",
			in: entity.User{
				Id:          1,
				FullName:    "John X",
				Password:    "changed",
				Age:         31,
				InitialFrom: "podzone",
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(updateRe).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: nil,
		},
		{
			name: "rows_0_not_found",
			in:   entity.User{Id: 1, Email: "x@example.com"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(updateRe).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: entity.ErrUserNotFound,
		},
		{
			name: "unique_violation",
			in:   entity.User{Id: 1, Email: "dup@example.com"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(updateRe).
					WillReturnError(errors.New("duplicate key value violates unique constraint"))
			},
			wantErr: entity.ErrUserAlreadyExists,
		},
		{
			name: "db_error",
			in:   entity.User{Id: 1, Email: "x@example.com"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(updateRe).
					WillReturnError(errors.New("update failed"))
			},
			wantErr: errors.New("update failed"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, db, mock := setupRepo(t)
			defer db.Close()

			tc.setup(mock)

			err := repo.Update(tc.in)
			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if errors.Is(tc.wantErr, entity.ErrUserNotFound) || errors.Is(tc.wantErr, entity.ErrUserAlreadyExists) {
					require.ErrorIs(t, err, tc.wantErr)
				} else {
					require.Contains(t, err.Error(), tc.wantErr.Error())
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_UpdateById_Table(t *testing.T) {
	t.Parallel()

	updateRe := `UPDATE users SET .* WHERE id = \$[0-9]+`

	cases := []struct {
		name    string
		id      uint
		in      entity.User
		setup   func(mock sqlmock.Sqlmock)
		wantErr error
	}{
		{
			name: "ok",
			id:   1,
			in:   entity.User{FullName: "A"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(updateRe).WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, db, mock := setupRepo(t)
			defer db.Close()

			tc.setup(mock)

			err := repo.UpdateById(tc.id, tc.in)
			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

