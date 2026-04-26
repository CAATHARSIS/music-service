package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/CAATHARSIS/music-service/internal/catalog/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrNotFound = errors.New("resource not found")
)

// Валидцию таблиц и колонок вынести в service!!!

var allowedTables = map[string]bool{
	"track_genres":  true,
	"artist_genres": true,
	"album_genres":  true,
}

var allowedColumns = map[string]bool{
	"track_id":  true,
	"artist_id": true,
	"album_id":  true,
}

type Repository interface {
	// Tracks
	CreateTrack(ctx context.Context, track *models.CreateTrackParams) (*models.Track, error)
	GetTrackByID(ctx context.Context, id string, opts *models.GetTrackOptions) (*models.Track, error)
	GetTrackByIDs(ctx context.Context, ids []string, opts *models.GetTrackOptions) ([]*models.Track, error)
	UpdateTrack(ctx context.Context, id string, params *models.UpdateTrackParams) (*models.Track, error)
	DeleteTrackByID(ctx context.Context, id string) error
	ListTracks(ctx context.Context, filter *models.TrackFilter) (*models.TrackListResult, error)
	IncrementPlays(ctx context.Context, trackID string, incrementBy int64) error
	SearchTracks(ctx context.Context, query string, opts *models.SearchOptions) ([]*models.Track, error)
}

type repository struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewRepository(db *sqlx.DB, logger *slog.Logger) Repository {
	return &repository{
		db:  db,
		log: logger,
	}
}

// Tracks

