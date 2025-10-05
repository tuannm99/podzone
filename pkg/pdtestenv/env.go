// Package pdtestenv spins up infra containers for integration tests.
// Supports Postgres, Redis, MongoDB, OpenSearch (ES-compatible), Kafka.
// Requires: github.com/testcontainers/testcontainers-go v0.39.0
package pdtestenv

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
	"github.com/stretchr/testify/require"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	// Kafka module (v0.39.0+)
	// tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

// Options lets you choose which services to start.
// Reuse: if true, containers are started with fixed names & reuse enabled.
//
//	To enable reuse globally, add `testcontainers.reuse.enable=true` to ~/.testcontainers.properties
type Options struct {
	StartPostgres      bool
	StartRedis         bool
	StartMongo         bool
	StartOpenSearch    bool
	StartElasticsearch bool
	// StartKafka       bool
	Reuse              bool
	Namespace          string // prefix for container names, default: "podzone"
	PostgresImage      string // default: "postgres:16-alpine"
	RedisImage         string // default: "redis:7-alpine"
	MongoImage         string // default: "mongo:6.0"
	OpenSearchImage    string // default: "opensearchproject/opensearch:2.12.0"
	ElasticsearchImage string
	KafkaImage         string // default: "confluentinc/cp-kafka:7.6.1"
	PostgresUser       string // default: "postgres"
	PostgresPassword   string // default: "postgres"
	PostgresDB         string // default: "testdb"
	ElasticsearchURL   string
}

func (o *Options) fillDefaults() {
	if o.Namespace == "" {
		o.Namespace = "podzone"
	}
	if o.PostgresImage == "" {
		o.PostgresImage = "postgres:16-alpine"
	}
	if o.RedisImage == "" {
		o.RedisImage = "redis:7-alpine"
	}
	if o.MongoImage == "" {
		o.MongoImage = "mongo:6.0"
	}
	if o.OpenSearchImage == "" {
		o.OpenSearchImage = "opensearchproject/opensearch:2.12.0"
	}
	if o.ElasticsearchImage == "" {
		o.ElasticsearchImage = "docker.elastic.co/elasticsearch/elasticsearch:8.9.0"
	}

	if o.KafkaImage == "" {
		o.KafkaImage = "confluentinc/cp-kafka:7.6.1"
	}
	if o.PostgresUser == "" {
		o.PostgresUser = "postgres"
	}
	if o.PostgresPassword == "" {
		o.PostgresPassword = "postgres"
	}
	if o.PostgresDB == "" {
		o.PostgresDB = "testdb"
	}
}

// Env holds URIs/DSNs & container handles.
type Env struct {
	PostgresDSN      string
	RedisURI         string
	MongoURI         string
	OpenSearchURL    string
	ElasticsearchURL string
	KafkaBootstrap   string

	// containers
	pgCont    tc.Container
	redisCont tc.Container
	mongoCont tc.Container
	osCont    tc.Container
	esCont    tc.Container
	// kafkaCont *tckafka.KafkaContainer

	// lifecycle
	reuse    bool
	cleanups []func()
}

// Teardown stops containers unless reuse=true.
func (e *Env) Teardown(ctx context.Context) {
	if e == nil {
		return
	}
	// If reuse is enabled, don't auto-terminate to allow reuse across packages.
	if e.reuse {
		return
	}
	for _, c := range e.cleanups {
		c()
	}
}

var globalOnce sync.Once

