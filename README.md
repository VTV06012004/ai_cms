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

Copy `.env.example` to `.env` and replace every `replace-with-...` value with
fresh random credentials before starting the services. Compose requires
`MYSQL_PASSWORD` and `MYSQL_ROOT_PASSWORD`; the API also reads `DB_DSN` when
run directly outside Compose. Never commit `.env`. If a credential was
previously committed, treat it as compromised and rotate it in the database
and every service that uses it; removing the file from a new commit does not
remove it from Git history.

`CORS_ALLOWED_ORIGINS` is a comma-separated list. Next.js public environment
values are embedded during image build, so rebuild the web service after
changing `NEXT_PUBLIC_API_URL`.

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
