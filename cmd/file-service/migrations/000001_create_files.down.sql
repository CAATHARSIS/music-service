DROP INDEX IF EXISTS idx_files_bucket_key;
DROP INDEX IF EXISTS idx_files_uploaded_by;

DROP TABLE IF EXISTS files;

DROP EXTENSION IF EXISTS "uuid-ossp";