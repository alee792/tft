package leaderboards

import (
	"context"
	"time"

	"github.com/alee792/teamfit/pkg/tft"
	"github.com/pkg/errors"
)

// Result for a single player's match.
type Result struct {
	MatchID   string    `json:"match_id"`
	StartedAt time.Time `json:"started_at"`
	tft.Participant
}

// GetResultsArgs allows users to query match results.
// Must be paired with additional identifiers, either a Leaderboard or list of Summoners.
type GetResultsArgs struct {
	GameLimit int
	Before    time.Time
	After     time.Time
}

// PUUIDResults returns Summoner's match results with PUUIDs as keys.
type PUUIDResults map[string][]Result

// NameResults returns Summoner's match results with names as keys.
type NameResults map[string][]Result

// GetResults for a set of Summoners.
func (s *Server) GetResults(ctx context.Context, puuids []string, in *GetResultsArgs) (PUUIDResults, error) {
	if in.GameLimit < 1 {
		in.GameLimit = 1
	}

	// Collate match IDs to query.
	var matches = make(map[string]*tft.Match)
	var matchIDs []string

	for _, id := range puuids {
		out, err := s.API.ListMatches(ctx, &tft.ListMatchesRequest{
			PUUID: id,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list matches for %s", id)
		}

		matchIDs = out.MatchIDs
		for _, id := range matchIDs {
			matches[id] = nil
		}
	}

	// Prepare results and Leaderboard.
	// Participants are really match results for participants.
	var results PUUIDResults = make(map[string][]Result) // Key = Summoner.PUUID
	for _, id := range puuids {
		results[id] = []Result{}
	}

	// Retrieve matches and collate results.
	for _, id := range matchIDs {
		out, err := s.API.GetMatch(ctx, &tft.GetMatchRequest{
			MatchID: id,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to retrieve match %s", id)
		}

		// Enforce time range.
		gameStart := UnixMS(out.Match.Info.GameTimestamp)
		if (!in.Before.IsZero() && gameStart.After(in.Before)) || (!in.After.IsZero() && gameStart.Before(in.After)) {
			continue
		}

		matches[id] = &out.Match

		// Append match results if player is tracked on Leaderboard.
		for _, p := range out.Match.Info.Participants {
			_, ok := results[p.PUUID]
			if !ok {
				continue
			}

			// Do not append results if a player exceeds the match limit.
			// The assumption is that listed match IDs are returned in chronological order.
			if len(results[p.PUUID]) >= in.GameLimit {
				break
			}

			results[p.PUUID] = append(results[p.PUUID], Result{
				MatchID:     id,
				StartedAt:   gameStart,
				Participant: p,
			})
		}
	}

	return results, nil
}

// GetResultsFromNames retrives results of Summoners attached to a Leaderboard.
func (s *Server) GetResultsFromNames(ctx context.Context, names []string, in *GetResultsArgs) (NameResults, error) {
	smnrs, err := s.API.GetSummoners(ctx, names)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve summoner PUUIDs")
	}

	var nameResults NameResults = make(map[string][]Result)
	var ids []string
	for _, smnr := range smnrs {
		nameResults[smnr.Name] = nil
		ids = append(ids, smnr.PUUID)
	}

	out, err := s.GetResults(ctx, ids, in)
	if err != nil {
		return nil, err
	}

	for _, smnr := range smnrs {
		nameResults[smnr.Name] = out[smnr.PUUID]
	}

	return nameResults, nil
}

// GetResultsFromLeaderboard retrieves results of Summoners attached to a Leaderboard.
func (s *Server) GetResultsFromLeaderboard(ctx context.Context, id string, in *GetResultsArgs) (NameResults, error) {
	board, err := s.Storage.GetLeaderboard(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "get leaderboard failed")
	}

	var names []string
	for _, smnr := range board.Summoners {
		names = append(names, smnr.Name)
	}

	return s.GetResultsFromNames(ctx, names, in)
}
