package db

import (
	"context"
	"database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"magtapp-backend/ent"
)

// NewEntClient opens a Postgres connection (via pgx) and wraps it as an
// Ent client, running schema migration on startup. Fine for this
// assignment's scope; a real production rollout would run migrations as
// a separate, reviewed step rather than on every boot.
func NewEntClient(dsn string) (*ent.Client, error) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	drv := entsql.OpenDB(dialect.Postgres, sqlDB)
	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, err
	}
	return client, nil
}
