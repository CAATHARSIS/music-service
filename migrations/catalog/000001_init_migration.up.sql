-- extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_trgm;
-- genres table
CREATE TABLE IF NOT EXISTS genres (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_name_len CHECK (char_length(name) <= 100)
);
-- artists table
CREATE TABLE IF NOT EXISTS artists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT UNIQUE NOT NULL,
    country TEXT,
    avatar_image_id UUID,
    total_plays BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_country_len CHECK (char_length(country) <= 100)
);
-- search vector for artists
ALTER TABLE artists
ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(name, '')), 'A')
    ) STORED;
CREATE INDEX IF NOT EXISTS idx_artists_search ON artists USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_artists_name_trgm ON artists USING GIN (name gin_trgm_ops);
-- album table
CREATE TABLE IF NOT EXISTS albums (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    year INT,
    artist_id UUID NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    cover_image_id UUID,
    album_type TEXT DEFAULT 'album',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_title_len CHECK (char_length(title) <= 255),
    CONSTRAINT check_year_badges CHECK (
        year > 1900
        AND year <= EXTRACT(
            YEAR
            FROM NOW()
        )
    ),
    CONSTRAINT check_album_type_len CHECK (char_length(album_type) <= 30)
);
-- search vector for albums
ALTER TABLE albums
ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(title, '')), 'A')
    ) STORED;
CREATE INDEX IF NOT EXISTS idx_albums_search ON albums USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_albums_title_trgm ON albums USING GIN (title gin_trgm_ops);
-- track table
CREATE TABLE IF NOT EXISTS tracks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    duration INT NOT NULL,
    year INT NOT NULL,
    artist_id UUID NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    album_id UUID REFERENCES albums(id) ON DELETE CASCADE,
    file_id UUID NOT NULL,
    cover_image_id UUID,
    plays_count BIGINT DEFAULT 0,
    track_number INT,
    lyrics TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_title_len CHECK (char_length(title) <= 255),
    CONSTRAINT check_year_badges CHECK (
        year > 1900
        AND year <= EXTRACT(
            YEAR
            FROM NOW()
        )
    )
);
-- search vector for tracks
ALTER TABLE tracks
ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(title, '')), 'A')
    ) STORED;
-- create indexes
CREATE INDEX IF NOT EXISTS idx_tracks_search ON tracks USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_tracks_title_trgm ON tracks USING GIN (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_tracks_artist_id ON tracks(artist_id);
CREATE INDEX IF NOT EXISTS idx_tracks_album_id ON tracks(album_id);
CREATE INDEX IF NOT EXISTS idx_tracks_year ON tracks(year);
CREATE INDEX IF NOT EXISTS idx_tracks_plays ON tracks(plays_count DESC);
-- relationship tracks to genres (many to many)
CREATE TABLE IF NOT EXISTS track_genres (
    track_id UUID NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (track_id, genre_id)
);
CREATE INDEX IF NOT EXISTS idx_track_genres_genre_id ON track_genres(genre_id);
-- relationship artists to genres (many to many)
CREATE TABLE IF NOT EXISTS artist_genres (
    artist_id UUID NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (artist_id, genre_id)
);
--relationship albums to genres (many to many)
CREATE TABLE IF NOT EXISTS album_genres (
    album_id UUID NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (album_id, genre_id)
)