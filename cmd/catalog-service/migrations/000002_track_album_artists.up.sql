CREATE TABLE IF NOT EXISTS track_artists (
    track_id UUID NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    artist_id UUID NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    PRIMARY KEY (track_id, artist_id)
);
CREATE TABLE IF NOT EXISTS album_artists (
    album_id UUID NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    artist_id UUID NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    PRIMARY KEY (album_id, artist_id)
);
INSERT INTO track_artists (track_id, artist_id)
SELECT id,
    artist_id
FROM tracks
WHERE artist_id IS NOT NULL;
INSERT INTO album_artists (album_id, artist_id)
SELECT id,
    artist_id
FROM albums
WHERE artist_id IS NOT NULL;
ALTER TABLE tracks DROP COLUMN IF EXISTS artist_id;
ALTER TABLE albums DROP COLUMN IF EXISTS artist_id;
CREATE INDEX IF NOT EXISTS idx_track_artists_artist_id ON track_artists(artist_id);
CREATE INDEX IF NOT EXISTS idx_album_artists_artist_id ON album_artists(artist_id);
DROP INDEX IF EXISTS idx_tracks_artist_id;