# DDD CMS

A CMS built with a Domain-Driven Design layout: a Go API (Gin + GORM + MySQL) and a
Next.js frontend, run together with Docker Compose.

## Quick start (Docker)

1. Clone the `main` branch:

   ```sh
   git clone -b main https://github.com/VTV06012004/ai_cms.git
   cd ai_cms
   ```

2. Create your local environment file from the template:

   ```sh
   cp .env.example .env        # Windows: copy .env.example .env
   ```

3. Edit `.env` and replace every `replace-with-...` value:

   | Variable | Purpose |
   | --- | --- |
   | `JWT_SECRET` | Signing key for login tokens. Required; the API exits without it. |
   | `MYSQL_PASSWORD`, `MYSQL_ROOT_PASSWORD` | Database passwords. Required by Compose. |
   | `INITIAL_ADMIN_EMAIL`, `INITIAL_ADMIN_PASSWORD` | Credentials of the first admin account (password: at least 6 characters). Used only when the database has no admin yet. |

   Use random values, for example `openssl rand -hex 32`. Never commit `.env`.

4. Start everything:

   ```sh
   docker compose up --build
   ```

5. Open the app and sign in with the admin email and password you set in `.env`:

   - Frontend: `http://localhost:3001` (sign in at `/login`)
   - API: `http://localhost:8081` (health check: `http://localhost:8081/health`)

MySQL is only exposed to the Compose network.

To stop the containers while keeping database data and uploaded files:

```sh
docker compose down
```

To also delete the database and start again from an empty one, add `-v`
(`docker compose down -v`).

## Database migrations

The schema is managed by versioned SQL files in `migrations/` (`0001_init.sql`, `0002_...`).
The API never changes the schema when it starts; a separate command applies the files:

- `docker compose up` runs the `migrate` service first and starts the API only after it succeeds.
- Manually: `docker compose run --rm migrate` (or `make migrate` with `DB_DSN` set).

Rules:

- Never edit a migration that has already run anywhere. `migrate` stores a checksum of every
  applied file and refuses to continue if one changed. Add a new file with the next number instead.
- In production, run the migration **before** deploying the new API binary.
- MySQL commits DDL statements immediately, so a migration that fails halfway is not rolled back.
  Fix the database by hand, then run `migrate` again.

### Existing database created before migrations

A database created by the old AutoMigrate startup has tables but no `schema_migrations` table,
and `migrate` refuses to touch it. Make sure it is on the latest schema, then record `0001`
as applied without running it:

```sh
docker compose run --rm migrate ./migrate -baseline 1
```

## Business rules

- **Editing a published post.** When an author edits a post that is already published
  (or scheduled), the post goes back to `review` and an editor must approve it again.
  Editors and admins can edit published posts without losing their status.
- **Publication time.** `published_at` is set the first time a post becomes published,
  whether through moderation, direct publishing by an editor or admin, or a scheduled
  publication.
- **Scheduled posts.** Only editors and admins can schedule a post. Editing a scheduled
  post without sending `scheduled_at` keeps the existing schedule; a new schedule must
  be in the future.

## Configuration notes

## API documentation

After starting the stack, open `http://localhost:8081/docs` for the Swagger UI.
Sign in via `POST /api/auth/login`, copy the token, click **Authorize** and paste it
to try the protected endpoints. The raw spec is at `/docs/openapi.yaml`
(source: `internal/delivery/http/openapi.yaml`). A test fails if a route is added
without documenting it, so keep the spec in sync.

- **Secrets.** There is no default JWT secret, and the application exits if
  `JWT_SECRET` is missing. Compose also requires `MYSQL_PASSWORD` and
  `MYSQL_ROOT_PASSWORD`. When the API runs directly outside Compose it reads
  `DB_DSN` instead, and exits if it is missing. Never commit `.env`. If a
  credential was ever committed, treat it as compromised and rotate it in the
  database and every service that uses it; removing the file in a new commit does
  not remove it from Git history.
- **Login tokens** expire after 15 minutes. A role change therefore takes effect
  for existing tokens no later than their expiration; users must sign in again to
  receive a token with their updated role.
- **First admin.** On a database with no admin account, set `INITIAL_ADMIN_EMAIL`
  and `INITIAL_ADMIN_PASSWORD` in `.env` before starting Compose. The API creates
  this root admin once; later starts do not change its credentials or role. Public
  registrations remain authors. Existing databases that already have an admin can
  start without these variables; their oldest admin is protected as the root admin.
- **CORS and frontend URL.** `CORS_ALLOWED_ORIGINS` is a comma-separated list.
  Next.js public environment values are embedded during image build, so rebuild
  the web service after changing `NEXT_PUBLIC_API_URL`.
- **Database passwords on an existing volume.** MySQL applies `MYSQL_PASSWORD` and
  `MYSQL_ROOT_PASSWORD` only when the data volume is first created. If you change
  them later, either keep using the old values or run `docker compose down -v` to
  recreate the database.

The frontend is copied from the reference project into `web/` and configured to
call the DDD API through `NEXT_PUBLIC_API_URL`. Do not edit the reference
`ai_cms` project to work on this copy.

## Troubleshooting

- **The API exits right after the first start with `connection refused`.** MySQL was
  still initialising. Run `docker compose up -d` again.
- **The build fails at the `gofmt` step (Windows).** Go files were checked out with
  CRLF line endings. The repository has a `.gitattributes` that forces LF for
  `*.go`, so clone with `-b main` as shown above. For an existing clone, run
  `git rm --cached -r . && git reset --hard` (this discards uncommitted changes).
- **`required variable ... is missing a value`.** Create `.env` from `.env.example`
  as described in the quick start.
- **The API starts but the post list is empty after a while.** Your browser still
  holds an expired token. Sign out and sign in again, or clear `token` from the
  browser's local storage.

## Layer responsibilities

- `internal/domain`: business entities, invariants, value types, and repository contracts.
- `internal/usecase`: application workflows; depends on domain contracts.
- `internal/delivery/http`: HTTP routing, request validation, and explicit API response DTO mapping.
- `internal/infrastructure`: MySQL/GORM persistence models and repository adapters, security, and file storage.
- `cmd/server`: application composition root; creates adapters and wires use cases
  to HTTP handlers.
- `web`: Next.js frontend.

Domain entities do not contain JSON or GORM tags. API representation is mapped
in Delivery, while database representation and conversion are handled in
Infrastructure. Use cases orchestrate workflows; entity methods enforce core
rules such as valid roles, post updates, and moderation transitions.

## Tests

Run the unit tests with `go test ./...`. The Docker build also checks `gofmt` and
runs the tests before building the server.

On Windows without a local Go install, run Go commands through Docker from the
repository root:

    docker run --rm -v "%cd%":/src -w /src golang:1.21 gofmt -l .
    docker run --rm -v "%cd%":/src -w /src golang:1.21 gofmt -w .
    docker run --rm -v "%cd%":/src -w /src golang:1.21 go test ./...

`gofmt -l .` lists files that are not formatted (the Docker build fails if there are
any), and `gofmt -w .` fixes them in place. Run the last two before committing.

Post slug uniqueness integration tests require a disposable MySQL 8.4 instance
and `MYSQL_TEST_DSN` with permission to create and drop databases. The tests
create isolated temporary databases and remove them when finished. Do not point
this variable at a production database. GitHub Actions runs the suite against a
temporary MySQL service.
