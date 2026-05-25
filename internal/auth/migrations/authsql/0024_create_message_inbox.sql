-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS message_inbox (
  consumer_name TEXT NOT NULL,
  message_id TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'processing',
  error_text TEXT NOT NULL DEFAULT '',
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  processed_at TIMESTAMPTZ NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (consumer_name, message_id)
);

CREATE INDEX IF NOT EXISTS idx_message_inbox_status
  ON message_inbox(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_message_inbox_status;
DROP TABLE IF EXISTS message_inbox;
-- +goose StatementEnd
