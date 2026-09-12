# Finance Planner – Specification

A small desktop application to plan a household budget: income and spendings
are entered once, grouped by category, and the application derives what has
to be transferred to the bank account and to the savings account every month.
This document is the complete specification of the application as it is
implemented. It is the reference for behaviour, data, technology and visual
design, so that the application could be rebuilt from it.

---

## 1. Purpose and planning model

### 1.1 Goal

Enter all recurring income and spendings, categorise them, and see at a glance:

- how much money is needed per month and per year,
- the saldo per month and per year,
- how much has to be transferred to the **bank account** each month,
- how much has to be put aside on the **savings account** each month,
- when non-monthly payments are due and how large the savings buffer has to be.

### 1.2 The planning model (background)

Spendings that are paid **monthly** are paid from the bank account, so their
sum is the amount that has to be transferred to the bank account every month.

Spendings that are paid **less often** (quarterly, half-yearly, yearly) are
saved up: every month 1/n of the amount (n = months between two payments) is
put on the savings account, so that the full amount is available when the
payment is due and can be moved from the savings account to the bank account.
This is a sliding window: after a payment the saving starts again from zero.

The statistics box makes both transfers visible, and the payment timeline
shows the resulting balance of the savings account over a year.

---

## 2. Functional specification

### 2.1 Kinds

There are two kinds of data: **income** and **spending**. Categories belong to
exactly one kind; entries inherit the kind of their category. An entry can
never be moved into a category of the other kind.

### 2.2 Categories

- A category has a **name** and a **kind** (income or spending).
- Categories must exist before entries can be added. If the user wants to add
  an entry while no category of that kind exists, the entry dialog explains
  this and offers to create a category first; after the category was created,
  the entry dialog opens again.
- Names are trimmed, must not be empty and must be unique per kind
  (case-insensitive). A category can be **renamed**.
- A category can only be **deleted** when it contains no entries. Deleting a
  category that still has entries is refused with an explanation.
- Categories can be **collapsed** and expanded individually; the collapsed
  state is saved. At the top of each block (income and spending) there are
  **expand all** and **collapse all** buttons for that block.
- Categories can be **reordered by drag & drop** within their block; the
  order is saved.

### 2.3 Entries

An entry (income or spending) has these fields:

| Field | Rule |
|---|---|
| Name | required, trimmed |
| Amount in € | required, greater than 0, stored as integer cents |
| Period | one of **monthly**, **quarterly**, **half-yearly**, **yearly**; this is the *master* value the user entered |
| Category | required, of the same kind |
| Due month | 1–12 or unset; only for non-monthly periods (monthly entries never keep a due month). Quarterly and half-yearly entries pay every 3 or 6 months starting at that month |
| Paused | if set, the entry stays in the table but is excluded from every subtotal and statistic |
| Notes | optional free text (contract number, cancellation date, …) |

Operations on entries:

- **Add**, **edit** and **delete** (delete asks for confirmation).
- **Duplicate**: opens the entry dialog prefilled with the values of an
  existing entry (name with a "(copy)" suffix) to create a new one.
- **Pause / resume** directly from the table row.
- **Reorder by drag & drop** within the category and **move by drag & drop**
  into another category of the same kind. No confirmation is required, but
  the target position must be clearly visible while dragging.
- Double-clicking a row opens the edit dialog.

### 2.4 Calculations

All amounts are integer cents. With *n* = months between two payments
(1, 3, 6 or 12):

- per month = round(amount / n) (rounded half away from zero to whole cents)
- per year = amount × 12 / n (exact)

Subtotals of a category and all statistics are sums of these per-entry
values, so the table columns always add up to the shown totals. Paused
entries contribute nothing.

### 2.5 Main view

The main view shows two blocks, **income first, spendings below**, and the
statistics box on the right.

Each block has:

- a header with the kind, the block totals per month and per year, and the
  buttons *Expand all*, *Collapse all*, *Category* (new category) and
  *Income* / *Spending* (new entry; in German the singular *Einnahme* /
  *Ausgabe*, while the block titles use the plural *Einnahmen* / *Ausgaben*);
