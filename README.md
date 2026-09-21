# MagTapp Pro — Backend Prototype

A minimal user + subscription service for the MagTapp founding engineer assignment.

**Stack**: Go, Echo, Ent (Postgres ORM), Redis, `oaswrap/spec` (OpenAPI docs), Air (hot reload).

## What it does

- `Authorization: Bearer <token>` middleware **creates or identifies** a user from that token (no separate signup step — this is the "basic auth" the brief asks for).
- `GET /v1/subscription` — returns the caller's subscription status. Redis is checked first; Postgres (via Ent) is the source of truth and repopulates the cache on a miss.
- `PUT /v1/subscription` — updates subscription status. Accepts an optional `Idempotency-Key` header so a retried call (mirroring a retried Apple/Google webhook in the real design) doesn't double-apply.
- `GET /v1/me` — returns the identified user.
- Full OpenAPI docs auto-generated and served at `GET /docs`.

## Why these pieces

- **Ent** over raw SQL/GORM: schema-first, generates a fully typed query API, and the generated migration (`client.Schema.Create`) keeps this prototype's setup to one command.
- **Redis for both caching and idempotency**: same store, two jobs — `entitlement:{user_id}` as a short-TTL read cache invalidated on every write, and `idempotency:{key}` as a `SETNX` lock so retried writes are safe.
- **oaswrap/spec**: code-first OpenAPI — routes and their docs are declared together, no separate YAML to keep in sync.
- Deliberately **not** using JWT/Firebase here, since the brief asks for "basic authentication." In the full system design (part 2), this same middleware slot is where Firebase JWT verification would go instead.

## Setup

```bash
docker compose up -d          # Postgres + Redis
cp .env.example .env
go mod tidy                   # resolves & pins dependency versions (needs network)
go generate ./ent             # generates ent/ client code from ent/schema
air                           # hot-reload dev server on :8080
```

Without Air: `go run .`

## Try it

```bash
# creates a user tied to this token on first call
curl -H "Authorization: Bearer demo-token-123" http://localhost:8080/v1/me

curl -H "Authorization: Bearer demo-token-123" http://localhost:8080/v1/subscription

curl -X PUT http://localhost:8080/v1/subscription \
  -H "Authorization: Bearer demo-token-123" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: test-key-1" \
  -d '{"status":"active","plan":"pro_monthly_99"}'

# same Idempotency-Key again → returns current state, does not re-apply
curl -X PUT http://localhost:8080/v1/subscription \
  -H "Authorization: Bearer demo-token-123" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: test-key-1" \
  -d '{"status":"expired"}'
```

Docs UI: `http://localhost:8080/docs`

## Known simplification

`go generate` and `go mod tidy` need real network access to `proxy.golang.org`, which wasn't available while writing this — so `ent/`'s generated client isn't checked in. Run the two commands above once on your machine and it builds normally.
