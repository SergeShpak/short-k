# Shortik: Simple URL Shortener service

This repository contains code of a simple URL Shortener service.

## Running locally

To run shortik locally use docker compose:

```bash
SHORTIK_DSN="postgresql://shortik:shortik@shortik-db:5432/shortik?sslmode=disable" docker compose up -d
```

# Development

This section contains information on the service development. Everything should run smoothly on a Linux AMD64 machine.

## Prerequisites

You should install the following tools locally to simplify the service development:

- `go` >= 1.22: https://go.dev/doc/install
- `docker`: https://docs.docker.com/engine/install/
- `lefthook`: a tool to manage git hooks, https://github.com/evilmartians/lefthook?tab=readme-ov-file#install

You should run `lefthook` install after fetching the repository for the first time: it will configure the required git hooks.

## IDE configuration

As the project uses [the `tools.go` pattern](https://www.jvt.me/posts/2022/06/15/go-tools-dependency-management/), build tags and so on, you may want to tweak the static analysis performed by your IDE. If you use VS code it suffices to define the following workspace settings:

```json
{
    "gopls": {
        "build.buildFlags": [
            "-tags=e2e_tests"
        ],
        "build.directoryFilters": [
            "-tools/"
        ]
    }
}
```

## Generated files

All files are generated with `go:generate`. If you want to regenerate anything, run this from the repository root:

```bash
go generate ./...
```

## Testing

To execute unit tests run

```bash
go test -v ./... -count=1
```

To run all tests (including e2e tests), first make sure there is a shortik service running somewhere. You can start the shortik service locally:

```bash
SHORTIK_DSN="postgresql://shortik:shortik@shortik-db:5432/shortik?sslmode=disable" docker compose up -d
```

The run:

```bash
SHORTIK_HOST=http://localhost:8080 go test -v --tags=e2e_tests ./... -count=1
```

## Working with DB

### Migrations

Migration files are located in `internal/infra/store/db/migrations`. They are applied automatically using [golang-migrate/migrate](https://github.com/golang-migrate/migrate).

### Add/change SQL queries

Here are the steps to add or change the existing SQL queries:

1. change/add queries in `internal/infra/store/db/internal/queries/queries.sql`
2. run `go generate ./...` from the repository root

# Roadmap

Here is the list of features to add.

## Short-term

- [ ] Raise test coverage (currently only `internal/infra/store/db` is covered well)
- [ ] Add test-coverage badge
- [ ] Finish documenting code (currently only `internal/infra/store/db` is docuemented well)
- [ ] Add better structured logging

## Mid-term

- [ ] Add clicks couter (using metrics probably)
- [ ] Configure observability
- [ ] Add JWT-sessions
- [ ] Cofigure TSL for the HTTP-server

## Long-term

- [ ] Add Terraform configuration to deploy the service in the cloud
- [ ] Implement user authentication/authorization
- [ ] Optimize slug duplicates checks (do not query DB each time)
- [ ] Add rate limiting
