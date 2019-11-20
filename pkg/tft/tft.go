package tft

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
)

type Client struct {
	client *http.Client
	Config Config
}

type Config struct {
	APIKey string
}

func NewClient(client *http.Client, cfg Config) *Client {
	return &Client{
		client: client,
		Config: cfg,
	}
}

type GetSummonerRequest struct {
	SummonerName string
}

type GetSummonerResponse struct {
	Summoner Summoner
}

func (c *Client) GetSummoner(ctx context.Context, in *GetSummonerRequest) (*GetSummonerResponse, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://na1.api.riotgames.com/tft/summoner/v1/summoners/by-name/", nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("X-Riot-Token", c.Config.APIKey)
	r.URL.Path = path.Join(r.URL.Path, in.SummonerName)

	resp, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}

	body := resp.Body
	defer resp.Body.Close()

	var s Summoner
	if err := json.NewDecoder(body).Decode(&s); err != nil {
		return nil, err
	}

	return &GetSummonerResponse{
		Summoner: s,
	}, nil
}

type ListMatchesRequest struct {
	PUUID string
}

type ListMatchesResponse struct {
	MatchIDs []string
}

func (c *Client) ListMatches(ctx context.Context, in *ListMatchesRequest) (*ListMatchesResponse, error) {
	if in.PUUID == "" {
		return nil, fmt.Errorf("invalid PUUID")
	}

	path := fmt.Sprintf("https://americas.api.riotgames.com/tft/match/v1/matches/by-puuid/%s/ids", in.PUUID)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("X-Riot-Token", c.Config.APIKey)

	resp, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}

	body := resp.Body
	defer resp.Body.Close()

	var matches []string
	if err := json.NewDecoder(body).Decode(&matches); err != nil {
		return nil, err
	}

	return &ListMatchesResponse{
		MatchIDs: matches,
	}, nil
}

type GetMatchRequest struct {
	MatchID string
}

type GetMatchResponse struct {
	Match Match
}

func (c *Client) GetMatch(ctx context.Context, in *GetMatchRequest) (*GetMatchResponse, error) {
	if in.MatchID == "" {
		return nil, fmt.Errorf("invalid matchID")
	}

	path := fmt.Sprintf("https://americas.api.riotgames.com/tft/match/v1/matches/%s", in.MatchID)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("X-Riot-Token", c.Config.APIKey)

	resp, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}

	body := resp.Body
	defer resp.Body.Close()

	var m Match
	if err := json.NewDecoder(body).Decode(&m); err != nil {
		return nil, err
	}

	return &GetMatchResponse{
		Match: m,
	}, nil
}

func (c *Client) GetMostRecentMatch(ctx context.Context, summoner string) (*Match, error) {
	sOut, err := c.GetSummoner(ctx, &GetSummonerRequest{
		SummonerName: summoner,
	})
	if err != nil {
		return nil, err
	}

	mmOut, err := c.ListMatches(ctx, &ListMatchesRequest{
		PUUID: sOut.Summoner.PUUID,
	})
	if err != nil {
		return nil, err
	}

	mOut, err := c.GetMatch(ctx, &GetMatchRequest{
		MatchID: mmOut.MatchIDs[0],
	})
	if err != nil {
		return nil, err
	}

	return &mOut.Match, nil
}

type Summoner struct {
	ProileIconID  int
	Name          string
	PUUID         string
	SummonerLevel int
	AccountID     string
	ID            string
	RevisionDate  int
}

type Match struct {
	Info     Info     `json:"info"`
	Metadata Metadata `json:"metadata"`
}

type Info struct {
	GameTimestamp int           `json:"game_datetime"`
	Participants  []Participant `json:"participants"`
	Set           int           `json:"tft_set_number"`
	GameLength    float32       `json:"game_length"`
	QueueID       int           `json:"queue_id"`
	GameVersion   string        `json:"game_version"`
}

type Participant struct {
	Placement            int       `json:"placement"`
	Level                int       `json:"level"`
	LastRound            int       `json:"last_round"`
	TimeEliminated       float32   `json:"time_eliminated"`
	Companion            Companion `json:"companion"`
	Traits               []Trait   `json:"traits"`
	PlayersEliminated    int       `json:"players_eliminated"`
	PUUID                string    `json:"puuid"`
	TotalDamageToPlayers int       `json:"total_damage_to_players"`
	Units                []Unit    `json:"units"`
}

type Companion struct {
	SkinID    int    `json:"skin_id"`
	ContentID string `json:"content_id"`
	Species   string `json:"species"`
}

type Trait struct {
	TierTotal   int    `json:"tier_total"`
	Name        string `json:"name"`
	TierCurrent int    `json:"tier_current"`
	NumUnits    int    `json:"num_units"`
}

type Unit struct {
	Tier        int    `json:"tier"`
	Items       []int  `json:"items"`
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	Rarity      int    `json:"rarity"`
}

type Metadata struct {
	DataVersion  string   `json:"data_version"`
	Participants []string `json:"participants"`
	MatchID      string   `json:"match_id"`
}
