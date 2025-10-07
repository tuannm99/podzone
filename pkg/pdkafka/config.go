// Podzone pdkafka
// Implements: config + Fx module, producer, consumer (pub/sub & queue),
// delivery semantics (at-most/at-least), request/reply helper, and
// transactional (EOS) skeleton.

package pdkafka

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

/* Example YAML
kafka:
  sample:
    brokers: ["localhost:9092"]
    client_id: podzone-auth
    producer:
      enabled: true
      idempotent: true
      required_acks: all
    consumer:
      enabled: true
      group: auth
      initial_offset: earliest
      isolation_level: read_uncommitted
*/

// -----------------------------------------------------------------------------
// Config
// -----------------------------------------------------------------------------

type Config struct {
	Brokers     []string      `mapstructure:"brokers"`
	ClientID    string        `mapstructure:"client_id"`
	Version     string        `mapstructure:"version"` // e.g. "3.6.0"; empty = default
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
	MetadataTO  time.Duration `mapstructure:"metadata_timeout"`

	Producer struct {
		Enabled         bool          `mapstructure:"enabled"`
		Idempotent      bool          `mapstructure:"idempotent"`
		TransactionalID string        `mapstructure:"transactional_id"`
		RequiredAcks    string        `mapstructure:"required_acks"` // none|local|all
		Compression     string        `mapstructure:"compression"`   // none|gzip|snappy|lz4|zstd
		MaxMessageBytes int           `mapstructure:"max_message_bytes"`
		FlushMessages   int           `mapstructure:"flush_messages"`
		FlushFrequency  time.Duration `mapstructure:"flush_frequency"`
		ReturnSuccesses bool          `mapstructure:"return_successes"`
		ReturnErrors    bool          `mapstructure:"return_errors"`
	} `mapstructure:"producer"`

	Consumer struct {
		Enabled           bool          `mapstructure:"enabled"`
		Group             string        `mapstructure:"group"`
		Assignor          string        `mapstructure:"assignor"`       // range|roundrobin|sticky
		InitialOffset     string        `mapstructure:"initial_offset"` // earliest|latest
		SessionTimeout    time.Duration `mapstructure:"session_timeout"`
		HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
		RebalanceTimeout  time.Duration `mapstructure:"rebalance_timeout"`
		IsolationLevel    string        `mapstructure:"isolation_level"` // read_uncommitted|read_committed
	} `mapstructure:"consumer"`

	TLS struct {
		Enable   bool   `mapstructure:"enable"`
		CAFile   string `mapstructure:"ca_file"`
		CertFile string `mapstructure:"cert_file"`
		KeyFile  string `mapstructure:"key_file"`
		Insecure bool   `mapstructure:"insecure_skip_verify"`
	} `mapstructure:"tls"`

	SASL struct {
		Enable    bool   `mapstructure:"enable"`
		Mechanism string `mapstructure:"mechanism"` // PLAIN|SCRAM-SHA-256|SCRAM-SHA-512|AWS_MSK_IAM
		User      string `mapstructure:"user"`
		Password  string `mapstructure:"password"`
	} `mapstructure:"sasl"`
}

