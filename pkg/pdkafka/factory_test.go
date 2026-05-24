package pdkafka

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSaramaConfig_MapsCoreSettings(t *testing.T) {
	cfg := &Config{
		Brokers:          []string{"localhost:9092"},
		ClientID:         "podzone-iam",
		Version:          "3.7.0",
		RequiredAcks:     RequiredAcksAll,
		Compression:      CompressionZSTD,
		AutoCreateTopics: true,
	}

	scfg, err := NewSaramaConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, scfg)

	assert.Equal(t, "podzone-iam", scfg.ClientID)
	assert.True(t, scfg.Metadata.AllowAutoTopicCreation)
	assert.True(t, scfg.Producer.Idempotent)
	assert.Equal(t, sarama.WaitForAll, scfg.Producer.RequiredAcks)
	assert.Equal(t, sarama.CompressionZSTD, scfg.Producer.Compression)
	assert.Equal(t, 1, scfg.Net.MaxOpenRequests)
	assert.Equal(t, sarama.OffsetNewest, scfg.Consumer.Offsets.Initial)
}

func TestNewSaramaConfig_MapsTLSAndSASL(t *testing.T) {
	cfg := &Config{
		Brokers:  []string{"localhost:9092"},
		ClientID: "podzone-auth",
		Version:  "3.7.0",
		SASL: SASLConfig{
			Enabled:   true,
			Mechanism: "SCRAM-SHA-256",
			Username:  "user",
			Password:  "pass",
		},
		TLS: TLSConfig{Enabled: true},
	}

	scfg, err := NewSaramaConfig(cfg)
	require.NoError(t, err)
	assert.True(t, scfg.Net.SASL.Enable)
	assert.Equal(t, string(sarama.SASLTypeSCRAMSHA256), string(scfg.Net.SASL.Mechanism))
	assert.True(t, scfg.Net.TLS.Enable)
	require.NotNil(t, scfg.Net.TLS.Config)
}

func TestNewSaramaConfig_InvalidInputs(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
	}{
		{
			name: "nil config",
			cfg:  nil,
		},
		{
			name: "invalid version",
			cfg: &Config{
				Brokers:  []string{"localhost:9092"},
				ClientID: "podzone",
				Version:  "not-a-version",
			},
		},
		{
			name: "unsupported sasl",
			cfg: &Config{
				Brokers:  []string{"localhost:9092"},
				ClientID: "podzone",
				Version:  "3.7.0",
				SASL: SASLConfig{
					Enabled:   true,
					Mechanism: "GSSAPI-UNKNOWN",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scfg, err := NewSaramaConfig(tt.cfg)
			require.Error(t, err)
			assert.Nil(t, scfg)
		})
	}
}

func TestConsumerGroupFactory_NewRequiresGroupID(t *testing.T) {
	factory := consumerGroupFactory{
		brokers: []string{"localhost:9092"},
		config:  sarama.NewConfig(),
	}

	group, err := factory.New("   ")
	require.Error(t, err)
	assert.Nil(t, group)
}

func TestNewSyncProducerFromConfigRejectsNilConfig(t *testing.T) {
	producer, err := NewSyncProducerFromConfig(nil, sarama.NewConfig())
	require.Error(t, err)
	assert.Nil(t, producer)
}

func TestNewClusterAdminFromConfigRejectsNilConfig(t *testing.T) {
	admin, err := NewClusterAdminFromConfig(nil, sarama.NewConfig())
	require.Error(t, err)
	assert.Nil(t, admin)
}
