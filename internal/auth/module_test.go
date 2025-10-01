package auth

// import (
// 	"regexp"
// 	"testing"
//
// 	"github.com/DATA-DOG/go-sqlmock"
// 	"github.com/jmoiron/sqlx"
// 	"github.com/spf13/viper"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"google.golang.org/grpc"
//
// 	"github.com/tuannm99/podzone/internal/auth/controller/grpchandler"
// 	"github.com/tuannm99/podzone/pkg/pdlog"
// )
//
// // ---- helpers ----
//
// func newSQLXMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
// 	t.Helper()
// 	raw, mock, err := sqlmock.New()
// 	require.NoError(t, err)
// 	db := sqlx.NewDb(raw, "sqlmock")
// 	cleanup := func() { _ = db.Close() }
// 	return db, mock, cleanup
// }
//
// var nop = &pdlog.NopLogger{}
//
// func TestRegisterGRPCServer_RegistersService(t *testing.T) {
// 	s := grpc.NewServer()
// 	srv := &grpchandler.AuthServer{}
// 	RegisterGRPCServer(s, srv, nop)
//
// 	// Kiểm tra service info có service của Auth
// 	info := s.GetServiceInfo()
// 	// tên service từ protoc thường là "podzone.auth.AuthService" hoặc "...AuthService"
// 	// nhưng ta có thể dựa chắc vào map key do RegisterAuthServiceServer tạo ra.
// 	// Để an toàn, assert rằng service name chứa "AuthService".
// 	found := false
// 	for name := range info {
// 		if regexp.MustCompile(`AuthService$`).MatchString(name) {
// 			found = true
// 			break
// 		}
// 	}
// 	assert.True(t, found, "AuthService should be registered on grpc.Server")
// }
//
// func TestRegisterMigration_ExecutesDDLs(t *testing.T) {
// 	sqlxDB, mock, cleanup := newSQLXMock(t)
// 	defer cleanup()
// 	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS users")).
// 		WillReturnResult(sqlmock.NewResult(0, 0))
// 	mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")).
// 		WillReturnResult(sqlmock.NewResult(0, 0))
// 	mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)")).
// 		WillReturnResult(sqlmock.NewResult(0, 0))
//
// 	params := MigrateParams{
// 		Logger: nop,
// 		DB:     sqlxDB,
// 		V:      viper.New(),
// 	}
//
// 	RegisterMigration(params)
//
// 	require.NoError(t, mock.ExpectationsWereMet())
// }
//
// func TestRegisterMigration_WhenExecFails_LogsAndReturns(t *testing.T) {
// 	sqlxDB, mock, cleanup := newSQLXMock(t)
// 	defer cleanup()
//
// 	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS users")).
// 		WillReturnError(assert.AnError)
//
// 	params := MigrateParams{
// 		Logger: nop,
// 		DB:     sqlxDB,
// 		V:      viper.New(),
// 	}
//
// 	RegisterMigration(params)
// 	require.NoError(t, mock.ExpectationsWereMet())
// }