- a column header: Name · Frequency (German: Zahlweise) · Due · Per month · Per year;
- the categories in their saved order, each as a collapsible group with a
  drag handle, the name, the number of entries, its subtotals per month and
  per year and the actions *add entry*, *rename*, *delete*;
- for spending categories, the **share of all spending** as a thin bar with a
  percentage in the category header.

Each entry row shows: drag handle · name (followed by a note icon with
tooltip when notes exist) · a **period badge** (monthly / quarterly /
half-yearly / yearly) · the **due month** as a full month name in its own
column (empty for monthly entries and unset due months) · per month ·
per year · actions in this order: edit, duplicate, pause/resume, delete. The value the user
entered (the master) is printed bold; the derived value is muted. Paused rows
carry a subtle diagonal stripe pattern and muted text and badge.

An empty block explains that a category has to be added first; while a
search or filter is active and nothing matches, it says so instead. When the
planner contains no data at all, a "getting started" card offers to load
**sample data**.

### 2.6 Drag & drop

- Categories are dragged by their header, entries by their row.
- While dragging an entry, a horizontal insertion line shows the exact target
  position between rows; a category that would receive the entry is
  highlighted, an empty or collapsed category shows a "drop here" area.
- While dragging a category, an insertion line shows the target position
  between categories of the same block.
- Inside a block every area accepts the drag: the gaps between categories
  are drop positions for category drags, and other areas (block header,
  padding) keep the last target, so the cursor never flips to "not allowed"
  while moving across the block. `dragenter` is cancelled at block level for
  compatible drags, because WebKit otherwise shows "not allowed" for a
  moment at every element boundary. Outside the blocks the indicator
  disappears. Dropping an item on its own position is a no-op.
- Drag & drop is disabled while a search or filter is active, because the
  visible order would not match the stored order.

### 2.7 Statistics box

A sticky box on the right side with these sections:

**Monthly transfers**
- *To bank account*: sum of all spendings paid monthly.
- *To savings account*: sum of the monthly share (amount / n) of all
  spendings that are not paid monthly.

**Overview**
- Income per month, average cost per month, saldo per month.
- *Savings goal per month*: an amount the user wants to put aside, editable
  in a modal (0 removes the goal); when set, *remaining after goal*
  (saldo − goal) is shown, red when negative.
- Income per year, cost per year, saldo per year.

**Payment timeline**
- Twelve columns (January–December): bars show the payments due in each
  month, a step line shows the balance of the savings account at the end of
  each month in the steady state. For an entry with due month *d* and
  period *n* the balance at the end of month *t* is
  monthly × ((t − d) mod n). Hovering a month shows its values.
- *Savings buffer needed (peak)*: the highest balance of the year.
- Notes tell how many non-monthly entries have no due month (they are not
  in the timeline) and how many entries are paused.

**Chart**
- A donut of spending by category (largest first, at most eight slices, the
  rest folded into "Other") with a legend naming every slice and its amount.
  (Income vs. spending is not charted; the numbers are in the overview.)

**Warnings** (banner above the blocks)
- red when spending exceeds income, showing the monthly gap;
- amber when the savings goal is not reachable, showing the shortfall.

### 2.8 Search and filter

The top bar contains a search box and, as a separate control next to it, a
period filter dropdown (all periods, monthly, quarterly, half-yearly, yearly,
paused only). The search matches entry names
and notes (case-insensitive). While a filter is active only matching entries
are listed, categories without matches are hidden, and a hint shows "x of y
entries shown". The search box has its own clear button (tooltip "Clear
search"); next to the hint a reset button (tooltip "Reset search and filter")
clears both the search text and the period filter. Subtotals stay those of
the whole category.

### 2.9 Dialogs, notifications and undo

- **Every data entry happens in a modal dialog** (entry, category, savings
  goal). Dialogs close with the *Escape* key, the close icon or a click on
  the backdrop; the first form field gets the focus.
- **Deletions and destructive actions ask for confirmation** in a modal
  (delete entry, delete category, load sample data, restore backup).
- **Errors** are shown in a modal (for actions outside a form) or inline in
  the form they belong to.
