package sqlstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/tuannm99/podzone/pkg/messaging"
)

type OutboxStore struct {
	db        *sqlx.DB
	tableName string
}

var _ messaging.OutboxStore = (*OutboxStore)(nil)

func NewOutboxStore(db *sqlx.DB, tableName string) (*OutboxStore, error) {
	if tableName == "" {
		tableName = "message_outbox"
	}
	if err := validateQualifiedIdentifier(tableName); err != nil {
		return nil, err
	}
	return &OutboxStore{db: db, tableName: tableName}, nil
}

func (s *OutboxStore) Append(ctx context.Context, tx messaging.Tx, record messaging.OutboxRecord) error {
	envelopeJSON, err := json.Marshal(record.Envelope)
	if err != nil {
		return fmt.Errorf("marshal outbox envelope: %w", err)
	}

	runner := sqlx.ExtContext(s.db)
	if txRunner, ok := tx.(sqlx.ExtContext); ok && txRunner != nil {
		runner = txRunner
	}

	_, err = runner.ExecContext(
		ctx,
		fmt.Sprintf(`
			INSERT INTO %s
				(id, topic, message_key, envelope_json, status, attempts,
				 next_attempt_at, published_at, error_text, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		`, s.tableName),
		record.ID,
		record.Topic,
		record.MessageKey,
		envelopeJSON,
		record.Status,
		record.Attempts,
		record.NextAttemptAt,
		record.PublishedAt,
		record.ErrorText,
		record.CreatedAt,
		record.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("append outbox row: %w", err)
	}
	return nil
}

func (s *OutboxStore) ListPending(ctx context.Context, limit int) ([]messaging.OutboxRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	rows := []outboxRow{}
	if err := s.db.SelectContext(ctx, &rows, fmt.Sprintf(`
		SELECT id, topic, message_key, envelope_json, status, attempts,
		       next_attempt_at, published_at, error_text, created_at, updated_at
		FROM %s
		WHERE status = 'pending' AND next_attempt_at <= now()
		ORDER BY created_at ASC
		LIMIT $1
	`, s.tableName), limit); err != nil {
		return nil, fmt.Errorf("list pending outbox rows: %w", err)
	}

	records := make([]messaging.OutboxRecord, 0, len(rows))
	for _, row := range rows {
		record, err := row.toRecord()
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *OutboxStore) MarkPublished(ctx context.Context, ids []string, publishedAt time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		SET status = 'published',
		    published_at = $2,
		    updated_at = $2
		WHERE id = ANY($1)
	`, s.tableName), pq.Array(ids), publishedAt)
	if err != nil {
		return fmt.Errorf("mark outbox rows published: %w", err)
	}
	return nil
}

func (s *OutboxStore) MarkFailed(ctx context.Context, id string, errText string, nextAttemptAt time.Time) error {
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		SET status = 'pending',
		    attempts = attempts + 1,
		    error_text = $2,
		    next_attempt_at = $3,
		    updated_at = now()
		WHERE id = $1
	`, s.tableName), id, errText, nextAttemptAt)
	if err != nil {
		return fmt.Errorf("mark outbox row failed: %w", err)
	}
	return nil
}

type outboxRow struct {
	ID            string     `db:"id"`
	Topic         string     `db:"topic"`
	MessageKey    string     `db:"message_key"`
	EnvelopeJSON  []byte     `db:"envelope_json"`
	Status        string     `db:"status"`
	Attempts      int        `db:"attempts"`
	NextAttemptAt time.Time  `db:"next_attempt_at"`
	PublishedAt   *time.Time `db:"published_at"`
	ErrorText     string     `db:"error_text"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

func (r outboxRow) toRecord() (messaging.OutboxRecord, error) {
	var envelope messaging.Envelope
	if err := json.Unmarshal(r.EnvelopeJSON, &envelope); err != nil {
		return messaging.OutboxRecord{}, fmt.Errorf("unmarshal outbox envelope %q: %w", r.ID, err)
	}
	return messaging.OutboxRecord{
		ID:            r.ID,
		Topic:         r.Topic,
		MessageKey:    r.MessageKey,
		Envelope:      envelope,
		Status:        r.Status,
		Attempts:      r.Attempts,
		NextAttemptAt: r.NextAttemptAt,
		PublishedAt:   r.PublishedAt,
		ErrorText:     r.ErrorText,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}, nil
}
