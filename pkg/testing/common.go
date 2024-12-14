package testing

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tuannm99/podzone/pkg/config"
	"github.com/tuannm99/podzone/pkg/database"
	"github.com/tuannm99/podzone/pkg/logging"
	"github.com/tuannm99/podzone/pkg/messaging"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// PostgresContainer represents a PostgreSQL container for testing
type PostgresContainer struct {
	Container     *dockertest.Resource
	ConnectionURL string
	DB            *sql.DB
	Pool          *docker.Pool
	Config        config.DatabaseConfig
}

// NewPostgresContainer creates a new PostgreSQL container for testing
func NewPostgresContainer(t *testing.T) (*PostgresContainer, error) {
	t.Helper()

	// Create a new docker pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	// Set a timeout for container startup
	pool.MaxWait = 2 * time.Minute

	// Create a PostgreSQL container
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=testdb",
			"listen_addresses='*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return nil, fmt.Errorf("could not start postgres container: %w", err)
	}

	// Set a cleanup function on test completion
	t.Cleanup(func() {
		// Kill the container
		if err := pool.Purge(resource); err != nil {
			t.Logf("Could not purge resource: %s", err)
		}
	})

	// Get host and port
	hostAndPort := resource.GetHostPort("5432/tcp")
	host, port, err := net.SplitHostPort(hostAndPort)
	if err != nil {
		return nil, fmt.Errorf("could not parse host:port: %w", err)
	}

	// Create connection URL
	connURL := fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", host, port)

	// Wait for the container to be ready
	if err := pool.Retry(func() error {
		db, err := sql.Open("postgres", connURL)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %w", err)
	}

	// Create a database connection
	db, err := sql.Open("postgres", connURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %w", err)
	}

	// Create config
	dbConfig := config.DatabaseConfig{
		Type:     "postgres",
		Host:     host,
		Port:     parsePort(port),
		Username: "postgres",
		Password: "postgres",
		Database: "testdb",
		SSLMode:  "disable",
		MaxConns: 10,
		MinConns: 2,
	}

	// Return the container
	return &PostgresContainer{
		Container:     resource,
		ConnectionURL: connURL,
		DB:            db,
		Pool:          pool,
		Config:        dbConfig,
	}, nil
}

// NewPgxPool creates a new pgx pool for the PostgreSQL container
func (c *PostgresContainer) NewPgxPool(t *testing.T) (*pgxpool.Pool, error) {
	t.Helper()

	// Create a new pool
	pool, err := pgxpool.New(context.Background(), c.ConnectionURL)
	if err != nil {
		return nil, fmt.Errorf("could not create pgx pool: %w", err)
	}

	// Set a cleanup function
	t.Cleanup(func() {
		pool.Close()
	})

	return pool, nil
}

// Close closes the PostgreSQL container
func (c *PostgresContainer) Close() error {
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			return err
		}
	}
	return c.Pool.Purge(c.Container)
}

// MongoDBContainer represents a MongoDB container for testing
type MongoDBContainer struct {
	Container     *dockertest.Resource
	ConnectionURL string
	Client        *mongo.Client
	Pool          *docker.Pool
	Config        config.DatabaseConfig
}

// NewMongoDBContainer creates a new MongoDB container for testing
func NewMongoDBContainer(t *testing.T) (*MongoDBContainer, error) {
	t.Helper()

	// Create a new docker pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	// Set a timeout for container startup
	pool.MaxWait = 2 * time.Minute

	// Create a MongoDB container
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "6",
		Env: []string{
			"MONGO_INITDB_ROOT_USERNAME=mongo",
			"MONGO_INITDB_ROOT_PASSWORD=mongo",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return nil, fmt.Errorf("could not start mongodb container: %w", err)
	}

	// Set a cleanup function on test completion
	t.Cleanup(func() {
		// Kill the container
		if err := pool.Purge(resource); err != nil {
			t.Logf("Could not purge resource: %s", err)
		}
	})

	// Get host and port
	hostAndPort := resource.GetHostPort("27017/tcp")
	host, port, err := net.SplitHostPort(hostAndPort)
	if err != nil {
		return nil, fmt.Errorf("could not parse host:port: %w", err)
	}

	// Create connection URL
	connURL := fmt.Sprintf("mongodb://mongo:mongo@%s:%s", host, port)

	// Wait for the container to be ready
	if err := pool.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(connURL))
		if err != nil {
			return err
		}
		defer client.Disconnect(ctx)

		return client.Ping(ctx, nil)
	}); err != nil {
		return nil, fmt.Errorf("could not connect to mongodb: %w", err)
	}

	// Create a database connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connURL))
	if err != nil {
		return nil, fmt.Errorf("could not connect to mongodb: %w", err)
	}

	// Create config
	dbConfig := config.DatabaseConfig{
		Type:     "mongodb",
		Host:     host,
		Port:     parsePort(port),
		Username: "mongo",
		Password: "mongo",
		Database: "testdb",
		MaxConns: 10,
		MinConns: 2,
	}

	// Return the container
	return &MongoDBContainer{
		Container:     resource,
		ConnectionURL: connURL,
		Client:        client,
		Pool:          pool,
		Config:        dbConfig,
	}, nil
}

