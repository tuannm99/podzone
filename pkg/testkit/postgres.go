package testkit

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresConnInfo struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func PostgresDSN(t *testing.T) string {
	t.Helper()
	startPostgres(t)
	if errPostgres != nil {
		if isDockerUnavailable(errPostgres) {
			t.Skipf("docker unavailable: %v", errPostgres)
		}
		t.Fatalf("start postgres container: %v", errPostgres)
	}
	registerCleanup(t, postgresC)
	return postgresDSN
}

func PostgresDSNWithDB(t *testing.T, dbName string) string {
	t.Helper()
	u, err := url.Parse(PostgresDSN(t))
	if err != nil {
		t.Fatalf("parse postgres dsn: %v", err)
	}
	u.Path = path.Join("/", dbName)
	return u.String()
}

func PostgresInfo(t *testing.T) PostgresConnInfo {
	t.Helper()
	u, err := url.Parse(PostgresDSN(t))
	if err != nil {
		t.Fatalf("parse postgres dsn: %v", err)
	}
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("parse postgres host: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse postgres port: %v", err)
	}
	user := ""
	pass := ""
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	sslMode := u.Query().Get("sslmode")
	return PostgresConnInfo{
		Host:     host,
		Port:     port,
		User:     user,
		Password: pass,
		DBName:   dbName,
		SSLMode:  sslMode,
	}
}

func PostgresDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := PostgresDSN(t)
	var (
		db  *sqlx.DB
		err error
	)
	for attempt := 0; attempt < 8; attempt++ {
		db, err = sqlx.Open("postgres", dsn)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = db.PingContext(ctx)
			cancel()
			if err == nil {
				t.Cleanup(func() { _ = db.Close() })
				return db
			}
			_ = db.Close()
		}
		time.Sleep(time.Duration(attempt+1) * 250 * time.Millisecond)
	}
	t.Fatalf("connect postgres: %v", err)
	return nil
}

func EnsurePostgresDB(t *testing.T, dbName string) {
	t.Helper()
	adminDSN := PostgresDSNWithDB(t, "postgres")
	admin, err := sqlx.Connect("postgres", adminDSN)
	if err != nil {
		t.Fatalf("connect postgres admin: %v", err)
	}
	defer admin.Close()

	var exists bool
	if err := admin.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`,
		dbName,
	).Scan(&exists); err != nil {
		t.Fatalf("check db exists: %v", err)
	}
	if exists {
		return
	}
	if _, err := admin.Exec(`CREATE DATABASE "` + dbName + `"`); err != nil {
		t.Fatalf("create db %s: %v", dbName, err)
	}
}

func startPostgres(t *testing.T) {
	t.Helper()
	postgresMu.Lock()
	defer postgresMu.Unlock()
	defer captureDockerPanic(&errPostgres)

	if isContainerRunning(postgresC) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(2 * time.Minute),
	}
	reuse := reuseEnabled()
	if reuse {
		req.Name = "podzone-it-postgres"
	}

	postgresC, errPostgres = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            reuse,
	})
	if errPostgres != nil {
		return
	}

	host, err := postgresC.Host(ctx)
	if err != nil {
		errPostgres = err
		return
	}
	port, err := postgresC.MappedPort(ctx, "5432/tcp")
	if err != nil {
		errPostgres = err
		return
	}
	postgresDSN = fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres?sslmode=disable", host, port.Port())
}
