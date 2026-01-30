package testkit

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	postgresOnce    sync.Once
	postgresCleanup sync.Once
	postgresDSN     string
	errPostgres     error
	postgresC       testcontainers.Container

	redisOnce    sync.Once
	redisCleanup sync.Once
	redisAddr    string
	errRedis     error
	redisC       testcontainers.Container

	mongoOnce    sync.Once
	mongoCleanup sync.Once
	mongoURI     string
	errMongo     error
	mongoC       testcontainers.Container

	esOnce    sync.Once
	esCleanup sync.Once
	esURL     string
	errES     error
	esC       testcontainers.Container
)

func reuseEnabled() bool {
	if v := os.Getenv("PODZONE_TC_REUSE"); v != "" {
		return v == "1" || strings.EqualFold(v, "true")
	}
	return strings.EqualFold(os.Getenv("TESTCONTAINERS_REUSE_ENABLE"), "true")
}

func registerCleanup(t *testing.T, c testcontainers.Container, once *sync.Once) {
	t.Helper()
	if c == nil || reuseEnabled() {
		return
	}
	once.Do(func() {
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = c.Terminate(ctx)
		})
	})
}

func isDockerUnavailable(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "cannot connect to the docker daemon") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "is the docker daemon running")
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
	registerCleanup(t, postgresC, &postgresCleanup)
	return postgresDSN
}

type PostgresConnInfo struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
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
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
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
	if err := admin.QueryRow(`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, dbName).Scan(&exists); err != nil {
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
	postgresOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		req := testcontainers.ContainerRequest{
			Image:        "postgres:15",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "podzone",
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
		postgresDSN = fmt.Sprintf("postgres://postgres:postgres@%s:%s/podzone?sslmode=disable", host, port.Port())
	})
}

func RedisAddr(t *testing.T) string {
	t.Helper()
	startRedis(t)
	if errRedis != nil {
		if isDockerUnavailable(errRedis) {
			t.Skipf("docker unavailable: %v", errRedis)
		}
		t.Fatalf("start redis container: %v", errRedis)
	}
	registerCleanup(t, redisC, &redisCleanup)
	return redisAddr
}

func RedisURI(t *testing.T, db int) string {
	t.Helper()
	return fmt.Sprintf("redis://%s/%d", RedisAddr(t), db)
}

func RedisClient(t *testing.T) *redis.Client {
	t.Helper()
	addr := RedisAddr(t)
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func startRedis(t *testing.T) {
	t.Helper()
	redisOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		req := testcontainers.ContainerRequest{
			Image:        "redis:7",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(2 * time.Minute),
		}
		reuse := reuseEnabled()
		if reuse {
			req.Name = "podzone-it-redis"
		}

		redisC, errRedis = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
			Reuse:            reuse,
		})
		if errRedis != nil {
			return
		}

		host, err := redisC.Host(ctx)
		if err != nil {
			errRedis = err
			return
		}
		port, err := redisC.MappedPort(ctx, "6379/tcp")
		if err != nil {
			errRedis = err
			return
		}
		redisAddr = fmt.Sprintf("%s:%s", host, port.Port())
	})
}

func MongoURI(t *testing.T) string {
	t.Helper()
	startMongo(t)
	if errMongo != nil {
		if isDockerUnavailable(errMongo) {
			t.Skipf("docker unavailable: %v", errMongo)
		}
		t.Fatalf("start mongo container: %v", errMongo)
	}
	registerCleanup(t, mongoC, &mongoCleanup)
	return mongoURI
}

func MongoClient(t *testing.T) *mongo.Client {
	t.Helper()
	uri := MongoURI(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("ping mongo: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	return client
}

func startMongo(t *testing.T) {
	t.Helper()
	mongoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		req := testcontainers.ContainerRequest{
			Image:        "mongo:6",
			ExposedPorts: []string{"27017/tcp"},
			WaitingFor:   wait.ForListeningPort("27017/tcp").WithStartupTimeout(2 * time.Minute),
		}
		reuse := reuseEnabled()
		if reuse {
			req.Name = "podzone-it-mongo"
		}

		mongoC, errMongo = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
			Reuse:            reuse,
		})
		if errMongo != nil {
			return
		}

		host, err := mongoC.Host(ctx)
		if err != nil {
			errMongo = err
			return
		}
		port, err := mongoC.MappedPort(ctx, "27017/tcp")
		if err != nil {
			errMongo = err
			return
		}
		mongoURI = fmt.Sprintf("mongodb://%s:%s", host, port.Port())
	})
}

func ElasticsearchURL(t *testing.T) string {
	t.Helper()
	startElasticsearch(t)
	if errES != nil {
		if isDockerUnavailable(errES) {
			t.Skipf("docker unavailable: %v", errES)
		}
		t.Fatalf("start elasticsearch container: %v", errES)
	}
	registerCleanup(t, esC, &esCleanup)
	return esURL
}

func startElasticsearch(t *testing.T) {
	t.Helper()
	esOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		req := testcontainers.ContainerRequest{
			Image:        "docker.elastic.co/elasticsearch/elasticsearch:7.17.10",
			ExposedPorts: []string{"9200/tcp"},
			Env: map[string]string{
				"discovery.type":         "single-node",
				"xpack.security.enabled": "false",
				"ES_JAVA_OPTS":           "-Xms512m -Xmx512m",
			},
			WaitingFor: wait.ForListeningPort("9200/tcp").WithStartupTimeout(3 * time.Minute),
		}
		reuse := reuseEnabled()
		if reuse {
			req.Name = "podzone-it-es"
		}

		esC, errES = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
			Reuse:            reuse,
		})
		if errES != nil {
			return
		}

		host, err := esC.Host(ctx)
		if err != nil {
			errES = err
			return
		}
		port, err := esC.MappedPort(ctx, "9200/tcp")
		if err != nil {
			errES = err
			return
		}
		esURL = fmt.Sprintf("http://%s:%s", host, port.Port())
	})
}
