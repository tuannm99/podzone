package pdsql

type (
	ProviderType string
	Config       struct {
		Provider           ProviderType `mapstructure:"provider"`
		URI                string       `mapstructure:"uri"`
		ShouldRunMigration bool         `mapstructure:"should_run_migration"`
	}
)

const (
	PostgresProvider ProviderType = "postgres"
	SqliteProvider   ProviderType = "sqlite"
	MysqlProvider    ProviderType = "mysql"
)