- **Success and information** are shown as popup toasts in the lower right
  corner. Toasts with an action stay 9 seconds, others 3.5 seconds, errors
  8 seconds.
- **Undo**: the toast after deleting an entry or a category offers *Undo*,
  which restores the item with its original id at its original position. The
  toast after an import offers *Undo import*, which restores the backup that
  was written right before the import. Toasts with an action show a thin
  countdown bar at their bottom edge that shrinks over the toast's lifetime,
  so it is visible how long the action is still available.
- Unhandled errors in the UI are never silent: they are shown in an error
  modal.

### 2.10 Import and export

**JSON export** writes the complete data file to a location chosen in a
native save dialog.

**JSON import** opens a native file dialog, validates the file, shows what it
contains (categories, entries, file version) and asks whether the data should
be **added** to or **replace** the current data:

- *Replace* discards all current categories and entries. The language,
  savings goal and window settings of this installation are kept.
- *Add* (merge): categories are matched by id (same id and kind) first, then
  by kind and name (case-insensitive); unmatched categories are added and keep
  their id. Entries whose id already exists are **skipped**; all others are
  added with their original id. The toast reports how many entries were
  added, how many already existed and how many categories were created.

**CSV export** writes all entries as a flat table (kind, category, name,
period, amount, per month, per year, due month, paused, notes) in UTF-8 with
byte order mark. In German the separator is `;` with a decimal comma,
otherwise `,` with a decimal point.

**Backups**: before ordinary changes (at most once every 10 minutes) and
always before an import or a restore, the current data file is copied to
`backups/data-<YYYYMMDD-HHMMSS.mmm>.json` next to the data file. The last 20
backups are kept. A *Backups* dialog lists them (time, number of categories
and entries) and restores one after confirmation; only files inside the
backup folder can be restored.

**Sample data**: an empty planner offers a small example set (income and
spending categories with typical household entries, including quarterly,
half-yearly and yearly ones with due months) in the current language. It can
only be loaded when there are no categories and entries.

### 2.11 Language

- The UI is available in **German** and **English**; a dropdown in the top
  right corner switches the language and the choice is saved.
- **German is the default** until a choice has been made: the saved language
  stays empty, and an empty or unknown language maps to German.
- Number and currency formatting follow the language (`1.234,56 €` in
  German, `€1,234.56` in English); amounts can be typed with comma or point.
- All texts, including backend error messages, come from language files with
  keys (see 4.5); adding a language means adding one file and one registry
  line.

### 2.12 Appearance

- A dropdown in the top right corner (next to the language dropdown)
  selects the color scheme: **Light**, **Dark** or **System** (follow the
  operating system). The choice is saved in `data.json`.
- **Light is the default** until a choice has been made (empty value in the
  file). "System" is resolved in the frontend and follows changes of the
  operating system setting live.

### 2.13 Keyboard shortcuts

| Key | Action |
|---|---|
| `n` | New spending |
| `i` | New income |
| `c` | New spending category |
| `/` or `Ctrl+F` | Focus the search box |
| `Esc` | Clear the search (when the search box is focused) / close a dialog |
| `?` | Show the list of shortcuts |

Shortcuts are ignored while a dialog is open or an input field has the focus.

### 2.14 Window

- Default size 1440 × 900, minimum 1200 × 700, so the table with all its
  columns and the statistics box always fit side by side.
- The last window size and position are saved and restored on the next start.
- The page content is centered and capped at 1600 px width so it does not
  stretch endlessly on very wide screens; below 1140 px the page keeps its
  layout instead of squishing.
- The application icon is a blue rounded square (`#2f5fd6`) with a white
  € sign (`build/appicon.png`, 1024 × 1024). The Windows `icon.ico`
  (embedded into the executable, shown by Explorer) contains the sizes 16,
  24, 32, 48 and 64 as uncompressed 32-bit bitmaps and 128 and 256 as PNG;
  the macOS `.icns` is generated from the PNG.

---

## 3. Data and persistence

### 3.1 Data file

All data lives in a single JSON file **`data.json` next to the binary**. It
is created on the first start, written atomically (temporary file, then
rename) after every change, and human-readable (indented).

