package messaging

import "fmt"

func Topic(service string, stream string) string {
	return fmt.Sprintf("podzone.%s.%s", service, stream)
}

func EventTopic(service string) string {
	return Topic(service, "events")
}

func RetryTopic(topic string, attempt int) string {
	return DefaultTopicStrategy().RetryTopic(topic, attempt)
}

func DeadLetterTopic(topic string) string {
	return DefaultTopicStrategy().DeadLetterTopic(topic)
}

func TopicsWithRetry(mainTopic string, maxAttempts int) []string {
	if maxAttempts <= 0 {
		return []string{mainTopic}
	}
	topics := make([]string, 0, maxAttempts+1)
	topics = append(topics, mainTopic)
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		topics = append(topics, RetryTopic(mainTopic, attempt))
	}
	return topics
}
