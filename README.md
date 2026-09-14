# Finance Planner

A small desktop application to plan monthly and yearly income and spendings.
Built with [Wails v3](https://v3.wails.io), Svelte 5 and TypeScript; the
specification is in [harness/specs.md](harness/specs.md); the reference
copies of the language files live next to it (see the specification, 7.10).

## How it works

- Income and spendings are grouped by **categories** (separate categories for
  income and spending). Categories have to exist before entries can be added
  and can only be deleted when they are empty.
- Each entry has a name, an amount in € and a **period** (monthly,
  quarterly, half-yearly or yearly). The period the user entered is the
  master; the monthly and yearly values are calculated from it and shown
  next to each other.
- Categories and entries can be reordered with **drag & drop**; entries can be
  dragged into another category of the same kind. The order is saved.
- The statistics box shows the two monthly transfers the planning approach
  is built on: the sum of monthly spendings (to the bank account) and 1/n of
  every spending paid less often (to the savings account), plus the saldo per
  month and per year. The timeline reads as a plan that is already running;
  see the specification, 1.2.
- The UI is available in **German and English**; German is the default until
  a language is chosen. The dropdown in the top right corner switches the
  language and the choice is saved in `data.json`.
- The top bar carries the search box, the period filter, the **Data** menu
  (import, export, CSV export, backups), Print, the command palette and the
  two setting dropdowns. Along the bottom edge a **status bar** shows the
  path of the data file (click to copy), how much is planned, and the
  version.
- All data is stored in `data.json` next to the binary. The file carries a
  `version` number so the structure can be migrated later. Data can be
  exported to and imported from JSON files (add to or replace current data).
  When adding, categories are matched by id, then by name; entries whose id
  already exists are skipped, so overlapping files never create duplicates.

## More features

- **Periods**: entries can be paid monthly, quarterly, half-yearly or yearly.
  Only monthly entries go to the bank account; all others are saved up with
  1/n of the amount per month.
- **Due month**: non-monthly entries can carry the month of a payment. The
  statistics box then shows a 12-month timeline of what is due when, the
  balance of the savings account, and the peak buffer it needs.
- **Paused entries** stay in the table but are excluded from every total.
- **Notes** per entry (shown as a tooltip and in the edit dialog).
- **Savings goal** per month with the remaining amount, and banners when
  spending exceeds income or the goal is not reachable.
- **Charts and levers**: share of each spending category (bar in the
  category header and a donut) and a "biggest levers" list of the five most
  expensive spendings per year.
- **Undo**: deleting an entry or category and importing a file can be undone
  from the notification.
- **Backups**: written to `backups/` next to `data.json` before changes (at
  most every 10 minutes) and before every import; the last 20 are kept and
  can be restored from the Backups dialog.
- **CSV export** for spreadsheets (UTF-8 with BOM; `;` and decimal comma in
  German, `,` and decimal point otherwise).
- **Multi-select** (header checkbox of a block or Ctrl+click) with bulk move,
  pause, resume and delete (with undo).
- **Fast entry**: *Save and add another* keeps the dialog open for the following
  entry (`Ctrl+Enter`), new entries start in the category and frequency last
  used, and a new category can continue straight into its first entry.
- **Command palette** (`Ctrl+K`) over actions, categories and entries.
- **Print / PDF** via the system print dialog (`Ctrl+P`); all categories are
  expanded, one column, light colors.
- **Sample data** for an empty planner, **search and period filter**,
  **duplicate entry**, an **appearance** setting (light by default, dark, or
  follow the system), explanatory **tooltips** on every value, and the
  window size and position are remembered.

### Keyboard shortcuts

| Key | Action |
|---|---|
| `n` | New spending |
| `i` | New income |
| `c` | New spending category |
| `/` or `Ctrl+F` | Search |
| `Ctrl+K` | Command palette |
| `Ctrl+P` | Print |
| `Esc` | Clear search / clear selection / close dialog |
| `?` | Show shortcuts |

Shortcuts are ignored while a dialog or an input field is focused.

## Development

```sh
cd frontend && npm install   # once after cloning
wails3 dev                   # run with hot reload
wails3 build                 # production build -> bin/finance-planner
wails3 task test             # Go tests + frontend unit tests
wails3 task test:e2e         # end-to-end tests in a browser
wails3 task test:e2e:report  # open the report of the last end-to-end run
```

The tests can also be run directly:

```sh
go test ./planner/...          # backend: model, calculations, persistence, service
cd frontend && npm test        # frontend helpers (Node's built-in test runner)
cd frontend && npm run check   # svelte-check / TypeScript
cd frontend && npx playwright test   # end-to-end, needs a built frontend
cd frontend && npx playwright show-report   # the report of the last run
```

The end-to-end tests drive the real application in Chromium: `cmd/e2e-host`
serves the built frontend and answers the Wails calls from a real service, so
the shipped code runs unmodified. The browser is a one-time setup:

```sh
cd frontend && npx playwright install chromium
sudo npx playwright install-deps chromium   # Linux: its system libraries
```

Note: `go build`/`go vet` at the module root need `frontend/dist` to exist
(it is embedded into the binary), so run `wails3 build` once first.

## Adding a language

1. Copy `frontend/src/i18n/en.ts` to `<code>.ts` and translate the values.
   The `Messages` type makes TypeScript report missing or unknown keys, and
   the i18n unit test checks that placeholders match.
2. Register it in the `locales` list in `frontend/src/i18n/index.ts` with its
   label and the BCP 47 tag used for currency formatting.

Backend errors are returned as codes (`planner/errors.go`) and translated
in the frontend under the `errors.*` keys.

## Layout

```
main.go                    window setup and service registration
planner/
  model.go                 data structures (Data, Category, Entry) and validation
  calc.go                  monthly/yearly conversion, statistics, view models
  store.go                 data.json loading, versioning/migration, atomic saving
  service.go               the service used by the frontend (CRUD, moves, import/export)
  errors.go                coded errors that the frontend translates
  backup.go                automatic backups next to data.json
  csv.go                   CSV export
  sample.go                example data set
  version.go               the application version (checked against build/config.yml)
  planner_test.go          tests against the specification
frontend/src/
  App.svelte               shell: top bar, language dropdown, blocks, statistics, dialogs
  i18n/en.ts, de.ts        language files (en.ts defines the key set)
  i18n/index.ts            language registry and translation helpers
  lib/i18n.svelte.ts       reactive t()/plural()/formatEuro() for components
  lib/store.svelte.ts      application state and service calls
  lib/dnd.svelte.ts        drag & drop state and drop handling
  lib/money.ts             € parsing and formatting (cents based)
  lib/theme.svelte.ts      light/dark/system handling
  lib/reorder.ts           drag index arithmetic
  lib/amountField.ts       amount validation while typing
  components/              Block, CategoryGroup, EntryRow, StatsPanel, Timeline,
                           Charts, Menu, StatusBar, SelectionBar, CommandPalette,
                           Toasts, Modal, Icon, dialogs
frontend/e2e/              end-to-end tests (Playwright)
cmd/e2e-host/              test host: serves the frontend and the service over HTTP
harness/
  specs.md                 the specification
  en.ts, de.ts             reference copies of the language files
build/config.yml           application metadata (name, identifier, version)
```
