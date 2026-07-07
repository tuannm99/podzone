package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/tuannm99/podzone/pkg/messaging"
)

type InboxStore struct {
	db        *sqlx.DB
	tableName string
}

var _ messaging.InboxStore = (*InboxStore)(nil)

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func NewInboxStore(db *sqlx.DB, tableName string) (*InboxStore, error) {
	if tableName == "" {
		tableName = "message_inbox"
	}
	if err := validateQualifiedIdentifier(tableName); err != nil {
		return nil, err
	}
	return &InboxStore{db: db, tableName: tableName}, nil
}

func (s *InboxStore) Begin(
	ctx context.Context,
	consumerName string,
	messageID string,
	now time.Time,
) (messaging.InboxDecision, error) {
	result, err := s.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		SET status = 'processing',
		    error_text = '',
		    started_at = $3,
		    updated_at = $3
		WHERE consumer_name = $1
		  AND message_id = $2
		  AND status = 'failed'
	`, s.tableName), consumerName, messageID, now)
	if err != nil {
		return "", fmt.Errorf("reset failed inbox row: %w", err)
	}
	if rows, err := result.RowsAffected(); err == nil && rows > 0 {
		return messaging.InboxDecisionAcquired, nil
	}

	result, err = s.db.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s (
			consumer_name,
			message_id,
			status,
			started_at,
			updated_at
		) VALUES ($1, $2, 'processing', $3, $3)
		ON CONFLICT DO NOTHING
	`, s.tableName), consumerName, messageID, now)
	if err != nil {
		return "", fmt.Errorf("insert inbox row: %w", err)
	}
	rows, err := result.RowsAffected()
	if err == nil && rows > 0 {
		return messaging.InboxDecisionAcquired, nil
	}

	var status string
	if err := s.db.GetContext(ctx, &status, fmt.Sprintf(`
		SELECT status
		FROM %s
		WHERE consumer_name = $1 AND message_id = $2
	`, s.tableName), consumerName, messageID); err != nil {
		if err != sql.ErrNoRows {
			return "", fmt.Errorf("lookup inbox row: %w", err)
		}
		// Row disappeared between the INSERT conflict and the SELECT (e.g. cleaned up
		// by a maintenance job). Retry the INSERT once; the slot is now free.
		result2, err2 := s.db.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO %s (
				consumer_name,
				message_id,
				status,
				started_at,
				updated_at
			) VALUES ($1, $2, 'processing', $3, $3)
			ON CONFLICT DO NOTHING
		`, s.tableName), consumerName, messageID, now)
		if err2 != nil {
			return "", fmt.Errorf("retry insert inbox row: %w", err2)
		}
		if rows2, err2 := result2.RowsAffected(); err2 == nil && rows2 > 0 {
			return messaging.InboxDecisionAcquired, nil
		}
		// Another consumer acquired it in the narrow window; treat as in-progress.
		return messaging.InboxDecisionInProgress, nil
	}
	if status == "completed" {
		return messaging.InboxDecisionDuplicate, nil
	}
	return messaging.InboxDecisionInProgress, nil
}

func (s *InboxStore) Complete(ctx context.Context, consumerName string, messageID string, processedAt time.Time) error {
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		SET status = 'completed',
		    processed_at = $3,
		    updated_at = $3
		WHERE consumer_name = $1 AND message_id = $2
	`, s.tableName), consumerName, messageID, processedAt)
	if err != nil {
		return fmt.Errorf("complete inbox row: %w", err)
	}
	return nil
}

func (s *InboxStore) Fail(
	ctx context.Context,
	consumerName string,
	messageID string,
	errText string,
	failedAt time.Time,
) error {
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		SET status = 'failed',
		    error_text = $3,
		    updated_at = $4
		WHERE consumer_name = $1 AND message_id = $2
	`, s.tableName), consumerName, messageID, errText, failedAt)
	if err != nil {
		return fmt.Errorf("fail inbox row: %w", err)
	}
	return nil
}

func validateQualifiedIdentifier(name string) error {
	for _, segment := range strings.Split(name, ".") {
		if !validIdentifier.MatchString(segment) {
			return fmt.Errorf("invalid inbox table name %q", name)
		}
	}
	return nil
}
