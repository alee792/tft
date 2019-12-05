package rest

import (
	"net/http"

	"github.com/go-chi/cors"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Route sets a server's routes.
func (s *Server) Route() {
	s.Router.Use(middleware.Logger)

	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	s.Router.Use(cors.Handler)

	s.Router.Route("/summoners", func(r chi.Router) {
		r.Route("/{name}", func(r chi.Router) {
			r.Use((pathToQuery("name", "name")))

			r.Get("/", s.GetSummonerHandler())
			r.Get("/stats", s.GetStatsByNameHandler())
			r.Get("/results", s.GetResultsByNameHandler())
		})
	})

	s.Router.Route("/stats", func(r chi.Router) {
		r.Get("/", s.GetStatsByNameHandler())
	})

	s.Router.Route("/boards", func(r chi.Router) {
		r.Route("/{name}", func(r chi.Router) {
			r.Use((pathToQuery("name", "name")))
			r.Get("/", s.GetLeaderboardByNameHandler())
		})

		r.Post("/", s.CreateLeaderboardHandler())
	})
}

func pathToQuery(pathParam, queryParam string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			param := chi.URLParam(r, pathParam)
			q := r.URL.Query()
			q.Add(queryParam, param)

			r.URL.RawQuery = q.Encode()

			next.ServeHTTP(w, r)
		})
	}
}
