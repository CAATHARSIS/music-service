package service

import (
	"time"

	playlistpb "github.com/CAATHARSIS/music-service/api/gen/playlist"
	"github.com/CAATHARSIS/music-service/internal/playlist/models"
)

func convertPlaylistToProto(p *models.Playlist) *playlistpb.Playlist {
    pb := &playlistpb.Playlist{
        Id:          p.ID,
        UserId:      p.UserID,
        Name:        p.Name,
        Description: p.Description,
        IsPublic:    p.IsPublic,
        CreatedAt:   p.CreatedAt.Format(time.RFC3339),
        UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
    }

    switch p.Type {
    case models.PlaylistTypeGenerated:
        pb.Type = playlistpb.PlaylistType_PLAYLIST_TYPE_GENERATED
    case models.PlaylistTypeFavoriets:
        pb.Type = playlistpb.PlaylistType_PLAYLIST_TYPE_FAVORITES
    case models.PlaylistTypeManual:
        pb.Type = playlistpb.PlaylistType_PLAYLIST_TYPE_MANUAL
	default:
		pb.Type = playlistpb.PlaylistType_PLAYLIST_TYPE_UNSPECIFIED
    }

    return pb
}

func convertPlaylistTrackToProto(t models.PlaylistTrack) *playlistpb.PlaylistTrack {
    return &playlistpb.PlaylistTrack{
        TrackId:  t.TrackID,
        Position: int32(t.Position),
        AddedAt:  t.AddedAt.Format(time.RFC3339),
    }
}