package resolver

import (
	pg "github.com/alee792/teamfit/pkg/storage/postgres"
	"github.com/jmoiron/sqlx" 
	// _ "github.com/lib/pq"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
)
 
type Resolver struct {
	postgres *postgres.Client
}

func (r *Resolver) ResolvePostgres() *postgres.Client {
	if r.postgres == nil {
		r.postgres = &postgres.Client{
			DB: sqlx.MustConnect("cloudsqlpostgres", "postgresql://postgres:1In0J9iFCABFag1e@35.238.9.164:5432/tft?sslMode=disable"),
		}
	}

	return r.postgres
}
