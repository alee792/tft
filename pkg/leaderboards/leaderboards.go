// Package leaderboards groups and sorts players by...various...metrics.
package leaderboards

import (
	"context"
	"time"

	"github.com/alee792/teamfit/pkg/tft"
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
	// UpdateLeaderboard(ctx context.Context, id string, board *Leaderboard) (*Leaderboard, error)
	// DeleteLeaderboard(ctx context.Context, id string) (*Leaderboard, error)
}

// API for TFT
type API interface {
	GetSummoner(ctx context.Context, name string) (*tft.Summoner, error)
	GetSummoners(ctx context.Context, names []string) ([]tft.Summoner, error)
	ListMatches(ctx context.Context, in *tft.ListMatchesRequest) (*tft.ListMatchesResponse, error)
	GetMatch(ctx context.Context, in *tft.GetMatchRequest) (*tft.GetMatchResponse, error)
	GetMostRecentMatch(ctx context.Context, summoner string) (*tft.Match, error)
	GetLeagueEntry(ctx context.Context, summonerID string) (*tft.LeagueEntry, error)
}

// Leaderboard is a statless group of Summoners.
type Leaderboard struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Summoners map[string]Summoner // Key = Summoner.Name `json:"summoners"`
}

// SummonerID ties a PUUID to a Name.
type SummonerID struct {
	PUUID string
	Name  string
}

// UnixMS parses a MS formatted epoch timestamp for a Go time.Time.
func UnixMS(ms int) time.Time {
	return time.Unix(int64(ms/1000), 0)
}