// Setup spins up requested services and returns Env.
// NOTE: sync.Once is per-process; nếu bạn chạy tests ở nhiều process (nhiều package),
// hãy bật Options.Reuse để tránh sinh quá nhiều container.
func Setup(t *testing.T, opts Options) *Env {
	t.Helper()
	opts.fillDefaults()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	t.Cleanup(cancel)

	env := &Env{reuse: opts.Reuse}
	// Collect cleanups in reverse order
	addCleanup := func(fn func()) { env.cleanups = append([]func(){fn}, env.cleanups...) }

	// We still guard by once to avoid duplicate spins inside 1 process run.
	globalOnce.Do(func() {})

	// Start services as requested
	if opts.StartPostgres {
		env.pgCont = startPostgres(t, ctx, opts, addCleanup)
		host, port := hostPort(t, ctx, env.pgCont, nat.Port("5432/tcp"))
		env.PostgresDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			opts.PostgresUser, opts.PostgresPassword, host, port, opts.PostgresDB)
	}

	if opts.StartRedis {
		env.redisCont = startRedis(t, ctx, opts, addCleanup)
		host, port := hostPort(t, ctx, env.redisCont, nat.Port("6379/tcp"))
		env.RedisURI = fmt.Sprintf("redis://%s:%s/0", host, port)
	}

	if opts.StartMongo {
		env.mongoCont = startMongo(t, ctx, opts, addCleanup)
		host, port := hostPort(t, ctx, env.mongoCont, nat.Port("27017/tcp"))
		env.MongoURI = fmt.Sprintf("mongodb://%s:%s", host, port)
	}

	if opts.StartOpenSearch {
		env.osCont = startOpenSearch(t, ctx, opts, addCleanup)
		host, port := hostPort(t, ctx, env.osCont, nat.Port("9200/tcp"))
		env.OpenSearchURL = fmt.Sprintf("http://%s:%s", host, port)
	}
	if opts.StartElasticsearch { // 👈
		env.esCont = startElasticsearch(t, ctx, opts, addCleanup)
		host, port := hostPort(t, ctx, env.esCont, nat.Port("9200/tcp"))
		env.ElasticsearchURL = fmt.Sprintf("http://%s:%s", host, port)
	}

	// if opts.StartKafka {
	// 	env.kafkaCont = startKafka(t, ctx, opts, addCleanup)
	// 	bootstrap, err := env.kafkaCont.BootstrapServers(ctx)
	// 	require.NoError(t, err)
	// 	env.KafkaBootstrap = bootstrap
	// }

	// Auto-teardown at the end of the test (no-op if reuse=true)
	t.Cleanup(func() { env.Teardown(context.Background()) })

	return env
}

// ---------- Helpers ----------

func hostPort(t *testing.T, ctx context.Context, c tc.Container, p nat.Port) (string, string) {
	t.Helper()
	host, err := c.Host(ctx)
	require.NoError(t, err)
	mp, err := c.MappedPort(ctx, p)
	require.NoError(t, err)
	return host, mp.Port()
}

func withReuse(req *tc.ContainerRequest, reuse bool, name string) {
	if reuse {
		req.Name = name
	}
}

// ---------- Starters ----------

func startPostgres(t *testing.T, ctx context.Context, o Options, addCleanup func(func())) tc.Container {
	env := map[string]string{
		"POSTGRES_USER":     o.PostgresUser,
		"POSTGRES_PASSWORD": o.PostgresPassword,
		"POSTGRES_DB":       o.PostgresDB,
	}
	req := tc.ContainerRequest{
		Image:        o.PostgresImage,
		Env:          env,
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort(nat.Port("5432/tcp")),
		).WithStartupTimeout(2 * time.Minute),
	}
	withReuse(&req, o.Reuse, fmt.Sprintf("%s-postgres", o.Namespace))

	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            o.Reuse,
	})
	require.NoError(t, err)
	addCleanup(func() { _ = c.Terminate(context.Background()) })
	return c
}

func startRedis(t *testing.T, ctx context.Context, o Options, addCleanup func(func())) tc.Container {
	req := tc.ContainerRequest{
		Image:        o.RedisImage,
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to accept connections"),
			wait.ForListeningPort(nat.Port("6379/tcp")),
		).WithStartupTimeout(1 * time.Minute),
	}
	withReuse(&req, o.Reuse, fmt.Sprintf("%s-redis", o.Namespace))

	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            o.Reuse,
	})
	require.NoError(t, err)
	addCleanup(func() { _ = c.Terminate(context.Background()) })
	return c
}

