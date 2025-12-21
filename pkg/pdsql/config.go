package pdsql

type (
	ProviderType string
	Config       struct {
		Provider           ProviderType `koanf:"provider"             mapstructure:"provider"`
		URI                string       `koanf:"uri"                  mapstructure:"uri"`
		ShouldRunMigration bool         `koanf:"should_run_migration" mapstructure:"should_run_migration"`
	}
)

const (
	PostgresProvider ProviderType = "postgres"
	SqliteProvider   ProviderType = "sqlite"
	MysqlProvider    ProviderType = "mysql"
)
