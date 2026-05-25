package messaging

import "time"

type OutboxRecord struct {
	ID            string
	Topic         string
	MessageKey    string
	Envelope      Envelope
	Status        string
	Attempts      int
	NextAttemptAt time.Time
	PublishedAt   *time.Time
	ErrorText     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
