package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport"
	"github.com/tuannm99/podzone/pkg/messaging"
	"go.uber.org/fx"
)

type outboxRepoParams struct {
	fx.In
	DB *sqlx.DB `name:"sql-iam"`
}

type OutboxRepositoryImpl struct {
	db *sqlx.DB
}

var _ outputport.OutboxRepository = (*OutboxRepositoryImpl)(nil)

func NewOutboxRepository(p outboxRepoParams) outputport.OutboxRepository {
	return &OutboxRepositoryImpl{db: p.DB}
}

func (r *OutboxRepositoryImpl) Append(ctx context.Context, tx messaging.Tx, record messaging.OutboxRecord) error {
	envelopeJSON, err := json.Marshal(record.Envelope)
	if err != nil {
		return err
	}
	execer := r.db
	if runner, ok := tx.(sqlx.ExtContext); ok && tx != nil {
		_, err = runner.ExecContext(
			ctx,
			`INSERT INTO message_outbox
				(id, topic, message_key, envelope_json, status, attempts, next_attempt_at, published_at, error_text, created_at, updated_at)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
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
		return err
	}
	_, err = execer.ExecContext(
		ctx,
		`INSERT INTO message_outbox
			(id, topic, message_key, envelope_json, status, attempts, next_attempt_at, published_at, error_text, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
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
	return err
}

func (r *OutboxRepositoryImpl) ListPending(ctx context.Context, limit int) ([]messaging.OutboxRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	rows := []struct {
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
	}{}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, topic, message_key, envelope_json, status, attempts, next_attempt_at, published_at, error_text, created_at, updated_at
		FROM message_outbox
		WHERE status = 'pending' AND next_attempt_at <= now()
		ORDER BY created_at ASC
		LIMIT $1
	`, limit); err != nil {
		return nil, err
	}
	out := make([]messaging.OutboxRecord, 0, len(rows))
	for _, row := range rows {
		var envelope messaging.Envelope
		if err := json.Unmarshal(row.EnvelopeJSON, &envelope); err != nil {
			return nil, err
		}
		out = append(out, messaging.OutboxRecord{
			ID:            row.ID,
			Topic:         row.Topic,
			MessageKey:    row.MessageKey,
			Envelope:      envelope,
			Status:        row.Status,
			Attempts:      row.Attempts,
			NextAttemptAt: row.NextAttemptAt,
			PublishedAt:   row.PublishedAt,
			ErrorText:     row.ErrorText,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *OutboxRepositoryImpl) MarkPublished(ctx context.Context, ids []string, publishedAt time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE message_outbox
		SET status = 'published', published_at = $2, updated_at = $2
		WHERE id = ANY($1)
	`, ids, publishedAt)
	return err
}

func (r *OutboxRepositoryImpl) MarkFailed(
	ctx context.Context,
	id string,
	errText string,
	nextAttemptAt time.Time,
) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE message_outbox
		SET status = 'pending', attempts = attempts + 1, error_text = $2, next_attempt_at = $3, updated_at = now()
		WHERE id = $1
	`, id, errText, nextAttemptAt)
	return err
}