```json
{
  "version": 3,
  "settings": {
    "language": "de",
    "theme": "dark",
    "savingsGoalCents": 30000,
    "window": { "width": 1440, "height": 900, "x": 120, "y": 80 }
  },
  "categories": [
    { "id": "5f2c…", "name": "Wohnen", "kind": "spending", "collapsed": false }
  ],
  "entries": [
    {
      "id": "9a1e…",
      "categoryId": "5f2c…",
      "name": "Kfz-Versicherung",
      "amountCents": 62000,
      "period": "yearly",
      "dueMonth": 1,
      "paused": false,
      "notes": "Vertrag 4711"
    }
  ]
}
```

- Ids are random 16-hex-character strings.
- The order of `categories` (per kind) and of `entries` (per category) is the
  display order.
- `language` is empty until the user chose a language; `theme` is empty
  until the user chose a color scheme (empty means light). `theme` was added
  to version 3 without a version bump, since an absent value is valid.
- `dueMonth`, `paused` and `notes` are omitted when they have their zero value.

### 3.2 Versioning and migration

The file carries a `version`. The application migrates older files step by
step when loading and refuses files with a newer version than it knows.

| Version | Change |
|---|---|
| 1 | initial structure (categories, entries with monthly/yearly period) |
| 2 | `settings.language` |
| 3 | periods quarterly/half-yearly; entry `dueMonth`, `paused`, `notes`; `settings.savingsGoalCents`, `settings.window` |

Loading validates referential integrity (unique ids, entries reference
existing categories, known kinds and periods, non-negative amounts, due month
0–12, valid language code) and rejects corrupt files with a clear error.

### 3.3 Backups

`backups/` next to `data.json`, files named `data-<timestamp>.json`, newest
20 kept. See 2.10.

### 3.4 Errors

The backend never returns free-text errors for user mistakes. It returns a
**code** with parameters (for example `category.inUse` with `name` and
`count`), which the frontend translates. Codes in use:

`kind.unknown`, `period.unknown`, `category.notFound`, `category.nameEmpty`,
`category.exists`, `category.inUse`, `category.idExists`, `entry.notFound`,
`entry.categoryRequired`, `entry.nameEmpty`, `entry.amountPositive`,
`entry.dueMonthInvalid`, `entry.kindMismatch`, `entry.idExists`,
`import.modeUnknown`, `import.invalidFile`, `language.invalid`,
`settings.savingsGoalNegative`, `settings.themeInvalid`, `sample.notEmpty`,
`backup.invalidPath`.

---

## 4. Technical implementation

### 4.1 Stack

- **Wails v3** (Go) for the desktop shell, native dialogs and the service
  layer; **Svelte 5** with **TypeScript** and Vite for the frontend.
- No additional npm packages beyond the Wails Svelte template. Charts are
  inline SVG, drag & drop uses the native HTML5 drag events, tests use Node's
  built-in test runner. Anything that has to be installed is asked for first.

### 4.2 Architecture

- The Go **service** (`planner` package) owns the data: every mutating method
  validates, changes the data, saves the file and returns the complete new
  view state (categories with entries, derived values, statistics). The
  frontend replaces its state with the result, so business logic exists only
  once.
- Go structs are exposed to TypeScript through generated **bindings**
  (`wails3 generate bindings -ts -i -clean=true`), regenerated whenever a
  service signature changes.
- The frontend keeps a single reactive **store** (state, open dialog, toasts,
  filter) and small components: application shell, blocks, category groups,
  entry rows, statistics panel, timeline, charts and one component per
  dialog.
- Props of the open dialog are bound to a local constant in the shell, so a
  dialog can close itself and still read its props afterwards (Svelte props
  are getters).

### 4.3 Project layout

