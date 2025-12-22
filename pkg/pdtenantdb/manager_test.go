package pdtenantdb_test

import (
	"context"
	"database/sql"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pdtenantdb "github.com/tuannm99/podzone/pkg/pdtenantdb"
	pdmocks "github.com/tuannm99/podzone/pkg/pdtenantdb/mocks"
)

func TestManager_DBForTenant_ReusesPoolByClusterAndDB(t *testing.T) {
	cfg := &pdtenantdb.Config{
		SharedDB:          "backoffice",
		ConnMaxLifetime:   1 * time.Minute,
		MaxOpenConns:      10,
		MaxIdleConns:      2,
		MaxDedicatedPools: 200,
		DedicatedIdleTTL:  30 * time.Minute,
	}

	resolver := pdmocks.NewMockPlacementResolver(t)
	registry := pdmocks.NewMockClusterRegistry(t)

	resolver.EXPECT().
		Resolve(mock.Anything, "t1").
		Return(pdtenantdb.Placement{
			TenantID:    "t1",
			ClusterName: "pg-01",
			Mode:        pdtenantdb.ModeSchema,
			DBName:      "backoffice",
			SchemaName:  "t_t1",
		}, nil).
		Twice()

	registry.EXPECT().
		GetCluster(mock.Anything, "pg-01").
		Return(pdtenantdb.ClusterConfig{
			Host: "pgbouncer.svc", Port: 6432, User: "u", Password: "p", SSLMode: "disable",
		}, nil).
		Once()

	oldOpen := pdtenantdb.SQLXOpen
	defer func() { pdtenantdb.SQLXOpen = oldOpen }()

	var openCount int32
	pdtenantdb.SQLXOpen = func(driverName, dsn string) (*sqlx.DB, error) {
		atomic.AddInt32(&openCount, 1)

		db, sm, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			return nil, err
		}
		sm.ExpectPing().WillReturnError(nil)

		return sqlx.NewDb(db, "sqlmock"), nil
	}

	m := pdtenantdb.NewManager(cfg, resolver, registry)

	db1, pl1, err := m.DBForTenant(context.Background(), "t1")
	require.NoError(t, err)
	require.NotNil(t, db1)
	require.Equal(t, "backoffice", pl1.DBName)

	db2, pl2, err := m.DBForTenant(context.Background(), "t1")
	require.NoError(t, err)
	require.Same(t, db1, db2)
	require.Equal(t, pl1.DBName, pl2.DBName)

	require.Equal(t, int32(1), atomic.LoadInt32(&openCount))
}

func TestManager_WithTenantTx_SetsSearchPathForSchemaMode(t *testing.T) {
	cfg := &pdtenantdb.Config{SharedDB: "backoffice"}

	resolver := pdmocks.NewMockPlacementResolver(t)
	registry := pdmocks.NewMockClusterRegistry(t)

	resolver.EXPECT().
		Resolve(mock.Anything, "t1").
		Return(pdtenantdb.Placement{
			TenantID: "t1", ClusterName: "pg-01",
			Mode: pdtenantdb.ModeSchema, DBName: "backoffice", SchemaName: "t_t1",
		}, nil).
		Once()

	registry.EXPECT().
		GetCluster(mock.Anything, "pg-01").
		Return(pdtenantdb.ClusterConfig{
			Host: "pgbouncer.svc", Port: 6432, User: "u", Password: "p", SSLMode: "disable",
		}, nil).
		Once()

	oldOpen := pdtenantdb.SQLXOpen
	defer func() { pdtenantdb.SQLXOpen = oldOpen }()

	var sm sqlmock.Sqlmock
	pdtenantdb.SQLXOpen = func(driverName, dsn string) (*sqlx.DB, error) {
		db, mockDB, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			return nil, err
		}
		sm = mockDB

		sm.ExpectPing().WillReturnError(nil)
		sm.ExpectBegin()
		sm.ExpectExec(`SET LOCAL search_path TO "t_t1", public`).
			WillReturnResult(sqlmock.NewResult(0, 0))
		sm.ExpectExec(`SELECT 1`).
			WillReturnResult(sqlmock.NewResult(0, 0))
		sm.ExpectCommit()

		return sqlx.NewDb(db, "sqlmock"), nil
	}

	m := pdtenantdb.NewManager(cfg, resolver, registry)

	err := m.WithTenantTx(context.Background(), "t1", &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(context.Background(), "SELECT 1")
		return err
	})
	require.NoError(t, err)
	require.NoError(t, sm.ExpectationsWereMet())
}

func TestManager_DedicatedPoolCapacity(t *testing.T) {
	cfg := &pdtenantdb.Config{
		SharedDB:          "backoffice",
		MaxDedicatedPools: 1,
	}

	resolver := pdmocks.NewMockPlacementResolver(t)
	registry := pdmocks.NewMockClusterRegistry(t)

	resolver.EXPECT().
		Resolve(mock.Anything, "a").
		Return(pdtenantdb.Placement{TenantID: "a", ClusterName: "pg-01", Mode: pdtenantdb.ModeDatabase, DBName: "bo_a"}, nil).
		Once()

	resolver.EXPECT().
		Resolve(mock.Anything, "b").
		Return(pdtenantdb.Placement{TenantID: "b", ClusterName: "pg-01", Mode: pdtenantdb.ModeDatabase, DBName: "bo_b"}, nil).
		Once()

	registry.EXPECT().
		GetCluster(mock.Anything, "pg-01").
		Return(pdtenantdb.ClusterConfig{
			Host: "pgbouncer.svc", Port: 6432, User: "u", Password: "p", SSLMode: "disable",
		}, nil).
		Once()

	oldOpen := pdtenantdb.SQLXOpen
	defer func() { pdtenantdb.SQLXOpen = oldOpen }()

	pdtenantdb.SQLXOpen = func(driverName, dsn string) (*sqlx.DB, error) {
		db, sm, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			return nil, err
		}
		sm.ExpectPing().WillReturnError(nil)
		return sqlx.NewDb(db, "sqlmock"), nil
	}

	m := pdtenantdb.NewManager(cfg, resolver, registry)

	_, _, err := m.DBForTenant(context.Background(), "a")
	require.NoError(t, err)

	_, _, err = m.DBForTenant(context.Background(), "b")
	require.Error(t, err)
	require.Contains(t, err.Error(), "dedicated pool capacity")
}
