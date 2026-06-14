CREATE TABLE IF NOT EXISTS generated_playlists (
    rule_id UUID PRIMARY KEY REFERENCES rules(id) ON DELETE CASCADE,
    playlist_id UUID NOT NULL,
    generated_at TIMESTAMP DEFAULT NOW()
);