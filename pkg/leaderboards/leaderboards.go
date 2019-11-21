// Package leaderboards groups and sorts players by...various...metrics.
package leaderboards

import (
	"context"
	"time"

	"github.com/alee792/teamfit/pkg/tft"
	"github.com/pkg/errors"
)

// Server of leaderboards and related stats.
type Server struct {
	API     API
	Storage Storage
}

// Storage persists Leaderboards.
type Storage interface {
	CreateLeaderboard(ctx context.Context, board *Leaderboard) (*Leaderboard, error)
	GetLeaderboard(ctx context.Context, id string) (*Leaderboard, error)
	UpdateLeaderboard(ctx context.Context, id string, board *Leaderboard) (*Leaderboard, error)
	DeleteLeaderboard(ctx context.Context, id string) (*Leaderboard, error)
}

// API for TFT
type API interface {
	GetSummoner(ctx context.Context, in *tft.GetSummonerRequest) (*tft.GetSummonerResponse, error)
	GetSummoners(ctx context.Context, names []string) ([]tft.Summoner, error)
	ListMatches(ctx context.Context, in *tft.ListMatchesRequest) (*tft.ListMatchesResponse, error)
	GetMatch(ctx context.Context, in *tft.GetMatchRequest) (*tft.GetMatchResponse, error)
	GetMostRecentMatch(ctx context.Context, summoner string) (*tft.Match, error)
}

// Leaderboard is a statless group of Summoners.
type Leaderboard struct {
	ID        string
	Name      string
	Summoners map[string]tft.Summoner // Key = Summoner.Name
}

// SummonerID ties a PUUID to a Name.
type SummonerID struct {
	PUUID string
	Name  string
}

// Result for a single player's match.
type Result struct {
	MatchID   string    `json:"match_id"`
	StartedAt time.Time `json:"started_at"`
	tft.Participant
}

// Stats are aggregations of game results.
type Stats struct {
	Games             int
	DamageDealt       int
	PlayersEliminated int
	BoardValue        int
	Wins              int
	TopFours          int
	AverageFinish     float32
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

func getBoardValue(uu []tft.Unit) int {
	var val int
	for _, u := range uu {
		val += u.Tier * u.Rarity
	}
	return val
}

// GetStatsRequest is a request for aggregated results.
type GetStatsRequest struct {
	GetResultsArgs
}

// GetStatsResponse returns aggregated results.
type GetStatsResponse struct {
	GameLimit   int
	Before      time.Time
	After       time.Time
	Leaderboard map[string]Stats // Key = Summoner.Name
}

func calculateStats(results []tft.Participant) *Stats {

	// stat.DamageDealt += p.TotalDamageToPlayers
	// stat.PlayersEliminated += p.PlayersEliminated
	// stat.BoardValue += getBoardValue(p.Units)
	// // Must divide by total games before returning!
	// stat.AverageFinish += float32(p.Placement)

	// switch place := p.Placement; {
	// case place == 1:
	// 	stat.TopFours++
	// 	stat.Wins++
	// case place > 5:
	// 	stat.TopFours++
	// }
	return nil
}

func UnixMS(ms int) time.Time {
	return time.Unix(int64(ms/1000), 0)
}
