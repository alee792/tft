package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alee792/teamfit/pkg/leaderboards"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type Server struct {
	Boarder *leaderboards.Server
	Router  chi.Router
	Logger  *zap.SugaredLogger
	Config  Config
}

type Config struct {
	Addr string
}

func NewServer(cfg Config, opts ...Option) (*Server, error) {
	s := &Server{
		Config: cfg,
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	// Set sensible defaults.

	if s.Router == nil {
		s.Router = chi.NewRouter()
	}

	if s.Logger == nil {
		s.Logger = zap.NewExample().Sugar()
	}

	s.Route()

	return s, nil
}

func (s *Server) GetStatsByNameHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		names, ok := q["name"]
		if !ok || len(names[0]) < 1 {
			http.Error(w, "missing name", http.StatusBadRequest)
			return
		}

		rawLimit := q.Get("matches")
		matches, _ := strconv.Atoi(rawLimit)
		if matches < 1 {
			matches = 10
		}

		// App logic.
		ctx := context.Background()
		out, err := s.Boarder.GetStats(ctx, names, &leaderboards.GetStatsArgs{
			GameLimit: matches,
		})
		if err != nil {
			return
		}

		// Respond.
		if err := s.respondJSON(w, out, http.StatusOK); err != nil {
			s.Logger.Warnw("json encoding failed", "err", err)
		}
	}
}

func (s *Server) GetResultsByNameHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		names, ok := q["name"]
		if !ok || len(names[0]) < 1 {
			http.Error(w, "missing name", http.StatusBadRequest)
			return
		}

		rawLimit := q.Get("matches")
		matches, _ := strconv.Atoi(rawLimit)
		if matches < 1 {
			matches = 10
		}

		// App logic.
		ctx := context.Background()
		out, err := s.Boarder.GetResultsFromNames(ctx, names, &leaderboards.GetResultsArgs{
			GameLimit: matches,
		})
		if err != nil {
			return
		}

		// Respond.
		if err := s.respondJSON(w, out, http.StatusOK); err != nil {
			s.Logger.Warnw("json encoding failed", "err", err)
		}
	}
}

func (s *Server) GetSummonerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		names, ok := q["name"]
		if !ok || len(names[0]) < 1 {
			http.Error(w, "missing name", http.StatusBadRequest)
			return
		}

		// App logic.
		ctx := context.Background()
		out, err := s.Boarder.GetSummoner(ctx, names[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println(out.LeagueEntry)

		// Respond.
		if err := s.respondJSON(w, out, http.StatusOK); err != nil {
			s.Logger.Warnw("json encoding failed", "err", err)
		}
	}
}

func (s *Server) respondJSON(w http.ResponseWriter, v interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(&v); err != nil {
		return err
	}

	return nil
}
