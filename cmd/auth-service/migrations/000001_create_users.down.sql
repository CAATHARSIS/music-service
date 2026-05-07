DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS INDEX idx_users_email;

DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";