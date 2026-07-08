// Package migrations embeds the SQL migration files so they ship inside the
// compiled binary (TDR-003) — no separate migration step or volume needed to
// run `docker compose up`.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
