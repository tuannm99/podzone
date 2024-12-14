package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tuannm99/podzone/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

// MongoDB represents a MongoDB database connection
type MongoDB struct {
	Client *mongo.Client
	DB     *mongo.Database
	logger *zap.Logger
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(ctx context.Context, cfg *config.DatabaseConfig, logger *zap.Logger) (*PostgresDB, error) {
	connString := cfg.ConnectionString()
	if connString == "" {
		return nil, fmt.Errorf("invalid database connection string")
	}

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Configure the connection pool
	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MinConns = int32(cfg.MinConns)
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Check connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL database",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database))

	return &PostgresDB{
		Pool:   pool,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.logger.Info("PostgreSQL connection closed")
	}
}

// NewMongoDB creates a new MongoDB database connection
func NewMongoDB(ctx context.Context, cfg *config.DatabaseConfig, logger *zap.Logger) (*MongoDB, error) {
	connString := cfg.ConnectionString()
	if connString == "" {
		return nil, fmt.Errorf("invalid database connection string")
	}

	// Create client options
	clientOptions := options.Client().ApplyURI(connString)
	clientOptions.SetMaxPoolSize(uint64(cfg.MaxConns))
	clientOptions.SetMinPoolSize(uint64(cfg.MinConns))
	clientOptions.SetMaxConnIdleTime(30 * time.Minute)

	// Create client
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to MongoDB: %w", err)
	}

	// Check connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("unable to ping MongoDB: %w", err)
	}

	logger.Info("Connected to MongoDB database",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database))

	// Get database
	db := client.Database(cfg.Database)

	return &MongoDB{
		Client: client,
		DB:     db,
		logger: logger,
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	if m.Client != nil {
		if err := m.Client.Disconnect(ctx); err != nil {
			return fmt.Errorf("error disconnecting from MongoDB: %w", err)
		}
		m.logger.Info("MongoDB connection closed")
	}
	return nil
}

// GetCollection returns a MongoDB collection
func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.DB.Collection(name)
}
