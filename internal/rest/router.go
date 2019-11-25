package rest

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Route sets a server's routes.
func (s *Server) Route() {
	s.Router.Use(middleware.Logger)

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