```
main.go                    window setup, service registration, window geometry
planner/
  model.go                 data structures, validation, language check
  calc.go                  monthly/yearly conversion, statistics, timeline, views
  store.go                 data.json loading, versioning/migration, atomic saving
  backup.go                automatic backups
  csv.go                   CSV export
  sample.go                example data
  errors.go                coded errors and their JSON marshalling
  service.go               the service used by the frontend
  planner_test.go          tests against this specification
frontend/src/
  App.svelte               shell: top bar, search, shortcuts, blocks, statistics, dialogs
  i18n/en.ts, de.ts        language files (en.ts defines the key set)
  i18n/index.ts            language registry, interpolation, plural helper
  lib/i18n.svelte.ts       reactive t()/plural()/formatting for components
  lib/store.svelte.ts      application state, service calls, filter, undo actions
  lib/dnd.svelte.ts        drag & drop state and drop handling
  lib/money.ts             € parsing and formatting (cents based)
  lib/reorder.ts           drag index arithmetic
  components/              Block, CategoryGroup, EntryRow, StatsPanel, Timeline, Charts, dialogs, Toasts
```

### 4.4 Persistence details

- Amounts are integer cents everywhere; the frontend converts to and from a
  human-readable € string.
- Saving writes a temporary file and renames it, so a crash never leaves a
  truncated data file.
- Window geometry is written debounced (500 ms) and only when it changed.

### 4.5 Internationalisation

- One TypeScript file per language exporting a flat object of dot-namespaced
  keys (`"stats.toBank": "To bank account"`). The English file defines the
  key set; every other file is typed against it, so a missing or unknown key
  is a compile error. A unit test additionally checks identical key sets and
  identical placeholders.
- Placeholders are written `{name}`. Plural forms use `<key>.one` /
  `<key>.other` and a `plural(key, count)` helper.
- A registry lists every language with its code, its native label and the
  BCP 47 tag used for number formatting (`de-DE`, `en-IE`).
- Backend error codes are translated under `errors.<code>`; kind names inside
  messages are translated too.

### 4.6 Tests

- **Go**: every requirement of section 2 and 3 that lives in the backend
  (category rules, validation, conversions and rounding, statistics, timeline,
  ordering and moves, restore, collapse state, persistence, versioning and
  migration, import merge/replace rules, backups, CSV, sample data, language
  setting, coded errors).
- **Node**: money parsing and formatting per locale, drag index arithmetic,
  language file completeness and interpolation, and a compile-level test that
  dialog props in the shell are not bound directly to the mutable dialog
  state.
- `svelte-check` must report no errors and no warnings.

### 4.7 Build, run and workflow

```sh
wails3 dev          # run with hot reload
wails3 build        # production build -> bin/finance-planner
wails3 task test    # Go tests + frontend unit tests
```

- Work happens on the `main` branch, structured in commits after stages that
  make sense (each commit builds and passes all tests).
- The window is tested by launching the built binary; regenerate bindings
  before building when the Go service changed.

---

## 5. Design and styling

### 5.1 Principles

- Calm, light, table-first interface: clear fonts, tabular numbers, few
  colors with a fixed meaning (green = income, rust = spending, blue =
  actions, violet = savings), generous but compact spacing.
- Color never carries information alone: badges have text, chart slices have
  a legend with names and amounts, status banners have icons.
- Dark mode is an explicit choice (light by default, dark, or follow the
  system) with its own token values (not an inverted palette); chart colors
  are re-stepped for the dark surface.
- Every visual value is a CSS custom property in `frontend/public/style.css`;
  components use tokens only.

### 5.2 Layout and dimensions

| Element | Value |
|---|---|
| Window default / minimum | 1440 × 900 / 1200 × 700 |
| Page max / min width | 1600 px / 1140 px, centered, 24 px side padding |
| Content grid | two columns: tables `minmax(0, 1fr)`, statistics 320 px, 20 px gap |
| Statistics box | sticky; both columns start 20 px below the top bar (margin on the columns, not padding on the scroll area, so the sticky box stays level with the income block) |
| Top bar | 10 px 24 px padding, title left, search box (200–420 px) and the period filter dropdown as separate controls in the middle, actions and language dropdown right |
| Table columns | `28px minmax(150px, 1fr) 120px 110px 125px 125px 120px` (handle, name, frequency, due, per month, per year, actions) |
| Row height | 38 px min; category header 40 px |
| Corner radius | 10 px cards and dialogs, 6 px buttons, badges and inputs |
| Shadows | cards `0 1px 2px rgba(20,26,40,.06), 0 4px 16px rgba(20,26,40,.06)`; dialogs `0 12px 40px rgba(20,26,40,.22)` |

