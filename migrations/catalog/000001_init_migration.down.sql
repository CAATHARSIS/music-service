-- relationships (many-to-many)
DROP TABLE IF EXISTS album_genres;
DROP TABLE IF EXISTS artist_genres;
DROP TABLE IF EXISTS track_genres;
-- indexes
DROP INDEX IF EXISTS idx_tracks_plays;
DROP INDEX IF EXISTS idx_tracks_year;
DROP INDEX IF EXISTS idx_tracks_album_id;
DROP INDEX IF EXISTS idx_tracks_artist_id;
DROP INDEX IF EXISTS idx_tracks_title_trgm;
DROP INDEX IF EXISTS idx_tracks_search;
ALTER TABLE tracks DROP COLUMN IF EXISTS search_vector;
-- tracks table
DROP TABLE IF EXISTS tracks;
--indexes
DROP INDEX IF EXISTS idx_albums_title_trgm;
DROP INDEX IF EXISTS idx_albums_search;
-- album table
ALTER TABLE albums DROP COLUMN IF EXISTS search_vector;
DROP TABLE IF EXISTS albums;
-- indexes
DROP INDEX IF EXISTS idx_artists_name_trgm;
DROP INDEX IF EXISTS idx_artists_search;
ALTER TABLE artists DROP COLUMN IF EXISTS search_vector;
-- artist table
DROP TABLE IF EXISTS artists;
-- genre table
DROP TABLE IF EXISTS genres;
-- extensions
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "uuid-ossp";