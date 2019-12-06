package jsonmap

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
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
	fileMux  *sync.Mutex
}

func NewClient(path string) (*Client, error) {
	c := &Client{
		Path:     path,
		Boards:   make(map[string]*leaderboards.Leaderboard),
		boardMux: &sync.Mutex{},
		fileMux:  &sync.Mutex{},
	}

	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 {
		return c, nil
	}

	if err := c.read(); err != nil {
		return nil, errors.Wrapf(err, "unable to read %s", path)
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

	bb, err := ioutil.ReadFile(c.Path)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(bytes.NewReader(bb)).Decode(&c.Boards); err != nil {
		return errors.Wrapf(err, "could not decode file at %s", c.Path)
	}

	return nil
}

func (c *Client) write() error {
	c.fileMux.Lock()
	defer c.fileMux.Unlock()

	f, err := os.OpenFile(c.Path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(&c.Boards); err != nil {
		return errors.Wrapf(err, "could not decode file at %s", c.Path)
	}

	return nil
}
