ALTER TABLE auth_sessions
ADD COLUMN IF NOT EXISTS session_policy_json TEXT NOT NULL DEFAULT '[]';
