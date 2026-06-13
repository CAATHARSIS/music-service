package models

import "time"

type Track struct {
	ID           string    `db:"id" json:"id"`
	Title        string    `db:"title" json:"title"`
	Duration     int       `db:"duration" json:"duration"`
	Year         int       `db:"year" json:"year"`
	FileID       string    `db:"file_id" json:"file_id"`
	CoverImageID *string   `db:"cover_image_id" json:"cover_image_id,omitempty"`
	TrackNumber  *int      `db:"track_number" json:"track_number,omitempty"`
	Lyrics       *string   `db:"lyrics" json:"lyrics,omitempty"`
	PlaysCount   int       `db:"plays_count" json:"plays_count"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`

	Artists []*Artist `json:"artist"`
	Genres  []*Genre  `json:"genres,omitempty"`

	AlbumID        *string    `db:"album_id" json:"-"`
	AlbumTitle     *string    `db:"album_title" json:"-"`
	AlbumYear      *int32     `db:"album_year" json:"-"`
	AlbumCoverID   *string    `db:"album_cover_image_id" json:"-"`
	AlbumType      *AlbumType `db:"album_type" json:"-"`
	AlbumCreatedAt *time.Time `db:"album_created_at" json:"-"`
	AlbumUpdatedAt *time.Time `db:"album_updated_at" json:"-"`
}

type Artist struct {
	ID            string    `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Country       *string   `db:"country" json:"country,omitempty"`
	AvatarImageID *string   `db:"avatar_image_id" json:"avatar_image_id,omitempty"`
	TotalPlays    int64     `db:"total_plays" json:"total_plays"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`

	Genres []*Genre `json:"genres,omitempty"`
}

type Album struct {
	ID           string    `db:"id" json:"id"`
	Title        string    `db:"title" json:"title"`
	Year         int32     `db:"year" json:"year"`
	AlbumType    AlbumType `db:"album_type" json:"album_type"`
	CoverImageID *string   `db:"cover_image_id" json:"cover_image_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`

	Artists []*Artist `json:"artist,omitempty"`
	Genres  []*Genre  `json:"genres,omitempty"`
}

type AlbumWithTracks struct {
	Album  *Album
	Tracks []*Track `json:"tracks"`
	Genres []*Genre `json:"genres"`
}

type Genre struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type CreateTrackParams struct {
	Title        string
	Duration     int
	Year         int
	ArtistIDs    []string
	AlbumID      *string
	GenreIDs     []string
	FileID       string
	CoverImageID *string
	TrackNumber  *int32
	Lyrics       *string
}

type GetTrackOptions struct {
	IncludeArtists bool
	IncludeAlbum   bool
	IncludeGenres  bool
}

type UpdateTrackParams struct {
	Title        *string
	Duration     *int32
	Year         *int32
	FileID       *string
	CoverImageID *string
	TrackNumber  *int32
	Lyrics       *string
	ArtistIDs    *[]string
	AlbumID      *string
	GenreIDs     *[]string
}

type CreateArtistParams struct {
	Name          string
	Country       *string
	AvatarImageID *string
	GenreIDs      []string
}

type UpdateArtistParams struct {
	Name          *string
	Country       *string
	AvatarImageID *string
	GenreIDs      *[]string
}

type SearchTracksOptions struct {
	Limit         int
	IncludeArtist bool
	IncludeAlbum  bool
}

type CreateAlbumParams struct {
	Title        string
	Year         int
	ArtistIDs    []string
	CoverImageID *string
	AlbumType    AlbumType
	GenresIDs    []string
}

type UpdateAlbumParams struct {
	Title        *string
	Year         *int32
	ArtistIDs    []string
	CoverImageID *string
	AlbumType    *AlbumType
	GenreIDs     []string
}

type SearchAlbumsOptions struct {
	Limit         int
	IncludeArtist bool
	IncludeTracks bool
}

type CreateGenreParams struct {
	Name        string
	Description *string
}

// Filters

type TrackFilter struct {
	ArtistID      string
	AlbumID       string
	GenreIDs      []string
	YearFrom      int
	YearTo        int
	MinDuration   int
	MaxDuration   int
	MinPlaysCount int
	MaxPlaysCount int
	SortBy        TrackSortBy
	SortOrder     SortOrder
	Page          int
	PageSize      int
}

type ArtistFilter struct {
	GenreIDs  []string
	Country   string
	MinPlays  int64
	SortBy    ArtistSortBy
	SortOrder SortOrder
	Page      int
	PageSize  int
}

type AlbumFilter struct {
	ArtistIDs []string
	GenreIDs  []string
	YearFrom  int
	YearTo    int
	AlbumType AlbumType
	SortBy    AlbumSortBy
	SortOrder SortOrder
	Page      int
	PageSize  int
}

// Pagination Result

type TrackListResult struct {
	Tracks   []*Track `json:"tracks"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}

type ArtistListResult struct {
	Artists  []*Artist `json:"artists"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

type AlbumListResult struct {
	Albums   []*Album `json:"albums"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}

// Additional Types

type TrackSortBy string

const (
	SortByTitle  TrackSortBy = "title"
	SortByArtist TrackSortBy = "artist"
	SortByYear   TrackSortBy = "year"
	SortByPlays  TrackSortBy = "plays"
)

type ArtistSortBy string

const (
	SortArtistsByName  ArtistSortBy = "name"
	SortArtistsByPlays ArtistSortBy = "plays"
)

type AlbumSortBy string

const (
	SortAlbumByTitle  AlbumSortBy = "title"
	SortAlbumByYear   AlbumSortBy = "year"
	SortAlbumByArtist AlbumSortBy = "artist"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type AlbumType string

const (
	AlbumTypeUnspecified AlbumType = "unspecified"
	AlbumTypeAlbum       AlbumType = "album"
	AlbumTypeEP          AlbumType = "EP"
	AlbumTypeSingle      AlbumType = "single"
)
