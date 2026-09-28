# AGENTS.md

## Scope: docs and tests only

Agents may write and edit **documentation and tests only**. Do not generate production logic — no new or modified non-test implementation code.

- Allowed: `*_test.go`, docs (`README*`, `docs/`, `doc.go`), and test-only fixtures/snapshots.
- Allowed: **doc comments** — write or edit docblocks on existing declarations (types, funcs, methods, vars, consts) in production `.go` files. This is the one production-file edit permitted; it must not change behavior, signatures, or move code.
- Not allowed: changes to production code (beyond the comments above), `Makefile`, `.github/workflows/`, or config.
- **Suggest, don't apply.** For anything outside docs/tests/doc comments, describe the exact change (file, line, proposed diff) and stop. Do not run the edit, and do not stage or commit it.
- If a test exposes a bug in production code, report it with a failing-test repro or a written description; leave the fix to a human.
- Docblocks in this repo follow standard Go style: start with the identifier's name, then a full sentence. Package docs live in `doc.go`. Don't duplicate what the code plainly shows — document the non-obvious (why, invariants, gotchas).

## Repository

Go module `gosalusa.com` (Go 1.27). A web framework **library** — not an app. There is no `main` at the root; consumers import subpackages. Three things live in the one module:

- the framework packages (`di`, `kernel`, `request`, `router`, `database/*`, `openapidoc`, ...)
- the `spice` CLI (`spice/`) that scaffolds and drives projects built on the framework
- a real, compiling, tested project template (`static/template/`) embedded via `go:embed` and copied by `spice init`

`static/template` is a normal Go package tree: `go build ./...`, `go vet ./...` and `go test ./...` all compile and exercise it. Changes there break the repo.

## Commands

```bash
go test ./...              # full suite - NEEDS MySQL + Postgres, see below
go test -short ./...       # SQLite only; safe with no databases running
go test ./database/builder/ -run TestHasMany   # single package / single test
make lint                  # golangci-lint run
make coverage              # installs go-test-coverage, writes cover.out, enforces thresholds
```

There is no `make test` target. Don't invent one.

## Tests require real databases

`internal/test` hardcodes connection details — **no env vars, no skip**:

| driver | host | db | user | password |
| --- | --- | --- | --- | --- |
| mysql | localhost:3306 | test_db | root | root_password |
| postgres | localhost:5432 | test_db | user | password |

If they aren't up, tests **fail** (not skip). Start them with `docker compose up -d`; the credentials are duplicated in `docker-compose.yml` and `.github/workflows/coverage.yml` — change all three together.

- `go test -short ./...` runs SQLite in-memory only; that is the only `testing.Short()` check in the repo (`internal/test/test.go:287`).
- `make coverage` and the Coverage CI workflow run the **full** suite with real DBs.
- Each subtest runs in its own transaction that is rolled back; schema is created once per test binary. MySQL/Postgres setup takes an `flock` on files in `os.TempDir()`, which serializes concurrent test binaries — expect slow parallel `go test ./...` on the DB packages.
- Coverage gates are 80% **per package and total** (`.testcoverage.yml`), excluding `^static/template` and `^internal/test`. New code without tests will fail CI.

### Snapshots

Only `database/migrate` uses cupaloy; snapshots live in `database/migrate/.snapshots/`. Update with `UPDATE_SNAPSHOTS=1 go test ./database/migrate/`. The update run still reports a failure by design — re-run to confirm. A *missing* snapshot is auto-created and also reported as a failure; commit it.

## Generated code

`database/builder/generated_Builder.go` and `generated_ModelBuilder.go` are produced by `internal/build/build.go`:

```bash
go generate ./database/builder/...
```

Neither file carries a `Code generated ... DO NOT EDIT` header, so tooling and reviewers will not warn you. **Edit the source types and regenerate; do not hand-edit these files.**

Do **not** run `go generate ./openapidoc/...` casually — its directive curls `redoc.redoc.ly` and overwrites the 300KB+ embedded `redoc.standalone.js.gz`. The asset is committed and already current.

