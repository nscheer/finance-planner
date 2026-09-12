# Finance Planner

A small desktop application to plan monthly and yearly income and spendings.
Built with [Wails v3](https://v3.wails.io), Svelte 5 and TypeScript; the
specification is in [project.md](project.md).

## How it works

- Income and spendings are grouped by **categories** (separate categories for
  income and spending). Categories have to exist before entries can be added
  and can only be deleted when they are empty.
- Each entry has a name, an amount in € and a **period** (per month or per
  year). The period the user entered is the master; the other value is
  calculated (× 12 or ÷ 12) and shown next to it.
- Categories and entries can be reordered with **drag & drop**; entries can be
  dragged into another category of the same kind. The order is saved.
- The statistics box shows the two monthly transfers the planning approach
  is built on: the sum of monthly spendings (to the bank account) and 1/12 of
  the yearly spendings (to the savings account), plus the saldo per month and
  per year.
- All data is stored in `data.json` next to the binary. The file carries a
  `version` number so the structure can be migrated later. Data can be
  exported to and imported from JSON files (add to or replace current data).

## Development

```sh
wails3 dev          # run with hot reload
wails3 build        # production build -> bin/finance-planner
wails3 task test    # Go tests + frontend unit tests
```

The tests can also be run directly:

```sh
go test ./planner/...          # backend: model, calculations, persistence, service
cd frontend && npm test        # frontend helpers (Node's built-in test runner)
cd frontend && npm run check   # svelte-check / TypeScript
```

Note: `go build`/`go vet` at the module root need `frontend/dist` to exist
(it is embedded into the binary), so run `wails3 build` once first.

## Layout

```
main.go                    window setup and service registration
planner/
  model.go                 data structures (Data, Category, Entry) and validation
  calc.go                  monthly/yearly conversion, statistics, view models
  store.go                 data.json loading, versioning/migration, atomic saving
  service.go               the service used by the frontend (CRUD, moves, import/export)
  planner_test.go          tests against project.md
frontend/src/
  App.svelte               shell: top bar, blocks, statistics, dialogs
  lib/store.svelte.ts      application state and service calls
  lib/dnd.svelte.ts        drag & drop state and drop handling
  lib/money.ts             € parsing and formatting (cents based)
  components/              Block, CategoryGroup, EntryRow, StatsPanel, dialogs
```
