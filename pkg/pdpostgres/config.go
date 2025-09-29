package pdpostgres

type (
	ProviderType string
	Config       struct {
		Provider ProviderType `mapstructure:"provider"`
		URI      string       `mapstructure:"uri"`
	}
)

const (
	PostgresProvider ProviderType = "postgres"
	SqliteProvider   ProviderType = "sqlite"
	MysqlProvider    ProviderType = "mysql"
)
