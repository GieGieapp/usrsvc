# User Service – Setup & Usage

## Overview

Small CRUD service for **customers** and **nationalities** with clean layers:

* HTTP handlers → Usecase → Repository (Postgres via pgx)
* Full unit tests with mocks/pgxmock
* Request validation & consistent error payloads

---

## Quick Start

```bash
# 1) clone & enter
git clone https://github.com/<you>/usrsvc.git
cd usrsvc

# 2) copy env and edit
cp .env.example .env

# 3) run DB migrations (DDL) + REQUIRED nationality seed
psql "$DATABASE_URL" -f db/migrations/001_init.sql
psql "$DATABASE_URL" -f db/seeds/nationality.sql

# 4) run app
go run ./cmd/usrsvc
```

---

## .env (required)

```
APP_PORT=8080
APP_ENV=development         # development|staging|production
LOG_LEVEL=info              # debug|info|warn|error

# Postgres (pick ONE style)
DATABASE_URL=postgres://user:pass@localhost:5432/usrsvc?sslmode=disable
# -- or split vars if you prefer:
# PGHOST=localhost
# PGPORT=5432
# PGUSER=user
# PGPASSWORD=pass
# PGDATABASE=usrsvc
# PGSSLMODE=disable
```

Optional tuning:

```
CORS_ALLOW_ORIGINS=http://localhost:3000
PG_POOL_MIN_CONNS=2
PG_POOL_MAX_CONNS=10
PG_POOL_MAX_CONN_LIFETIME=30m
READ_TIMEOUT=15
WRITE_TIMEOUT=15
IDLE_TIMEOUT=60
```

> Never commit `.env`. Add to `.gitignore`.

---

## Database Schema (minimal)

Tables used:

* `nationality (nationality_id PK, nationality_name TEXT, nationality_code TEXT NULL)`
* `customer (cst_id PK, nationality_id FK, cst_name, cst_dob DATE, cst_phoneNum, cst_email UNIQUE)`
* `family_list (cst_id FK, fl_relation, fl_name, fl_dob DATE)`

Example DDL (excerpt):

```sql
CREATE TABLE IF NOT EXISTS nationality (
  nationality_id   SERIAL PRIMARY KEY,
  nationality_name TEXT NOT NULL,
  nationality_code TEXT
);
```

---

## REQUIRED: Seed Initial Nationalities

The app expects base nationalities to exist. Run once on a fresh DB:

**File:** `db/seeds/nationality.sql`

```sql
-- Minimal global set (extend as needed)
INSERT INTO nationality (nationality_name, nationality_code) VALUES
  ('Indonesia', 'ID'),
  ('Malaysia', 'MY'),
  ('Singapore', 'SG'),
  ('Thailand', 'TH'),
  ('Philippines', 'PH')
ON CONFLICT DO NOTHING; -- if you use a unique constraint on (nationality_name)
```

**Apply:**

```bash
psql "$DATABASE_URL" -f db/seeds/nationality.sql
```

> If you don’t keep a seeds file, you can run a one-off:

```bash
psql "$DATABASE_URL" -c "INSERT INTO nationality (nationality_name, nationality_code) VALUES
('Indonesia','ID'),('Malaysia','MY'),('Singapore','SG'),('Thailand','TH'),('Philippines','PH')
ON CONFLICT DO NOTHING;"
```

---

## API Summary

### GET `/nationalities`

List all nationalities.
**200** → `[{ "id":1, "name":"Indonesia", "code":"ID" }, ...]`

### GET `/users?page=1&size=10&search=AL`

Paginated list with optional search.
**200** → `{"data":[...], "total":42}`

### GET `/users/{id}`

**200** → Customer
**404** → Not found

### POST `/users`

Create customer (+ optional family). Dates must be `YYYY-MM-DD`.
**201** → Created
**409** → Email exists
**422** → Validation error

### PUT `/users/{id}`

Update customer.
**200** → `{"status":"ok"}`

### DELETE `/users/{id}`

**200** → `{"status":"ok"}`

---

## Tests

Install dev deps:

```bash
go get github.com/stretchr/testify
go get github.com/pashagolub/pgxmock/v3
go install github.com/vektra/mockery/v2@latest
```

Run:

```bash
go test ./... -v
```

* Handlers: success/validation/conflict/not found/internal errors
* Usecase: page/size normalization, repo mapping
* Repository: pgxmock (insert tx, unique violation `23505`, commit/rollback, list nationalities incl. scan/query errors)

---

## Notes

* Keep type consistency across layers (`int` vs `int32`) to avoid mock issues.
* `nationality_code` is optional (`NULL` allowed); adjust seeds if you require it non-null.
* Always run the **nationality seed** on new environments (local, CI, staging, prod) before serving traffic.
