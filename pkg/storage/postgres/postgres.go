package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	boards "github.com/alee792/teamfit/pkg/leaderboards"
	"github.com/alee792/teamfit/pkg/tft"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Client struct {
	DB *sqlx.DB
}

func (c *Client) Initialize(ctx context.Context) error {
	_, err := c.DB.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS leaderboards (
		id text PRIMARY KEY,
		name text
	)
	`)
	if err != nil {
		return errors.Wrap(err, "failed to create leaderboards table")
	}

	_, err = c.DB.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS summoners (
		puuid text PRIMARY KEY,
		id text UNIQUE,
		name text,
		accountID text,
		profileIconID int,
		summonerLevel int,
		revisionDate timestamp
	)
	`)
	if err != nil {
		return errors.Wrap(err, "failed to create summoners table")
	}

	_, err = c.DB.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS leagueEntries (	
		summonerID text REFERENCES summoners(id),
		inactive bool,
		freshBlood bool,
		veteran bool,
		hotStreak bool,
		queueType text,
		summonerName text,
		wins int,
		losses int,
		rank text,
		leagueID text,
		tier text,
		progress text,
		target int
	)
	`)
	if err != nil {
		return errors.Wrap(err, "failed to create leagueEntries table")
	}

	_, err = c.DB.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS leaderboardsMembership (
		puuid text REFERENCES summoners(puuid),
		name text,
		boardID text REFERENCES leaderboards(id)
	)
	`)
	if err != nil {
		return errors.Wrap(err, "failed to create summoners table")
	}

	return nil
}

func (c *Client) CreateLeaderboard(ctx context.Context, board *boards.Leaderboard) (*boards.Leaderboard, error) {
	q := sq.Insert("leaderboards").
		Columns("id", "name").
		Values(board.ID, board.Name).
		RunWith(c.DB).
		PlaceholderFormat(sq.Dollar)

	var out boards.Leaderboard

	err := q.QueryRowContext(ctx).Scan(&out)
	if err != nil {
		return nil, err
	}

	for _, smnr := range board.Summoners {
		smnr := smnr
		if err := c.CreateLeaderboardsMembership(ctx, board.ID, &smnr); err != nil {
			return nil, errors.Wrap(err, "failed to create leaderboard membership")
		}
	}

	return &out, nil
}

func (c *Client) CreateLeaderboardsMembership(ctx context.Context, boardID string, smnr *tft.Summoner) error {
	q := sq.Insert("leaderboardsMembership").
		Columns("puiid", "name", "boardID").
		Values(smnr.PUUID, smnr.Name, boardID).
		RunWith(c.DB)

	if _, err := q.ExecContext(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetLeaderboard(ctx context.Context, id string) (*boards.Leaderboard, error) {
	q := sq.Select().
		From("leaderboards").
		Join("leaderboardsMembership ON leaderboards.ID = leaderboardsMembership.boardID").
		Where(sq.Eq{"boardID": id}).
		RunWith(c.DB)

	var out *boards.Leaderboard
	if err := q.QueryRowContext(ctx).Scan(&out); err != nil {
		return nil, err
	}

	return out, nil
}