// Close closes the MongoDB container
func (c *MongoDBContainer) Close() error {
	if c.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := c.Client.Disconnect(ctx); err != nil {
			return err
		}
	}
	return c.Pool.Purge(c.Container)
}

// RedisContainer represents a Redis container for testing
type RedisContainer struct {
	Container     *dockertest.Resource
	ConnectionURL string
	Client        *redis.Client
	Pool          *docker.Pool
	Config        config.RedisConfig
}

// NewRedisContainer creates a new Redis container for testing
func NewRedisContainer(t *testing.T) (*RedisContainer, error) {
	t.Helper()

	// Create a new docker pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	// Set a timeout for container startup
	pool.MaxWait = 2 * time.Minute

	// Create a Redis container
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "7",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return nil, fmt.Errorf("could not start redis container: %w", err)
	}

	// Set a cleanup function on test completion
	t.Cleanup(func() {
		// Kill the container
		if err := pool.Purge(resource); err != nil {
			t.Logf("Could not purge resource: %s", err)
		}
	})

	// Get host and port
	hostAndPort := resource.GetHostPort("6379/tcp")
	host, port, err := net.SplitHostPort(hostAndPort)
	if err != nil {
		return nil, fmt.Errorf("could not parse host:port: %w", err)
	}

	// Create connection URL
	connURL := fmt.Sprintf("redis://%s:%s", host, port)

	// Wait for the container to be ready
	if err := pool.Retry(func() error {
		client := redis.NewClient(&redis.Options{
			Addr: hostAndPort,
		})
		return client.Ping(context.Background()).Err()
	}); err != nil {
		return nil, fmt.Errorf("could not connect to redis: %w", err)
	}

	// Create a Redis client
	client := redis.NewClient(&redis.Options{
		Addr: hostAndPort,
	})

	// Create config
	redisConfig := config.RedisConfig{
		Host:     host,
		Port:     parsePort(port),
		Password: "",
		DB:       0,
	}

	// Return the container
	return &RedisContainer{
		Container:     resource,
		ConnectionURL: connURL,
		Client:        client,
		Pool:          pool,
		Config:        redisConfig,
	}, nil
}

// Close closes the Redis container
func (c *RedisContainer) Close() error {
	if c.Client != nil {
		if err := c.Client.Close(); err != nil {
			return err
		}
	}
	return c.Pool.Purge(c.Container)
}

// KafkaContainer represents a Kafka container for testing
type KafkaContainer struct {
	Container     *dockertest.Resource
	ConnectionURL string
	Pool          *docker.Pool
}

// NewKafkaContainer creates a new Kafka container for testing
func NewKafkaContainer(t *testing.T) (*KafkaContainer, error) {
	t.Helper()

	// Create a new docker pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	// Set a timeout for container startup
	pool.MaxWait = 2 * time.Minute

	// Create a ZooKeeper container
	zkResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "confluentinc/cp-zookeeper",
		Tag:        "7.3.2",
		Env: []string{
			"ZOOKEEPER_CLIENT_PORT=2181",
			"ZOOKEEPER_TICK_TIME=2000",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return nil, fmt.Errorf("could not start zookeeper container: %w", err)
	}

	// Set a cleanup function on test completion
	t.Cleanup(func() {
		// Kill the container
		if err := pool.Purge(zkResource); err != nil {
			t.Logf("Could not purge zookeeper: %s", err)
		}
	})

	// Get ZooKeeper host and port
	zkHostAndPort := zkResource.GetHostPort("2181/tcp")

	// Create a Kafka container
	kafkaResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "confluentinc/cp-kafka",
		Tag:        "7.3.2",
		Env: []string{
			fmt.Sprintf("KAFKA_ZOOKEEPER_CONNECT=%s", zkHostAndPort),
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT",
			"KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://kafka:9092,PLAINTEXT_HOST://localhost:9093",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE=true",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
		config.Links = []string{zkResource.Container.Name + ":zookeeper"}
	})
	if err != nil {
		return nil, fmt.Errorf("could not start kafka container: %w", err)
	}

	// Set a cleanup function on test completion
	t.Cleanup(func() {
		// Kill the container
		if err := pool.Purge(kafkaResource); err != nil {
			t.Logf("Could not purge kafka: %s", err)
		}
	})

	// Get Kafka host and port
	kafkaHostAndPort := kafkaResource.GetHostPort("9093/tcp")

	// Wait for the container to be ready
	if err := pool.Retry(func() error {
		// In a real test, you would check if Kafka is ready
		// For simplicity, we'll just wait
		time.Sleep(5 * time.Second)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not connect to kafka: %w", err)
	}

	// Return the container
	return &KafkaContainer{
		Container:     kafkaResource,
		ConnectionURL: kafkaHostAndPort,
		Pool:          pool,
	}, nil
}

