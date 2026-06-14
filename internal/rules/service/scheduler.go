package service

import (
	"context"
	"time"

	playlistpb "github.com/CAATHARSIS/music-service/api/gen/playlist"
)

func (s *RuleEngineService) StartScheduler(ctx context.Context) {
	s.log.Info("scheduler started, will run daily at midnight UTC")

	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
		duration := next.Sub(now)

		s.log.Info("next scheduled run", "in", duration.Round(time.Minute))

		select {
		case <-time.After(duration):
			s.regenerateAllActiveRules(ctx)
		case <-ctx.Done():
			s.log.Info("scheduler stopped")
			return
		}
	}
}

func (s *RuleEngineService) regenerateAllActiveRules(ctx context.Context) {
	s.log.Info("starting daily playlist regeneration")

	rules, err := s.repo.GetActiveRules(ctx)
	if err != nil {
		s.log.Error("failed to get active rules", "error", err)
		return
	}

	for _, rule := range rules {
		s.log.Info("regenerating playlist for rule", "rule_id", rule.ID, "name", rule.Name)

		trackIDs, err := s.executeRule(ctx, rule)
		if err != nil {
			s.log.Error("rule execution failed", "rule_id", rule.ID, "error", err)
			continue
		}

		playlistID, err := s.repo.GetGeneratedPlaylistID(ctx, rule.ID)
        if err != nil || playlistID == "" {
            resp, err := s.playlistClient.CreatePlaylist(ctx, &playlistpb.CreatePlaylistRequest{
                UserId: rule.UserID,
                Name:   rule.Name,
            })
            if err != nil {
                continue
            }
            playlistID = resp.Id
            s.repo.SaveGeneratedPlaylist(ctx, rule.ID, playlistID)
        } else {
            s.playlistClient.ClearTracks(ctx, &playlistpb.ClearTracksRequest{
                PlaylistId: playlistID,
            })
        }

        for _, trackID := range trackIDs {
            if trackID != "" {
                s.playlistClient.AddTrack(ctx, &playlistpb.AddTrackRequest{
                    PlaylistId: playlistID,
                    UserId:     rule.UserID,
                    TrackId:    trackID,
                })
            }
        }

		s.repo.MarkExecuted(ctx, rule.ID)
		s.log.Info("playlist regenerated", "rule_id", rule.ID, "tracks", len(trackIDs))
	}

	s.log.Info("daily playlist regeneration completed", "rules_processed", len(rules))
}
