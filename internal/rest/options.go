package rest

import "github.com/alee792/teamfit/pkg/leaderboards"

type Option func(*Server) error

func WithBoarder(b *leaderboards.Server) Option {
	return func(s *Server) error {
		s.Boarder = b
		return nil
	}
}