### 5.3 Typography

- Font stack: `Inter, system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif`; monospace for key caps.
- Base 14 px, line height 1.45, antialiased, `font-variant-numeric: tabular-nums` everywhere.
- Sizes: app title 18 px, block title 17 px, dialog title 17 px, statistics
  heading 15 px, section labels 11 px uppercase with 0.06 em letter spacing,
  hints and secondary text 11.5–12.5 px, badges 11 px uppercase.
- Weights: 400 for entry names, 500 for buttons, 600 for headings and master
  values, 700 for totals and category titles. Category titles are 15 px in
  the primary text color so they stand apart from the 14 px regular entry
  names; only the thin accent bar of the header carries the block color.
  Derived values use the muted text color.
- Text selection and the default cursor are disabled on the page (desktop
  application feel); inputs and error texts remain selectable.

### 5.4 Color tokens

Light mode (default) and dark mode (applied as `data-theme="dark"` on the
root element; "system" is resolved to one of the two in the frontend):

| Token | Light | Dark | Use |
|---|---|---|---|
| `--bg` | `#f5f6fa` | `#12151c` | page background |
| `--surface` | `#ffffff` | `#1b1f2a` | cards, rows, dialogs |
| `--surface-2` | `#f8f9fc` | `#202533` | hover, search box, tiles |
| `--surface-3` | `#eef0f5` | `#2a3040` | category headers, pressed state |
| `--border` | `#dfe3ea` | `#2f3646` | dividers |
| `--border-strong` | `#c8cdd8` | `#3d4557` | inputs, buttons |
| `--text` | `#1c2130` | `#e8ebf2` | primary text |
| `--text-2` | `#5b6474` | `#aab2c3` | secondary text |
| `--text-3` | `#8a93a5` | `#7c8598` | muted text, icons, derived values |
| `--accent` / `-soft` / `-strong` | `#2f5fd6` / `#e6edfb` / `#1f47ad` | `#5b8def` / `#1f2c47` / `#7fa6f5` | primary buttons, focus, drop indicators, monthly badge |
| `--income` / `-soft` / `-strong` | `#1e8a5a` / `#e5f4ec` / `#166b45` | `#3cbf82` / `#17302a` / `#5ed39c` | income block and button, positive values |
| `--spending` / `-soft` / `-strong` | `#c2542c` / `#fbeae3` / `#9c4222` | `#e07a55` / `#3a241c` / `#f0956f` | spending block, share bars, due bars |
| `--positive` / `--negative` | `#1e8a5a` / `#c9302c` | `#3cbf82` / `#ef6b6b` | signed saldo values |
| `--danger` / `-soft` | `#c9302c` / `#fbe7e6` | `#ef6b6b` / `#3a1f21` | delete buttons, error banner and toast |
| `--warn` / `-soft` | `#8a5a10` / `#fff4dc` | `#e0b25c` / `#3a2f16` | savings goal warning |
| `--savings` / `-soft` | `#6b3fa0` / `#f3ecfb` | `#a98ae6` / `#2b2340` | savings tile, timeline line, yearly badge |
| `--quarterly` / `-soft` | `#1b6f7d` / `#e6f4f6` | `#5fc0cf` / `#172e33` | quarterly badge |
| `--halfyearly` / `-soft` | `#8a5a10` / `#fbf0e0` | `#e0b25c` / `#3a2f16` | half-yearly badge |
| `--toast-bg` | `#1c2130` | `#2a3040` | info toasts |

### 5.5 Chart palette

Categorical colors for the spending donut, assigned in fixed order by
category rank (largest first); a ninth and further categories are folded
into "Other". The palette passes the lightness, chroma, color-vision and
contrast checks of the palette validator in both modes; in light mode three
colors sit below 3:1 contrast against white, which is why every slice is also
named with its amount in the legend.

