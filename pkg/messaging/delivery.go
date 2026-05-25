package messaging

import (
	"strconv"
	"strings"
)

const (
	HeaderAttempt          = "podzone.attempt"
	HeaderMaxAttempts      = "podzone.max_attempts"
	HeaderOriginalTopic    = "podzone.original_topic"
	HeaderLastError        = "podzone.last_error"
	HeaderDeadLetterReason = "podzone.dead_letter_reason"
	HeaderRedriveCount     = "podzone.redrive_count"
	HeaderConsumerName     = "podzone.consumer_name"
)

type DeliveryMetadata struct {
	Attempt          int
	MaxAttempts      int
	OriginalTopic    string
	LastError        string
	DeadLetterReason string
	RedriveCount     int
	ConsumerName     string
}

func ReadDeliveryMetadata(env Envelope) DeliveryMetadata {
	headers := env.Headers
	if headers == nil {
		return DeliveryMetadata{}
	}
	return DeliveryMetadata{
		Attempt:          parseHeaderInt(headers[HeaderAttempt]),
		MaxAttempts:      parseHeaderInt(headers[HeaderMaxAttempts]),
		OriginalTopic:    strings.TrimSpace(headers[HeaderOriginalTopic]),
		LastError:        strings.TrimSpace(headers[HeaderLastError]),
		DeadLetterReason: strings.TrimSpace(headers[HeaderDeadLetterReason]),
		RedriveCount:     parseHeaderInt(headers[HeaderRedriveCount]),
		ConsumerName:     strings.TrimSpace(headers[HeaderConsumerName]),
	}
}

func WithDeliveryMetadata(env Envelope, metadata DeliveryMetadata) Envelope {
	clone := env.Clone()
	if clone.Headers == nil {
		clone.Headers = make(map[string]string)
	}
	setHeaderInt(clone.Headers, HeaderAttempt, metadata.Attempt)
	setHeaderInt(clone.Headers, HeaderMaxAttempts, metadata.MaxAttempts)
	setHeaderValue(clone.Headers, HeaderOriginalTopic, metadata.OriginalTopic)
	setHeaderValue(clone.Headers, HeaderLastError, metadata.LastError)
	setHeaderValue(clone.Headers, HeaderDeadLetterReason, metadata.DeadLetterReason)
	setHeaderInt(clone.Headers, HeaderRedriveCount, metadata.RedriveCount)
	setHeaderValue(clone.Headers, HeaderConsumerName, metadata.ConsumerName)
	return clone
}

func (e Envelope) Clone() Envelope {
	clone := e
	if len(e.Headers) > 0 {
		clone.Headers = make(map[string]string, len(e.Headers))
		for k, v := range e.Headers {
			clone.Headers[k] = v
		}
	}
	if len(e.Payload) > 0 {
		clone.Payload = append(clone.Payload[:0:0], e.Payload...)
	}
	return clone
}

func parseHeaderInt(raw string) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	return value
}

func setHeaderInt(headers map[string]string, key string, value int) {
	if value <= 0 {
		delete(headers, key)
		return
	}
	headers[key] = strconv.Itoa(value)
}

func setHeaderValue(headers map[string]string, key string, value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		delete(headers, key)
		return
	}
	headers[key] = trimmed
}
