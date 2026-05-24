CREATE INDEX IF NOT EXISTS idx_tracks_artist_id ON tracks(artist_id);
DROP INDEX IF EXISTS idx_album_artists_artist_id;
DROP INDEX IF EXISTS idx_track_artists_artist_id;
ALTER TABLE albums
ADD COLUMN IF NOT EXISTS artist_id UUID REFERENCES artists(id) ON DELETE CASCADE;
ALTER TABLE tracks
ADD COLUMN IF NOT EXISTS artist_id UUID REFERENCES artists(id) ON DELETE CASCADE;
UPDATE tracks t
SET artist_id = (
        SELECT ta.artist_id
        FROM track_artists ta
        WHERE ta.track_id = t.id
        LIMIT 1
    );
UPDATE tracks t
SET artist_id = (
        SELECT aa.artist_id
        FROM album_artists aa
        WHERE aa.album_id = t.id
        LIMIT 1
    );
DROP TABLE IF EXISTS album_artists;
DROP TABLE IF EXISTS track_artists;