func startMongo(t *testing.T, ctx context.Context, o Options, addCleanup func(func())) tc.Container {
	req := tc.ContainerRequest{
		Image:        o.MongoImage,
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForListeningPort(nat.Port("27017/tcp")).WithStartupTimeout(1 * time.Minute),
	}
	withReuse(&req, o.Reuse, fmt.Sprintf("%s-mongo", o.Namespace))

	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            o.Reuse,
	})
	require.NoError(t, err)
	addCleanup(func() { _ = c.Terminate(context.Background()) })
	return c
}

func startOpenSearch(t *testing.T, ctx context.Context, o Options, addCleanup func(func())) tc.Container {
	req := tc.ContainerRequest{
		Image: o.OpenSearchImage,
		Env: map[string]string{
			"DISABLE_SECURITY_PLUGIN":           "true",       // no auth
			"OPENSEARCH_INITIAL_ADMIN_PASSWORD": "opensearch", // ignored if security disabled
			"discovery.type":                    "single-node",
			"plugins.security.disabled":         "true",
			"OPENSEARCH_JAVA_OPTS":              "-Xms256m -Xmx256m",
		},
		ExposedPorts: []string{"9200/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port("9200/tcp")),
			wait.ForHTTP("/").WithPort("9200/tcp"),
		).WithStartupTimeout(2 * time.Minute),
	}
	withReuse(&req, o.Reuse, fmt.Sprintf("%s-opensearch", o.Namespace))

	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            o.Reuse,
	})
	require.NoError(t, err)
	addCleanup(func() { _ = c.Terminate(context.Background()) })
	return c
}

// ---- Starter cho Elasticsearch 8 (tắt security, single-node) ----
func startElasticsearch(t *testing.T, ctx context.Context, o Options, addCleanup func(func())) tc.Container {
	req := tc.ContainerRequest{
		Image:        o.ElasticsearchImage, // docker.elastic.co/elasticsearch/elasticsearch:8.12.0
		ExposedPorts: []string{"9200/tcp", "9300/tcp"},
		Env: map[string]string{
			"discovery.type":         "single-node",
			"xpack.security.enabled": "false",
			"ES_JAVA_OPTS":           "-Xms256m -Xmx256m",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port("9200/tcp")),
			wait.ForHTTP("/").WithPort("9200/tcp"),
		).WithStartupTimeout(8 * time.Minute),
	}
	withReuse(&req, o.Reuse, fmt.Sprintf("%s-elasticsearch", o.Namespace))

	req.HostConfigModifier = func(h *container.HostConfig) {
		h.Resources.Memory = 0 // no hard limit
		h.Ulimits = append(h.Ulimits, &units.Ulimit{Name: "memlock", Soft: -1, Hard: -1})
	}

	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            o.Reuse,
	})
	require.NoError(t, err)
	addCleanup(func() { _ = c.Terminate(context.Background()) })
	return c
}

// func startKafka(t *testing.T, ctx context.Context, o Options, addCleanup func(func())) *tckafka.KafkaContainer {
// 	// Requires v0.33.0+; v0.39.0 OK
// 	kc, err := tckafka.RunContainer(
// 		ctx,
// 		tckafka.WithImage(o.KafkaImage),
// 		tc.WithReuse(o.Reuse),
// 		tc.WithName(fmt.Sprintf("%s-kafka", o.Namespace)),
// 		// Tuỳ chọn: chờ log cụ thể
// 		tckafka.WithWaitStrategy(
// 			wait.ForLog("Kafka startTimeMs").WithStartupTimeout(3*time.Minute),
// 		),
// 	)
// 	require.NoError(t, err)
// 	addCleanup(func() { _ = kc.Terminate(context.Background()) })
// 	return kc
// }