| Slot | Light | Dark |
|---|---|---|
| 1 | `#2a78d6` | `#3987e5` |
| 2 | `#eb6834` | `#d95926` |
| 3 | `#1baf7a` | `#199e70` |
| 4 | `#eda100` | `#c98500` |
| 5 | `#e87ba4` | `#d55181` |
| 6 | `#008300` | `#008300` |
| 7 | `#4a3aa7` | `#9085e9` |
| 8 | `#e34948` | `#e66767` |
| Other | `#9aa3b5` | `#6b7385` |

Chart marks: donut with a 12-unit stroke on a 100-unit view box, 2-unit gaps
between slices, the hovered slice grows to 14; the timeline uses thin bars
(55 % of the column) and a 2 px step line. The share bar in a category
header keeps a 24 px gap to the due column.

### 5.6 Components

- **Buttons**: 7 × 12 px padding (small: 4 × 9 px, 13 px text), 1 px strong
  border, white surface, hover darkens the surface; primary in accent, income
  and spending buttons in their color, danger in red; icon buttons are
  28 × 28 px transparent squares that get a surface-3 background on hover.
- **Inputs and selects**: 8 × 10 px padding, strong border, accent border on
  focus, 2 px accent outline for keyboard focus; the amount input carries a
  trailing € sign; the period choice is a segmented control (accent-soft
  background for the active segment).
- **Badges**: pill, 11 px uppercase, semantic color pairs (monthly = accent,
  quarterly, half-yearly, yearly = savings). A row never shows more than one
  badge, so the frequency column never wraps.
- **Category header**: surface-3 background with a 3 px left accent bar in
  the block color (the left padding is reduced by 3 px so the columns stay
  aligned with the rows), title 15 px bold in the primary text color,
  chevron that rotates 90° when expanded, count pill in secondary text,
  subtotals right-aligned in secondary text, actions appear on hover. The
  bar is the only colored element of the header.
- **Entry row**: separated by 1 px borders, hover surface-2, actions appear
  on hover, dragged rows at 35 % opacity. Paused rows: a diagonal stripe
  pattern (`repeating-linear-gradient(135deg, var(--stripe) 0 3px,
  transparent 3px 10px)`, stripe color `rgba(20,26,40,.07)` light /
  `rgba(255,255,255,.07)` dark) over the normal row background; name, amounts
  and due month in muted text; the period badge in surface-3 with muted text.
- **Due column**: text centered in its column, so it sits midway between the
  period badge and the right-aligned amounts.
- **Drop indicators**: 3 px accent line between rows and 3 px line in the
  block color between categories, both drawn as absolutely positioned
  overlays without transitions so the table never shifts while dragging; a target category gets an accent border and a
  soft accent ring; empty and collapsed targets show a soft accent area with
  "Drop here".
- **Statistics tiles**: surface-2 with a 4 px left border in accent (bank)
  or savings color; value 22 px semibold.
- **Banners**: 10 × 14 px padding, icon plus text, danger or warn colors.
- **Dialogs**: centered panel on a dark translucent backdrop
  (`rgba(10,13,20,.55)`), 12 px radius, title row with close icon, body,
  footer with right-aligned buttons; short fade and pop-in animation.
- **Toasts**: lower right, dark surface with white text, success in
  income-strong, error in danger, optional outlined action button, close
  icon; slide-in animation; toasts with an action have a 3 px white
  (70 % opacity) countdown bar along the bottom edge that scales from full
  width to zero linearly over the toast's duration.
- **Key caps** in the shortcut list: bordered inline blocks in the monospace
  font.

### 5.7 Interaction details

- Actions in rows and category headers are hidden until hover to keep the
  table calm.
- The appearance dropdown (sun icon) and the language dropdown (globe icon)
  sit at the far right of the top bar.
- The current month is printed bold in the timeline axis; the readout below
  the timeline always uses two fixed lines (due, on savings account) so
  hovering never reflows the box.
- Focus is visible everywhere (2 px accent outline).

---

## 6. Working agreement

- Read the whole specification before starting. Ask open questions first;
  once implementation has started, do not interrupt with questions until it
  is finished.
- Keep this document in sync with the implementation: every new or changed
  behaviour is specified here first or documented here afterwards.
- Do not install anything without asking.
- Commit in meaningful stages on `main`; every commit builds and passes the
  tests.
