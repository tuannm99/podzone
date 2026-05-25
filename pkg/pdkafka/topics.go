package pdkafka

import (
	"errors"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/tuannm99/podzone/pkg/messaging"
)

func BootstrapTopics(admin Admin, cfg *Config) error {
	if admin == nil || cfg == nil || !cfg.Topics.Enabled {
		return nil
	}

	existing, err := admin.ListTopics()
	if err != nil {
		return fmt.Errorf("list kafka topics: %w", err)
	}

	topics := expandTopics(cfg.Topics)
	for _, topic := range topics {
		if _, ok := existing[topic]; ok {
			continue
		}
		if err := admin.CreateTopic(topic, &sarama.TopicDetail{
			NumPartitions:     cfg.Topics.DefaultPartitions,
			ReplicationFactor: cfg.Topics.DefaultReplication,
		}, false); err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {
			return fmt.Errorf("create kafka topic %s: %w", topic, err)
		}
	}
	return nil
}

func expandTopics(cfg messaging.TopicBootstrapConfig) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(cfg.MainTopics))
	for _, mainTopic := range cfg.MainTopics {
		if mainTopic == "" {
			continue
		}
		if _, ok := seen[mainTopic]; !ok {
			out = append(out, mainTopic)
			seen[mainTopic] = struct{}{}
		}
		for _, attempt := range cfg.RetryAttempts {
			retryTopic := messaging.RetryTopic(mainTopic, attempt)
			if _, ok := seen[retryTopic]; !ok {
				out = append(out, retryTopic)
				seen[retryTopic] = struct{}{}
			}
		}
		if cfg.CreateDeadLetter {
			deadTopic := messaging.DeadLetterTopic(mainTopic)
			if _, ok := seen[deadTopic]; !ok {
				out = append(out, deadTopic)
				seen[deadTopic] = struct{}{}
			}
		}
	}
	return out
}