func (r *repository) CreateTrack(ctx context.Context, trackParams *models.CreateTrackParams) (*models.Track, error) {
	const query = `
	INSERT INTO
		TRACKS (
			ID,
			TITLE,
			DURATION,
			YEAR,
			ARTIST_ID,
			ALBUM_ID,
			FILE_ID,
			COVER_IMAGE_ID,
			TRACK_NUMBER,
			LYRICS,
			CREATED_AT,
			UPDATED_AT
		)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING
		ID,
		TITLE,
		DURATION,
		YEAR,
		FILE_ID,
		COVER_IMAGE_ID,
		TRACK_NUMBER,
		LYRICS,
		PLAYS_COUNT,
		CREATED_AT,
		UPDATED_AT
	`

	track := &models.Track{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := r.db.QueryRowContext(ctx, query,
		track.ID,
		trackParams.Title,
		trackParams.Duration,
		trackParams.Year,
		trackParams.ArtistID,
		trackParams.AlbumID,
		trackParams.FileID,
		trackParams.CoverImageID,
		trackParams.TrackNumber,
		trackParams.Lyrics,
		track.CreatedAt,
		track.UpdatedAt,
	).Scan(
		&track.ID,
		&track.Title,
		&track.Duration,
		&track.Year,
		&track.FileID,
		&track.CoverImageID,
		&track.TrackNumber,
		&track.Lyrics,
		&track.PlaysCount,
		&track.CreatedAt,
		&track.UpdatedAt,
	)

	if err != nil {
		r.log.Error("failed to crate track", "error", err)
		return nil, fmt.Errorf("create track: %w", err)
	}

	if len(trackParams.GenreIDs) > 0 {
		if err := r.addTrackGenres(ctx, track.ID, trackParams.GenreIDs); err != nil {
			r.log.Error("failed to add track genres", "error", err)
		}
	}

	// track.Artist, _ = r.getArtist(ctx, trackParams.ArtistID)
	// track.Album, _ = r.getAlbum(ctx, trackParams.AlbumID)
	// track.Genres, _ = r.getTrackGenres(ctx, track.ID)

	return track, nil
}

func (r *repository) GetTrackByID(ctx context.Context, id string, opts *models.GetTrackOptions) (*models.Track, error) {
	if opts == nil {
		opts = &models.GetTrackOptions{}
	}

	var track models.Track

	query := `
		SELECT
			t.id, t.title, t.duration, t.year, t.file_id, t.cover_image_id,
			t.track_number, t.lyrics, t.plays_count,
			t.artist_id, t.album_id, t.created_at, t.updated_at
	`

	if opts.IncludeArtist {
		query += `,
			a.id as "artist.id",
			a.name as "artist.name",
			a.country as "artist.country",
			a.avatar_image_id as "artist.avatar_image_id"
		`
	}

	if opts.IncludeAlbum {
		query += `,
			al.id as "album.id",
			al.title as "album.title",
			al.year as "album.year"
		`
	}

	query += ` FROM tracks t`

	if opts.IncludeArtist {
		query += ` JOIN artists a ON t.artist_id = a.id`
	}

	if opts.IncludeAlbum {
		query += ` LEFT JOIN albums al ON t.album_id = al.id`
	}

	query += ` WHERE t.id = $1`

	err := r.db.GetContext(ctx, &track, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get track: %w", err)
	}

	if opts.IncludeGenres {
		track.Genres, _ = r.getTrackGenres(ctx, id)
	}

	return &track, nil
}

func (r *repository) GetTrackByIDs(ctx context.Context, ids []string, opts *models.GetTrackOptions) ([]*models.Track, error) {
	if len(ids) == 0 {
		return []*models.Track{}, nil
	}

	if opts == nil {
		opts = &models.GetTrackOptions{}
	}

	query := r.buildGetTracksByIDsQuery(opts)

	var tracks []*models.Track
	err := r.db.SelectContext(ctx, &tracks, query, pq.Array(ids))
	if err != nil {
		r.log.Error("failed to get tracks by ids", "error", err)
		return nil, fmt.Errorf("get tracks by ids: %w", err)
	}

	if opts.IncludeGenres && len(tracks) > 0 {
		genresMap, err := r.getTracksGenresBatch(ctx, ids)
		if err == nil {
			for _, track := range tracks {
				track.Genres = genresMap[track.ID]
			}
		} else {
			r.log.Warn("failed to load genres while getting tracks by ids", "error", err)
		}
	}

	return tracks, nil
}

func (r *repository) buildGetTracksByIDsQuery(opts *models.GetTrackOptions) string {
	selectParts := []string{
		"t.id",
		"t.title",
		"t.duration",
		"t.year",
		"t.file_id",
		"t.cover_image_id",
		"t.track_number",
		"t.plays_count",
		"t.artist_id",
		"t.album_id",
		"t.created_at",
		"t.updated_at",
	}

	fromPart := "FROM tracks t"

	if opts.IncludeArtist {
		selectParts = append(selectParts,
			"a.id as \"artist.id\"",
			"a.name as \"artist.name\"",
			"a.country as \"artist.country\"",
			"a.avatar_image_id as \"artist.avatar_image_id\"",
		)
		fromPart += " JOIN artists a ON t.artist_id = a.id"
	}

	if opts.IncludeAlbum {
		selectParts = append(selectParts,
			"al.id as \"album_id\"",
			"al.title as \"album_title\"",
			"al.year as \"album_year\"",
			"al.cover_image_id as \"cover_image_id\"",
		)
		fromPart += " LEFT JOIN albums al ON t.album_id = al.id"
	}

	return fmt.Sprintf(`
		SELECT %s
		%s
		WHERE t.id = ANY($1)
		ORDER BY t.title ASC
	`, strings.Join(selectParts, ", "), fromPart)
}

func (r *repository) getTracksGenresBatch(ctx context.Context, trackIDs []string) (map[string][]*models.Genre, error) {
	query := `
		SELECT
			tg.track_id,
			g.id,
			g.name,
			g.description,
			g.created_at,
			g.updated_at
		FROM
			track_genres tg
		JOIN
			genres g ON tg.genre_id = g.id
		WHERE
			tg.track_id = ANY($1)
		ORDER BY
			g.name
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(trackIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]*models.Genre)
	for rows.Next() {
		var trackID string
		var genre models.Genre

		err := rows.Scan(
			&trackID,
			&genre.ID, genre.Name, genre.Description,
			&genre.CreatedAt, &genre.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		result[trackID] = append(result[trackID], &genre)
	}

	return result, nil
}

func (r *repository) UpdateTrack(ctx context.Context, id string, params *models.UpdateTrackParams) (*models.Track, error) {
	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{id}
	argIdx := 2

	if params.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *params.Title)
		argIdx++
	}

	if params.Duration != nil {
		setParts = append(setParts, fmt.Sprintf("duration = $%d", argIdx))
		args = append(args, *params.Duration)
		argIdx++
	}

	if params.Year != nil {
		setParts = append(setParts, fmt.Sprintf("year = $%d", argIdx))
		args = append(args, *params.Year)
		argIdx++
	}

	if params.FileID != nil {
		setParts = append(setParts, fmt.Sprintf("file_id = $%d", argIdx))
		args = append(args, *params.FileID)
		argIdx++
	}

	if params.CoverImageID != nil {
		setParts = append(setParts, fmt.Sprintf("cover_image_id = $%d", argIdx))
		args = append(args, *params.CoverImageID)
		argIdx++
	}

	if params.TrackNumber != nil {
		setParts = append(setParts, fmt.Sprintf("track_number = $%d", argIdx))
		args = append(args, *params.TrackNumber)
		argIdx++
	}

	if params.Lyrics != nil {
		setParts = append(setParts, fmt.Sprintf("lyrics = $%d", argIdx))
		args = append(args, *params.Lyrics)
		argIdx++
	}

	if params.ArtistID != nil {
		setParts = append(setParts, fmt.Sprintf("artist_id = $%d", argIdx))
		args = append(args, *params.ArtistID)
		argIdx++
	}

	if params.AlbumID != nil {
		setParts = append(setParts, fmt.Sprintf("album_id = $%d", argIdx))
		args = append(args, *params.AlbumID)
	}

	if len(setParts) == 1 {
		return r.GetTrackByID(ctx, id, nil)
	}

	query := fmt.Sprintf(`
		UPDATE
			tracks
		SET
			%s
		WHERE
			id = $1
		RETURNING
			id,
			title,
			duration,
			year,
			file_id,
			cover_image_id,
			track_number,
			lyrics,
			plays_count,
			created_at,
			updated_at
	`, strings.Join(setParts, ", "))

	var track models.Track
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&track.ID,
		&track.Title,
		&track.Duration,
		&track.Year,
		&track.FileID,
		&track.CoverImageID,
		&track.TrackNumber,
		&track.Lyrics,
		&track.PlaysCount,
		&track.CreatedAt,
		&track.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		r.log.Error("failed to update track", "error", err)
		return nil, fmt.Errorf("update track: %w", err)
	}

	if params.GenreIDs != nil {
		if err := r.setTrackGenres(ctx, id, *params.GenreIDs); err != nil {
			r.log.Error("failed to update track genres", "error", err)
		}
	}

	return r.GetTrackByID(ctx, id, nil)
}

func (r *repository) DeleteTrackByID(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tracks WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	// from track_genres table deleting will be by postgres by cascade

	return nil
}

func (r *repository) ListTracks(ctx context.Context, filter *models.TrackFilter) (*models.TrackListResult, error) {
	if filter == nil {
		filter = &models.TrackFilter{}
	}

	// Вынести в service!!!
	if filter.Page <= 0 {
		filter.Page = 1
	}

	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	//!!!

	wherePart := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if filter.ArtistID != "" {
		wherePart = append(wherePart, fmt.Sprintf("t.artist_id = $%d", argIdx))
		args = append(args, filter.ArtistID)
		argIdx++
	}

	if filter.AlbumID != "" {
		wherePart = append(wherePart, fmt.Sprintf("t.album_id = $%d", argIdx))
		args = append(args, filter.AlbumID)
		argIdx++
	}

	if len(filter.GenreIDs) > 0 {
		wherePart = append(wherePart, fmt.Sprintf(`
			EXISTS (
				SELECT 1 FROM track_genres tg
				WHERE tg.track_id = t.id AND tg.genre_id = ANY($%d)
			)
		`, argIdx))
		args = append(args, pq.Array(filter.GenreIDs))
		argIdx++
	}

	if filter.YearFrom > 0 {
		wherePart = append(wherePart, fmt.Sprintf("t.year >= $%d", argIdx))
		args = append(args, filter.YearFrom)
		argIdx++
	}

	if filter.YearTo > 0 {
		wherePart = append(wherePart, fmt.Sprintf("t.year <= $%d", argIdx))
		args = append(args, filter.YearTo)
		argIdx++
	}

	if filter.MinDuration > 0 {
		wherePart = append(wherePart, fmt.Sprintf("t.duration >= $%d", argIdx))
		args = append(args, filter.MinDuration)
		argIdx++
	}

	if filter.MaxDuration > 0 {
		wherePart = append(wherePart, fmt.Sprintf("t.duration <= $%d", argIdx))
		args = append(args, filter.MaxDuration)
		argIdx++
	}

	orderBy := "t.created_at DESC"
	switch filter.SortBy {
	case models.SortByTitle:
		orderBy = "t.title"
	case models.SortByArtist:
		orderBy = "a.name"
	case models.SortByYear:
		orderBy = "t.year"
	case models.SortByPlays:
		orderBy = "t.plays_count"
	}

	if filter.SortOrder != "" {
		if filter.SortOrder == models.SortOrderDesc {
			orderBy += " DESC"
		} else {
			orderBy += " ASC"
		}
	}

	query := fmt.Sprintf(`
		SELECT
			t.id,
			t.title,
			t.duration,
			t.year,
			t.file_id, 
			t.cover_image_id,
			t.track_number,
			t.lyrics,
			t.plays_count,
			t.created_at,
			t.updated_at,
			a.id as "artist.id",
			a.name as "artist.name",
			al.id as "album.id",
			al.title as "album.title",
			al.year as "album.year"
		FROM
			tracks t
		JOIN
			artists a ON t.artist_id = a.id
		LEFT JOIN
			albums al ON t.album_id = al.id
		WHERE
			%s
		ORDER BY
			%s
		LIMIT
			$%d
		OFFSET
			$%d
	`, strings.Join(wherePart, " AND "), orderBy, argIdx, argIdx+1)

	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)

	var tracks []*models.Track
	err := r.db.SelectContext(ctx, &tracks, query, args...)
	if err != nil {
		r.log.Error("failed to list tracks", "error", err)
		return nil, fmt.Errorf("list tracks: %w", err)
	}

	return &models.TrackListResult{
		Tracks:   tracks,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (r *repository) IncrementPlays(ctx context.Context, trackID string, incrementBy int64) error {
	query := `
		UPDATE
			tracks
		SET
			plays_count = plays_count + $1,
			updated_at = NOW()
		WHERE
			id = $2
	`

	result, err := r.db.ExecContext(ctx, query, incrementBy, trackID)
	if err != nil {
		r.log.Error("failed to increment plays", "error", err)
		return fmt.Errorf("increment plays: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *repository) SearchTracks(ctx context.Context, query string, opts *models.SearchOptions) ([]*models.Track, error) {
	if opts == nil {
		opts = &models.SearchOptions{}
	}

	safeQuery := sanitizeTSQuery(query)
	if safeQuery != "" {
		tracks, err := r.searchTracksFullText(ctx, safeQuery, opts)
		if err != nil && len(tracks) >= 1 {
			return tracks, nil
		}
	}

	tracks, err := r.searchTracksFuzzy(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func (r *repository) searchTracksFullText(ctx context.Context, safeQuery string, opts *models.SearchOptions) ([]*models.Track, error) {
	sql := r.buildSearchTracksQuery(true, opts)

	var tracks []*models.Track
	err := r.db.SelectContext(ctx, &tracks, sql, safeQuery, opts.Limit)
	return tracks, err
}

func (r *repository) searchTracksFuzzy(ctx context.Context, query string, opts *models.SearchOptions) ([]*models.Track, error) {
	sql := r.buildSearchTracksQuery(false, opts)
	
	var tracks []*models.Track
	err := r.db.SelectContext(ctx, &tracks, sql, query, opts.Limit)
	return tracks, err
}

func (r *repository) buildSearchTracksQuery(isFullText bool, opts *models.SearchOptions) string {
	selectPart := []string{
		"t.id",
		"t.title",
		"t.duration",
		"t.year",
		"t.file_id",
		"t.cover_image_id",
		"t.track_number",
		"t.lyrics",
		"t.plays_count",
		"t.created_at",
		"t.updated_at",
		"t.artist_id",
		"t.album_id",
	}

	fromPart := "FROM tracks t"
	wherePart := ""
	orderByPart := ""

	if opts.IncludeArtist {
		selectPart = append(selectPart, 
			"a.id as \"artist.id\"",
            "a.name as \"artist.name\"",
            "a.country as \"artist.country\"",
            "a.avatar_image_id as \"artist.avatar_image_id\"",
            "a.total_plays as \"artist.total_plays\"",
		)
		fromPart += " JOIN artists a ON t.artist_id = a.id"
	}

	if opts.IncludeAlbum {
		selectPart = append(selectPart, 
			"al.id as \"album.id\"",
            "al.title as \"album.title\"",
            "al.year as \"album.year\"",
            "al.cover_image_id as \"album.cover_image_id\"",
		)
		fromPart += " LEFT JOIN albums al ON t.album_id = al.id"
	}

	if isFullText {
		wherePart = "WHERE t.search_vector @@ to_tsquery('simple', $1)"
		selectPart = append(selectPart, "ts_rank(t.search_vector, to_tsquery('simple', $1)) as rank")
		orderByPart = "ORDER BY rank DESC"
	} else {
		wherePart = "WHERE t.title % $1"
		selectPart = append(selectPart, "similarity(t.title, $1) as sim")
		orderByPart = "ORDER BY sim DESC"
	}

	return fmt.Sprintf(`
		SELECT %s
		%s
		%s
		%s
		LIMIT $2
	`, strings.Join(selectPart, ", "), fromPart, wherePart, orderByPart)
}

// Genre Methods

func (r *repository) addTrackGenres(ctx context.Context, trackID string, genreIDs []string) error {
	return r.addGenres(ctx, "track_genres", "track_id", trackID, genreIDs)
}

func (r *repository) getTrackGenres(ctx context.Context, trackID string) ([]*models.Genre, error) {
	return r.GetGenres(ctx, "track_genres", "track_id", trackID)
}

func (r *repository) setTrackGenres(ctx context.Context, trackID string, genresIDs []string) error {
	return r.SetGenres(ctx, "track_genres", "track_id", trackID, genresIDs)
}

func (r *repository) addGenres(ctx context.Context, table, column, id string, genresIDs []string) error {
	if len(genresIDs) == 0 {
		return nil
	}

	if err := r.checkWhiteList(table, column); err != nil {
		return err
	}

	valueStrings := make([]string, 0, len(genresIDs))
	valueArgs := make([]interface{}, 0, len(genresIDs)*2)

	for i, genreID := range genresIDs {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, id, genreID)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (%s, GENRE_ID)
		VALUES %s
		ON CONFLICT DO NOTHING
	`, table, id, strings.Join(valueStrings, ","))

	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	return err
}