func NewConfigFromViper(v *viper.Viper, name string) (*Config, error) {
	key := fmt.Sprintf("kafka.%s", name)
	if !v.IsSet(key) {
		return nil, fmt.Errorf("missing config key %q", key)
	}
	cfg := DefaultConfig()
	sub := v.Sub(key)
	if sub == nil {
		return nil, fmt.Errorf("invalid config section %q", key)
	}
	if err := sub.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func DefaultConfig() *Config {
	c := &Config{}
	c.Brokers = []string{"localhost:9092"}
	c.ClientID = "podzone"
	c.Producer.RequiredAcks = "all"
	c.Producer.ReturnSuccesses = true
	c.Consumer.InitialOffset = "latest"
	c.Consumer.Assignor = "range"
	c.Consumer.IsolationLevel = "read_uncommitted"
	return c
}

// -----------------------------------------------------------------------------
// Sarama config builder
// -----------------------------------------------------------------------------

func buildSaramaConfig(c *Config) (*sarama.Config, error) {
	sc := sarama.NewConfig()
	sc.ClientID = firstNonEmpty(c.ClientID, "podzone")

	if v, err := parseKafkaVersion(c.Version); err == nil {
		sc.Version = v
	}
	if c.DialTimeout > 0 {
		sc.Net.DialTimeout = c.DialTimeout
	}
	if c.MetadataTO > 0 {
		sc.Metadata.RefreshFrequency = c.MetadataTO
	}

	// Producer
	sc.Producer.Return.Successes = c.Producer.ReturnSuccesses
	sc.Producer.Return.Errors = c.Producer.ReturnErrors
	sc.Producer.Idempotent = c.Producer.Idempotent
	if c.Producer.TransactionalID != "" {
		sc.Producer.Transaction.ID = c.Producer.TransactionalID
	}
	if c.Producer.MaxMessageBytes > 0 {
		sc.Producer.MaxMessageBytes = c.Producer.MaxMessageBytes
	}
	if c.Producer.FlushMessages > 0 {
		sc.Producer.Flush.Messages = c.Producer.FlushMessages
	}
	if c.Producer.FlushFrequency > 0 {
		sc.Producer.Flush.Frequency = c.Producer.FlushFrequency
	}
	sc.Producer.RequiredAcks = mapRequiredAcks(c.Producer.RequiredAcks)
	sc.Producer.Compression = mapCompression(c.Producer.Compression)

	// Consumer
	sc.Consumer.Group.Rebalance.Strategy = mapAssignor(c.Consumer.Assignor)
	sc.Consumer.Offsets.Initial = mapInitialOffset(c.Consumer.InitialOffset)
	if strings.EqualFold(c.Consumer.IsolationLevel, "read_committed") {
		sc.Consumer.IsolationLevel = sarama.ReadCommitted
	}
	if c.Consumer.SessionTimeout > 0 {
		sc.Consumer.Group.Session.Timeout = c.Consumer.SessionTimeout
	}
	if c.Consumer.HeartbeatInterval > 0 {
		sc.Consumer.Group.Heartbeat.Interval = c.Consumer.HeartbeatInterval
	}
	if c.Consumer.RebalanceTimeout > 0 {
		sc.Consumer.Group.Rebalance.Timeout = c.Consumer.RebalanceTimeout
	}

	// TLS/SASL
	if c.TLS.Enable {
		tlsCfg, err := buildTLS(c)
		if err != nil {
			return nil, err
		}
		sc.Net.TLS.Enable = true
		sc.Net.TLS.Config = tlsCfg
	}
	if c.SASL.Enable {
		sc.Net.SASL.Enable = true
		sc.Net.SASL.User = c.SASL.User
		sc.Net.SASL.Password = c.SASL.Password
		sc.Net.SASL.Mechanism = sarama.SASLMechanism(strings.ToUpper(c.SASL.Mechanism))
	}
	return sc, nil
}

func buildTLS(c *Config) (*tls.Config, error) {
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: c.TLS.Insecure} //nolint:gosec
	if c.TLS.CAFile != "" {
		pem, err := os.ReadFile(c.TLS.CAFile)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(pem)
		tlsCfg.RootCAs = pool
	}
	if c.TLS.CertFile != "" && c.TLS.KeyFile != "" {
		crt, err := tls.LoadX509KeyPair(c.TLS.CertFile, c.TLS.KeyFile)
		if err != nil {
			return nil, err
		}
		tlsCfg.Certificates = []tls.Certificate{crt}
	}
	return tlsCfg, nil
}

func parseKafkaVersion(v string) (sarama.KafkaVersion, error) {
	if strings.TrimSpace(v) == "" {
		return sarama.KafkaVersion{}, fmt.Errorf("empty version")
	}
	return sarama.ParseKafkaVersion(v)
}

func mapRequiredAcks(s string) sarama.RequiredAcks {
	switch strings.ToLower(s) {
	case "none":
		return sarama.NoResponse
	case "local":
		return sarama.WaitForLocal
	default:
		return sarama.WaitForAll
	}
}

func mapCompression(s string) sarama.CompressionCodec {
	switch strings.ToLower(s) {
	case "gzip":
		return sarama.CompressionGZIP
	case "snappy":
		return sarama.CompressionSnappy
	case "lz4":
		return sarama.CompressionLZ4
	case "zstd":
		return sarama.CompressionZSTD
	default:
		return sarama.CompressionNone
	}
}

func mapAssignor(s string) sarama.BalanceStrategy {
	switch strings.ToLower(s) {
	case "roundrobin":
		return sarama.NewBalanceStrategyRoundRobin()
	case "sticky":
		return sarama.NewBalanceStrategySticky()
	default:
		return sarama.NewBalanceStrategyRange()
	}
}

func mapInitialOffset(s string) int64 {
	switch strings.ToLower(s) {
	case "earliest":
		return sarama.OffsetOldest
	default:
		return sarama.OffsetNewest
	}
}

func firstNonEmpty(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
