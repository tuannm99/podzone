package messaging

import (
	"fmt"
	"strings"
)

type TopicStrategy struct {
	RetrySuffix string
	DeadSuffix  string
}

func DefaultTopicStrategy() TopicStrategy {
	return TopicStrategy{
		RetrySuffix: ".retry",
		DeadSuffix:  ".dlt",
	}
}

func (s TopicStrategy) RetryTopic(topic string, attempt int) string {
	retrySuffix := s.RetrySuffix
	if strings.TrimSpace(retrySuffix) == "" {
		retrySuffix = DefaultTopicStrategy().RetrySuffix
	}
	return fmt.Sprintf("%s%s.%d", topic, retrySuffix, attempt)
}

func (s TopicStrategy) DeadLetterTopic(topic string) string {
	deadSuffix := s.DeadSuffix
	if strings.TrimSpace(deadSuffix) == "" {
		deadSuffix = DefaultTopicStrategy().DeadSuffix
	}
	return topic + deadSuffix
}

type DeadLetterPolicy struct {
	Strategy TopicStrategy
}

func (p DeadLetterPolicy) TopicFor(topic string) string {
	return p.Strategy.DeadLetterTopic(topic)
}
