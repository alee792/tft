package jsonmap

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/alee792/teamfit/pkg/leaderboards"
	"github.com/pkg/errors"
)

var _ leaderboards.Storage = &Client{}

type Client struct {
	// Path to the JSON encoded leaderboard map.
	Path     string
	Boards   map[string]*leaderboards.Leaderboard
	boardMux *sync.Mutex
	enc      *json.Encoder
	dec      *json.Decoder
	file     *os.File
	fileMux  *sync.Mutex
}

func NewClient(path string) (*Client, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return nil, err
	}

	c := &Client{
		Path:     path,
		Boards:   make(map[string]*leaderboards.Leaderboard),
		boardMux: &sync.Mutex{},
		fileMux:  &sync.Mutex{},
		enc:      json.NewEncoder(f),
		dec:      json.NewDecoder(f),
		file:     f,
	}

	// Don't read if it's empty.
	stat, err := f.Stat()
	if err == nil {
		if stat.Size() == 0 {
			return c, nil
		}
	}

	if err := c.read(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) CreateLeaderboard(ctx context.Context, board *leaderboards.Leaderboard) (*leaderboards.Leaderboard, error) {
	c.boardMux.Lock()
	defer c.boardMux.Unlock()

	c.Boards[board.Name] = board

	if err := c.write(); err != nil {
		return nil, err
	}

	return c.Boards[board.Name], nil
}

func (c *Client) GetLeaderboard(ctx context.Context, id string) (*leaderboards.Leaderboard, error) {
	c.boardMux.Lock()
	defer c.boardMux.Unlock()

	return c.Boards[id], nil
}

func (c *Client) read() error {
	c.fileMux.Lock()
	defer c.fileMux.Unlock()

	if err := c.dec.Decode(&c.Boards); err != nil {
		return errors.Wrapf(err, "could not decode file at %s", c.Path)
	}

	return nil
}

func (c *Client) write() error {
	c.fileMux.Lock()
	defer c.fileMux.Unlock()

	if err := c.enc.Encode(&c.Boards); err != nil {
		return errors.Wrapf(err, "could not decode file at %s", c.Path)
	}

	return nil
}
