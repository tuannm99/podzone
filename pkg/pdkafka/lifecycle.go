package pdkafka

import (
	"context"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

func registerLifecycle(lc fx.Lifecycle, producer Producer, admin Admin, log pdlog.Logger, cfg *Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			_, controllerID, err := admin.DescribeCluster()
			if err != nil {
				log.Error("Kafka cluster probe failed", "error", err, "brokers", cfg.Brokers)
				return err
			}
			log.Info(
				"Kafka cluster probe OK",
				"brokers",
				cfg.Brokers,
				"controller_id",
				controllerID,
				"client_id",
				cfg.ClientID,
			)
			if err := BootstrapTopics(admin, cfg); err != nil {
				log.Error("Kafka topic bootstrap failed", "error", err, "client_id", cfg.ClientID)
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing Kafka clients", "client_id", cfg.ClientID)
			if err := producer.Close(); err != nil {
				log.Error("Close Kafka producer failed", "error", err)
				return err
			}
			if err := admin.Close(); err != nil {
				log.Error("Close Kafka admin failed", "error", err)
				return err
			}
			return nil
		},
	})
}