func (r *repository) GetGenres(ctx context.Context, table, column, id string) ([]*models.Genre, error) {
	if err := r.checkWhiteList(table, column); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT g.id, g.name, g.description, g.created_at, g.updated_at
		FROM genres g
		JOIN %s tg ON g.id = tg.genre_id
		WHERE tg.%s = $1
		ORDER BY g.name
	`, table, column)

	var genres []*models.Genre
	err := r.db.SelectContext(ctx, &genres, query, id)
	return genres, err
}

func (r *repository) SetGenres(ctx context.Context, table, column, id string, genresIDs []string) error {
	if err := r.checkWhiteList(table, column); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, column), id)
	if err != nil {
		return err
	}

	if len(genresIDs) > 0 {
		valueStrings := make([]string, 0, len(genresIDs))
		valueArgs := make([]interface{}, 0, len(genresIDs)*2)

		for i, genreID := range genresIDs {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
			valueArgs = append(valueArgs, id, genreID)
		}

		query := fmt.Sprintf("INSERT INTO %s (%s, genre_id) VALUES %s", table, column, strings.Join(valueStrings, ", "))
		_, err = tx.ExecContext(ctx, query, valueArgs...)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) checkWhiteList(table, column string) error {
	if _, ok := allowedTables[table]; !ok {
		return fmt.Errorf("invalid table name: %s", table)
	}

	if _, ok := allowedColumns[column]; !ok {
		return fmt.Errorf("invalid column name: %s", column)
	}

	return nil
}

// Help Functions

func sanitizeTSQuery(query string) string {
	specialChars := []string{"'", "\\", ":", "&", "|", "!", "(", ")", "<", ">", "*"}
	sanitized := query
	for _, char := range specialChars {
		sanitized = strings.ReplaceAll(sanitized, char, "")
	}

	words := strings.Fields(sanitized)
	if len(words) == 0 {
		return ""
	}

	for i, word := range words {
		words[i] = strings.ToLower(word)
	}

	return strings.Join(words, " & ")
}