`//go:generate spice generate:migration` appears in template models and in `pubsub/dbpubsub`. Note it **overwrites `migrations.go`** in the migration dir on every run.

## Conventions that will bite you

**Model → table mapping** (`database/model/doc.go`, `internal/helpers/tag.go`):
- There is **no automatic snake_case conversion.** An untagged field is used as the column name verbatim.
- Table name is `snake_case(TypeName) + "s"` unless the type already ends in `s`, or it implements `Table() string`.
- Tags: `db:"name,primary,autoincrement"`, plus `nullable`, `readonly`, `index`, `unique`, `type:<sql type>`, `size:<n>`, and `db:"-"` to exclude. Pointers and `optional.Optional[T]` are nullable.

**Migrations** (`database/migrate`):
- Migrations self-register in `init()`; there is no filesystem discovery. `Migrations.Up` sorts by `Name` **lexicographically**, which only works because `spice make:migration` prefixes `YYYYMMDD_HHMMSS-`.
- `Migration.Down` is generated but there is no `Migrations.Down`/rollback runner. Only `Up` exists.
- A migration's `Up` and its bookkeeping row share one `sqlx.Tx`, so a failure rolls back both.

**DI** (`di`):
- Factories are keyed by `reflect.Type` only. The `inject` tag's first value is passed to the factory as a name but is **not** part of the key.
- `inject:"name,optional"` is the flag form; without it a missing registration is an error.
- Dependency cycles are only caught by an explicit `dp.Validate(ctx)`. `Resolve`/`Fill` will not detect them.
- `RegisterSingleton` builds eagerly and **panics** on factory error; `RegisterLazySingleton` caches the error and replays it on every later build.

**Query builder**:
- `Builder.Clone` does not copy `ForUpdate` — a cloned locking query silently loses the lock mode.
- `Get` returns an empty slice and nil error on `sql.ErrNoRows`; `First` returns the zero value and nil.

**Kernel**:
- `Bootstrap` builds the root handler *before* registering services, so anything the handler injects must come from the provider, not from `Services`.
- `Bootstrap` returns `ErrAlreadyBootstrapped` on a second call — build a fresh `*kernel.Kernel` per test. `kerneltest.NewTestKernelFactory` does this for you.

**spice config** is `spice.yml`, not `spice.yaml` (`spice/util/config.go:32`).

## Test helpers worth knowing

- `internal/test` — `test.Run(t, name, cb)` fans a test out across sqlite/mysql/pgsql subtests; `test.QueryTest`/`UpdateQueryTest`/`CreateTableTest`/... are table-driven SQL-encoder helpers; shared models `test.Foo`, `test.Bar`, `test.FooSoftDelete`.
- `testing/handlertest` — fluent in-process HTTP. `handlertest.New(ctx, t, handler).WithJSONHeaders().PostJSON(...)`, assertions via `AssertStatus2XX`, `AssertJSON[T]`, `AssertJSONContains(path, matcher)`. Bodies decode to `float64`, so numeric expectations are `match.Equal(1.0)`. `AssertJSONContains` cannot index into JSON arrays.
- `testing/match` — `match.Equal`, `Same`, `Len`, `Greater`/`Less`, `All`. Ordering matchers type-assert to `T`, so use `float64` against decoded JSON. `Len` **passes** when length can't be determined.
- `database/dbtest` — `dbtest.NewRunner(open)` for consumer projects; `.Factory[T]` with `.Count(n)`/`.State(...)`/`.Create(tx)`.
- Other doubles: `email/emailtest`, `router/routertest`, `pubsub/pubsubtest`, `di.TestDependencyProviderContext()`.

## Repo hygiene

- `.golangci.yml` is **v2 schema** and contains only exclusion rules — default linter set applies. Test files are excluded from `staticcheck` SA1029 and `errcheck` "is not checked".
- `cover.out` and `.vscode` are gitignored; `cover.out` is ~14MB if it exists locally, leave it.
- Default branch is `main`. Commits are short imperative summaries, usually with a `(#PR)` suffix.