// Close closes the Kafka container
func (c *KafkaContainer) Close() error {
	return c.Pool.Purge(c.Container)
}

// NewMessagingClient creates a new messaging client for testing
func NewMessagingClient(t *testing.T, kafkaContainer *KafkaContainer) (messaging.Client, error) {
	t.Helper()

	// Create logger
	logger, err := logging.NewLogger("debug", "test")
	if err != nil {
		return nil, err
	}

	// Create messaging config
	config := messaging.Config{
		Brokers:          []string{kafkaContainer.ConnectionURL},
		ClientID:         "test-client",
		GroupID:          "test-group",
		PublishTimeout:   5 * time.Second,
		SubscribeTimeout: 5 * time.Second,
		MaxRetries:       3,
		RetryInterval:    1 * time.Second,
	}

	// Create messaging client
	client, err := messaging.NewKafkaClient(config, logger)
	if err != nil {
		return nil, err
	}

	// Set a cleanup function
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Logf("Could not close messaging client: %s", err)
		}
	})

	return client, nil
}

// BufferedGRPC provides a gRPC server and client connection for testing
type BufferedGRPC struct {
	Server     *grpc.Server
	ClientConn *grpc.ClientConn
	Listener   *bufconn.Listener
}

// NewBufferedGRPC creates a new buffered gRPC setup for testing
func NewBufferedGRPC(t *testing.T, bufSize int) *BufferedGRPC {
	t.Helper()

	if bufSize <= 0 {
		bufSize = 1024 * 1024 // 1MB default
	}

	// Create a buffered listener
	listener := bufconn.Listen(bufSize)

	// Create a gRPC server
	server := grpc.NewServer()

	// Set up cleanup
	t.Cleanup(func() {
		server.Stop()
		listener.Close()
	})

	// Start the server
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("Failed to serve: %v", err)
		}
	}()

	// Create a client connection
	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientConn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	t.Cleanup(func() {
		clientConn.Close()
	})

	return &BufferedGRPC{
		Server:     server,
		ClientConn: clientConn,
		Listener:   listener,
	}
}

// ExecuteSQL executes a SQL file against a database
func ExecuteSQL(t *testing.T, db *sql.DB, sqlPath string) error {
	t.Helper()

	// Read SQL file
	sqlBytes, err := ioutil.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("could not read SQL file: %w", err)
	}

	// Execute SQL
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("could not execute SQL: %w", err)
	}

	return nil
}

// TempFile creates a temporary file with content
func TempFile(t *testing.T, content []byte) (string, error) {
	t.Helper()

	// Create temporary file
	tmpFile, err := ioutil.TempFile("", "test-*.tmp")
	if err != nil {
		return "", fmt.Errorf("could not create temporary file: %w", err)
	}

	// Write content to file
	if _, err := tmpFile.Write(content); err != nil {
		return "", fmt.Errorf("could not write to temporary file: %w", err)
	}

	// Close file
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("could not close temporary file: %w", err)
	}

	// Set cleanup function
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return tmpFile.Name(), nil
}

// TempDir creates a temporary directory
func TempDir(t *testing.T) (string, error) {
	t.Helper()

	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "test-*")
	if err != nil {
		return "", fmt.Errorf("could not create temporary directory: %w", err)
	}

	// Set cleanup function
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return tmpDir, nil
}

// LoadFixture loads a test fixture from a file
func LoadFixture(t *testing.T, path string) ([]byte, error) {
	t.Helper()

	// Read fixture file
	fixture, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read fixture file: %w", err)
	}

	return fixture, nil
}

// MockServer provides a simple HTTP server for testing
type MockServer struct {
	Server *httptest.Server
	URL    string
	Mux    *http.ServeMux
}

// NewMockServer creates a new HTTP server for testing
func NewMockServer(t *testing.T) *MockServer {
	t.Helper()

	// Create a new server mux
	mux := http.NewServeMux()

	// Create a new server
	server := httptest.NewServer(mux)

	// Set cleanup function
	t.Cleanup(func() {
		server.Close()
	})

	return &MockServer{
		Server: server,
		URL:    server.URL,
		Mux:    mux,
	}
}

// Helper functions

// parsePort parses a port string to an integer
func parsePort(port string) int {
	p, _ := strconv.Atoi(port)
	return p
}
