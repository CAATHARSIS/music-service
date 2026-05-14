CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    condition JSONB NOT NULL,
    track_limit INT DEFAULT 50,
    cron_schedule VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    last_executed TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rules_user_id ON rules(user_id);