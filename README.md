# DDD CMS

## Run with Docker

From this directory:

```sh
docker compose up --build -d
```

The frontend is available at `http://localhost:3001` (open `/login` to sign in).
The API is available at `http://localhost:8081`; MySQL is only exposed to the
Compose network. The health endpoint is `http://localhost:8081/health`.

To stop the containers while keeping database data and uploaded files:

```sh
docker compose down
```

The frontend is copied from the reference project into `web/` and configured to
call the DDD API through `NEXT_PUBLIC_API_URL`. Do not edit the reference
`ai_cms` project to work on this copy.

Set a secret before starting the API. There is no default JWT secret, and the
application exits if `JWT_SECRET` is missing. For local Docker Compose use,
create a `.env` file in this directory and set `JWT_SECRET` to a randomly
generated value (for example, generate one with `openssl rand -base64 32`).
Do not commit `.env` or use a predictable value.

The API issues JWTs that expire after 15 minutes. A role change therefore takes
effect for existing tokens no later than their expiration; users must sign in
again to receive a token with their updated role.

Database credentials and `CORS_ALLOWED_ORIGINS` can also be configured with
environment variables or the local `.env` file.
`CORS_ALLOWED_ORIGINS` is a comma-separated list and defaults to the local
frontend origins on ports 3000 and 3001. Next.js public environment values are
embedded during image build, so rebuild the web service after changing
`NEXT_PUBLIC_API_URL`.

On a database with no admin account, configure `INITIAL_ADMIN_EMAIL` and
`INITIAL_ADMIN_PASSWORD` in a local `.env` file before starting Compose. The API
creates this root admin once; later starts do not change its credentials or
role. Public registrations remain authors. Choose a strong password and do not
commit the `.env` file. Existing databases with admin accounts can start without
these bootstrap variables; their oldest admin is protected as the root admin.

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

Post slug uniqueness integration tests require a disposable MySQL 8.4 instance
and `MYSQL_TEST_DSN` with permission to create and drop databases. The tests
create isolated temporary databases and remove them when finished. Do not point
this variable at a production database. GitHub Actions runs the suite against a
temporary MySQL service.
