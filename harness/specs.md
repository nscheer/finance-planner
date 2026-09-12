# Finance Planner – Specification

A small desktop application to plan a household budget: income and spendings
are entered once, grouped by category, and the application derives what has
to be transferred to the bank account and to the savings account every month.

This document is the complete specification of the application as it is
implemented. It is the reference for behaviour, data, technology and visual
design, so that the application could be rebuilt from it.

This file lives in `harness/` together with the reference copies of the
language files (see 7.10).

**Contents**

1. Purpose and planning model
2. Data model and rules
3. User interface
4. Data exchange and safety
5. Settings
6. Persistence
7. Technical implementation
8. Design and styling
9. Working agreement

---

## 1. Purpose and planning model

### 1.1 Goal

Enter all recurring income and spendings, categorise them, and see at a glance:

- how much money is needed per month and per year,
- the saldo per month and per year,
- how much has to be transferred to the **bank account** each month,
- how much has to be put aside on the **savings account** each month,
- when non-monthly payments are due and how large the savings buffer has to be,
- which spendings are the biggest levers.

### 1.2 The planning model

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

## 2. Data model and rules

### 2.1 Kinds

There are two kinds of data: **income** and **spending**. Categories belong to
exactly one kind; entries inherit the kind of their category. An entry can
never be moved into a category of the other kind.

### 2.2 Categories

- A category has a **name** and a **kind**.
- Categories must exist before entries can be added. If the user wants to add
  an entry while no category of that kind exists, the entry dialog explains
  this and offers to create a category first; afterwards the entry dialog
  opens again.
- Names are trimmed, must not be empty and must be unique per kind
  (case-insensitive). A category can be **renamed**.
- A category can only be **deleted** when it contains no entries; otherwise
  the deletion is refused with an explanation.
- Categories can be **collapsed** individually; the collapsed state is saved.
- The order of categories inside a block is the user's order (drag & drop)
  and is saved.

### 2.3 Entries

| Field | Rule |
|---|---|
| Name | required, trimmed |
| Amount in € | required, greater than 0, stored as integer cents |
| Period | **monthly**, **quarterly**, **half-yearly** or **yearly**; the *master* value the user entered |
| Category | required, of the same kind |
| Due month | 1–12 or unset; only for non-monthly periods (monthly entries never keep one). Quarterly and half-yearly entries pay every 3 or 6 months starting at that month |
| Paused | the entry stays in the table but is excluded from every subtotal and statistic |
| Notes | optional free text (contract number, cancellation date, …) |

Amount input is interpreted according to the **language**: the language's
decimal mark (`,` in German, `.` in English) is the decimal separator, the
other character is accepted only as a thousands separator in front of
exactly three digits. German: `12,5` = 12.50, `1.234,56` = 1234.56,
`10.123` = 10123, `12.50` is invalid. English: `12.5` = 12.50,
`1,234.56` = 1234.56, `10,123` = 10123, `12,50` is invalid. At most two
decimals; spaces and the € sign are ignored; a leading minus is parsed but
rejected by validation (amounts must be positive). Invalid input shows the
message "Please enter a valid amount, e.g. 12.50." (German: "z. B. 12,50"). Edit fields show the stored amount with the decimal mark of the
current language (`1234,56` in German, `1234.56` in English), and the
placeholder of an empty amount field is `0,00` or `0.00` accordingly.

Operations: add, edit, delete (with confirmation and undo), duplicate
(dialog prefilled with the values, name with a "(copy)" suffix), pause and
resume, reorder within the category and move to another category of the
same kind (drag & drop or bulk action). The order of entries inside a
category is saved.

### 2.4 Calculations

All amounts are integer cents. With *n* = months between two payments
(1, 3, 6 or 12):

- per month = round(amount / n), rounded half away from zero to whole cents
- per year = amount × 12 / n (exact)

Subtotals, block totals and all statistics are sums of these per-entry
values, so the table columns always add up to the shown totals. Paused
entries contribute nothing.

Derived statistics (see 3.7 for their presentation):

- *to bank account* = sum of the monthly values of all spendings paid monthly
- *to savings account* = sum of the monthly values of all other spendings
- *average cost per month* = bank + savings; *saldo* = income − cost
- *remaining after goal* = saldo per month − savings goal
- *timeline*: for a scheduled entry with due month *d* and period *n* the
  savings balance at the end of month *t* is monthly × ((t − d) mod n); the
  due amount of month *t* is the full amount when (t − d) mod n = 0. Entries
  without due month are not part of the timeline. *Peak buffer* = highest
  balance of the twelve months.
- *biggest levers* = the five active spendings with the highest yearly cost
  (ties broken by name), with their share of all monthly spending and of the
  monthly income
- *share of a spending category* = its monthly total / all monthly spending,
  and / monthly income

---

## 3. User interface

### 3.1 Window and layout

- Default window size 1440 × 900, minimum 1200 × 700, so the table with all
  its columns and the statistics box always fit side by side. Size and
  position are restored on the next start (see 5.3).
- The page is centered and capped at 1600 px width; below 1140 px it keeps
  its layout instead of squishing. The full-width main area scrolls, so the
  scrollbar sits at the window edge.
- Layout: a top bar, below it the income block, the spending block and,
  on the right, the sticky statistics box.

### 3.2 Top bar

From left to right:

1. Application title and the path of the data file.
2. **Search box** with its own clear button, and next to it the **period
   filter** dropdown (all periods, monthly, quarterly, half-yearly, yearly,
   paused only). While a filter is active a hint shows "x of y entries
   shown" and a reset button clears search and filter (see 3.6).
3. Buttons **Import**, **Export**, **Export CSV**, **Backups**, **Print**;
   icon buttons for the **command palette** (prompt icon `>_`) and the
   **shortcut list** (keyboard icon).
4. The **appearance** dropdown (sun icon) and the **language** dropdown
   (globe icon).

### 3.3 Main view

Two blocks, **income first, spendings below**. Each block has:

- a header with the kind, the block totals per month and per year, and the
  buttons *Expand all*, *Collapse all*, *Category* (new category) and
  *Income* / *Spending* (new entry; German uses the singular *Einnahme* /
  *Ausgabe* on the buttons and the plural *Einnahmen* / *Ausgaben* as block
  title);
- a column-title row: selection checkbox (see 3.5) · Name · Frequency
  (German: Zahlweise) · Due · Per month · Per year;
- the categories in their saved order, each as a collapsible group with a
  drag handle, the name, the number of entries, for spending categories the
  **share of all spending** as a thin bar with a percentage, its subtotals
  per month and per year, and the actions *add entry*, *rename*, *delete*
  (visible on hover).

Each entry row shows: drag handle (or checkbox in selection mode) · name,
followed by a note icon with tooltip when notes exist · a **period badge**
(monthly / quarterly / half-yearly / yearly) · the **due month** as a full
month name, centered, empty for monthly entries and unset due months ·
per month · per year · the actions edit, duplicate, pause/resume, delete
(visible on hover). The value the user entered is printed bold, the derived
value is muted. Paused rows carry a subtle diagonal stripe pattern and muted
text and badge. Double-clicking a row opens the edit dialog.

*Expand all* and *Collapse all* are disabled while the block has no
categories.

Empty states: a block without categories explains that a category has to be
added first; while a search or filter matches nothing it says so instead.
The timeline shows a hint to give non-monthly spendings a due month while
no entry is scheduled; the donut shows "No spendings yet." while there is
no active spending.
A planner without any data shows a "getting started" card that offers to
load **sample data** (see 4.4).

Every displayed money value or percentage carries a tooltip that explains
what it means and how it was calculated (block and category totals, entered
vs. calculated entry amounts with the formula, the two transfers, every
overview row, savings goal and remainder, peak buffer, timeline readout,
levers and donut legend). Tooltips use line breaks to stay short per line.

### 3.4 Drag & drop

- Categories are dragged by their header and reordered within their block;
  entries are dragged by their row, reordered within their category or moved
  into another category of the same kind. No confirmation is needed.
- While dragging an entry, a horizontal insertion line shows the exact
  target position between rows; a category that would receive the entry is
  highlighted, an empty or collapsed category shows a "drop here" area.
  While dragging a category, an insertion line shows the target position
  between categories.
- Inside a block every area accepts the drag: the gaps between categories
  are drop positions for category drags, all other areas keep the last
  target, so the cursor never flips to "not allowed" while moving across the
  block. Outside the blocks the indicator disappears. Dropping an item on its
  own position is a no-op.
- Drag & drop is disabled while a search or filter is active, because the
  visible order would not match the stored order.
- Move contract: a drop target is "insert before item *i*" of the target
  list. The backend's move operations take the index in the list *after*
  the dragged item was removed, so when an item moves downwards inside the
  same list the frontend subtracts one; a target index past the end appends.
  Dropping on the own position is a no-op.

### 3.5 Selection mode and bulk actions

- The checkbox in the column-title row of a block switches the **selection
  mode** on: every row shows a checkbox instead of the drag grip. Hovering
  never changes a row. Ctrl/Cmd+click on a row toggles it and enters the
  mode as well; Shift+click selects a range inside a category; the command
  palette offers "Select entries".
- In the mode the header checkbox is tri-state for its block (none / some /
  all visible entries selected); clicking it selects all entries of the
  block, or deselects them when all are selected.
- Selected rows are tinted in the soft accent color.
- A **toolbar floats at the bottom center** of the window for the whole time
  the mode is on; it never moves the content. It shows the count ("No
  entries selected" while empty) and offers: move to another category of
  the same kind (disabled when income and spending entries are mixed),
  pause, resume, delete (confirmation, then an Undo that restores all
  entries at their former positions), and a last button that reads *Done*
  while nothing is selected and *Clear selection* otherwise. The action
  buttons are disabled while nothing is selected.
- *Esc* and the *Done* / *Clear selection* button leave the mode; the grips
  return. Changing the search or filter clears the selection; after every
  state update the selection is pruned to entries that still exist.
- Bulk operations are all-or-nothing: an unknown id fails the whole call
  and nothing changes. A bulk move appends the entries to the target
  category in table order.

### 3.6 Search and filter

The search matches entry names and notes (case-insensitive); the period
filter restricts to one period or to paused entries. While a filter is
active only matching entries are listed, categories without matches are
hidden, and the hint "x of y entries shown" appears. Subtotals stay those of
the whole category.

### 3.7 Statistics box

A sticky box on the right with these sections, in this order:

1. **Monthly transfers**: tiles *To bank account* and *To savings account*.
2. **Overview**: income per month, average cost per month, saldo per month;
   *savings goal per month* (editable in a modal, 0 removes it) and, when
   set, *remaining after goal* (red when negative); income per year, cost per
   year, saldo per year.
3. **Payment timeline**: twelve columns (January–December) with bars for the
   payments due per month and a step line for the savings balance at the end
   of each month; the current month is printed bold. Hovering a month fills
   the two fixed readout lines below (*due*, *on savings account*). Below:
   *Savings buffer needed (peak)* and notes on how many non-monthly entries
   have no due month and how many entries are paused.
4. **Spending by category**: a donut (largest first, at most eight slices,
   the rest folded into "Other") with a legend naming every slice and its
   amount; hovering highlights a slice and shows its share in the center.
5. **Biggest levers**: the five most expensive active spendings per year with
   rank, name, category, yearly amount and share of all spending; the
   headline's tooltip explains the list, each item's tooltip shows its own
   values; clicking an item opens its edit dialog.
6. A short note explaining the planning model.

**Warnings** appear as a banner above the blocks: red when spending exceeds
income (with the monthly gap), otherwise amber when the savings goal is not
reachable (with the shortfall). Only one banner is shown; the red one has
priority.

### 3.8 Dialogs, notifications and undo

- **Every data entry happens in a modal dialog** (entry, category, savings
  goal). Dialogs close with *Esc*, the close icon or a click on the backdrop;
  the first form field gets the focus. Validation errors appear inline in the
  form.
- **Entry dialog**: name, amount with € suffix and a segmented control for
  the period, a live preview line "= x per month · y per year" while typing,
  the category (preselected when the dialog was opened from a category
  header, otherwise the first category of the kind), the due month select
  (only for non-monthly periods, with a hint naming the payment interval),
  notes, and the paused checkbox. Saving a monthly entry clears the due
  month. Duplicating uses the same dialog with the values prefilled.
- **Category dialog**: name only (add or rename). **Savings goal dialog**:
  one amount, empty or 0 removes the goal.
- **Import dialog**: file name, number of categories and entries, file
  version, and two option buttons with descriptions (add / replace).
- **Backups dialog**: one row per backup with date and time (medium date,
  short time in the current language), the counts and a *Restore* button.
- **Confirmations** in a modal: delete entry, delete category, delete
  selection, load sample data, restore backup.
- **Errors** outside a form are shown in a modal; unhandled errors in the UI
  are never silent.
- **Success and information** are popup toasts in the lower right corner:
  3.5 s, errors 8 s, toasts with an action 9 s. Toasts with an action show a
  thin countdown bar that shrinks over the toast's lifetime.
- **Undo**: deleting an entry, a category or a selection offers *Undo*, which
  restores the items with their original ids at their original positions. An
  import offers *Undo import*, which restores the backup written right before
  it.
- Further dialogs: import (add or replace), backups (list and restore),
  shortcuts, command palette.

### 3.9 Keyboard shortcuts and command palette

| Key | Action |
|---|---|
| `n` | New spending |
| `i` | New income |
| `c` | New spending category |
| `/` or `Ctrl+F` | Focus the search box |
| `Ctrl+K` | Command palette |
| `Ctrl+P` | Print |
| `Esc` | Clear the search (when the search box is focused) / leave selection mode / close a dialog |
| `?` | Show the list of shortcuts |

Single-key shortcuts are ignored while a dialog is open or a text field
(input, textarea, select) has the focus; a focused checkbox or button does
not block them.

The **command palette** (`Ctrl+K` or the prompt icon) is one text box that
searches, case-insensitively, over **actions** (new entry or category per
kind, import, export, CSV export, print, backups, savings goal, expand and
collapse all per block, select entries, shortcuts, appearance and language
choices, sample data on an empty planner), all **categories** (activating
one expands it and scrolls it into view) and all **entries** (activating one
opens its edit dialog). Results are grouped by section, at most 30 are
shown; arrow keys move the highlight, Enter runs it, Esc closes.

### 3.10 Print and PDF

*Print* in the top bar (or `Ctrl+P`) opens the system print dialog, which
also allows saving as PDF. While printing every category is rendered
expanded. The print layout is a single column: a header with application
name, print date and data file path, then the income and spending tables,
then the statistics box. Buttons, drag handles, the top bar, banners, toasts
and dialogs are hidden; shadows become borders; light colors are forced even
in dark mode; category groups and statistics sections avoid page breaks
inside; page margins are 15 mm.

---

## 4. Data exchange and safety

### 4.1 JSON export and import

**Export** writes the complete data file to a location chosen in a native
save dialog.

**Import** opens a native file dialog, validates the file, shows what it
contains (categories, entries, file version) and asks whether the data should
be **added** to or **replace** the current data:

- *Replace* discards all current categories and entries; the settings of this
  installation (language, appearance, savings goal, window) are kept.
- *Add* (merge): categories are matched by id (same id and kind) first, then
  by kind and name (case-insensitive); unmatched categories are added and keep
  their id (a category whose id exists with a different kind gets a fresh
  id). Entries whose id already exists are **skipped**; all others are added
  with their original id (an empty id gets a fresh one). The result reports
  categories added and reused, entries added and skipped, and the path of
  the backup written before the import; the toast shows the counts.

Both modes can be undone via the backup written right before the import.

### 4.2 CSV export

All entries as a flat table in UTF-8 with byte order mark, CRLF line
endings and RFC 4180 quoting. Income rows come first, then spending, each
in category order and entry order. In German the separator is `;` with a
decimal comma, otherwise `,` with a decimal point. Amounts have two
decimals; the due month is the number 1–12 or empty.

| Column | English | German |
|---|---|---|
| kind | Kind (`Income` / `Spending`) | Art (`Einnahme` / `Ausgabe`) |
| category | Category | Kategorie |
| name | Name | Name |
| period | Period (`monthly`, `quarterly`, `half-yearly`, `yearly`) | Zeitraum (`monatlich`, `vierteljährlich`, `halbjährlich`, `jährlich`) |
| amount | Amount | Betrag |
| per month | Per month | Pro Monat |
| per year | Per year | Pro Jahr |
| due month | Due month | Fälligkeitsmonat |
| paused | Paused (`yes` / `no`) | Pausiert (`ja` / `nein`) |
| notes | Notes | Notizen |

### 4.3 Backups

Before ordinary changes (at most once every 10 minutes) and always before an
import or a restore, the data file is copied to
`backups/data-<YYYYMMDD-HHMMSS.mmm>.json` next to it. The last 20 backups are
kept. The *Backups* dialog lists them (time, number of categories and
entries) and restores one after confirmation; only files inside the backup
folder can be restored, and the current data is backed up before a restore.
A restore, like a replace import, keeps all settings of this installation.
No backup is written while no data file exists yet; the timestamp has
millisecond resolution so consecutive backups never collide.

### 4.4 Sample data

An empty planner offers this example set in the current language (German /
English). It can only be loaded when there are no categories and entries.
Categories are created in the order of first appearance.

| Category | Entry | Amount | Period | Due month |
|---|---|---|---|---|
| Gehalt / Salary (income) | Gehalt / Salary | 2800,00 € | monthly | – |
| Gehalt / Salary (income) | Weihnachtsgeld / Christmas bonus | 1500,00 € | yearly | November |
| Sonstiges / Other (income) | Kindergeld / Child benefit | 250,00 € | monthly | – |
| Wohnen / Housing | Miete / Rent | 950,00 € | monthly | – |
| Wohnen / Housing | Strom / Electricity | 75,00 € | monthly | – |
| Wohnen / Housing | Internet / Internet | 39,99 € | monthly | – |
| Versicherungen / Insurance | Kfz-Versicherung / Car insurance | 620,00 € | yearly | January |
| Versicherungen / Insurance | Haftpflicht / Liability insurance | 65,00 € | yearly | April |
| Versicherungen / Insurance | Hausrat / Household insurance | 48,00 € | half-yearly | March |
| Mobilität / Mobility | Kfz-Steuer / Car tax | 180,00 € | yearly | July |
| Mobilität / Mobility | Tanken / Fuel | 120,00 € | monthly | – |
| Freizeit / Leisure | Fitnessstudio / Gym | 29,90 € | monthly | – |
| Freizeit / Leisure | Streaming / Streaming | 12,99 € | monthly | – |
| Freizeit / Leisure | Urlaub / Holiday | 1800,00 € | yearly | August |
| Rücklagen / Reserves | Zahnzusatzversicherung / Dental insurance | 90,00 € | quarterly | February |

---

## 5. Settings

All settings are stored in `data.json` (see 6.1).

### 5.1 Language

- German and English; the dropdown in the top bar switches the language and
  the choice is saved.
- **German is the default** until a choice has been made (empty value in the
  file); an empty or unknown language maps to German.
- Number and currency formatting follow the language (`1.234,56 €` in
  German, `€1,234.56` in English); amounts can be typed with comma or point.
- All texts, including backend error messages, come from language files with
  keys (see 7.5); the complete copy in both languages is part of this
  specification as the files `harness/en.ts` and `harness/de.ts` (see 7.10).
  Adding a language means adding one file and one registry line.
- Month names are the keys `month.1` to `month.12`; abbreviations (timeline
  axis, readout) are their first three letters.

### 5.2 Appearance

- **Light**, **Dark** or **System** (follow the operating system) via the
  dropdown in the top bar.
- **Light is the default** until a choice has been made (empty value).
  "System" is resolved in the frontend and follows changes of the operating
  system setting live.

### 5.3 Window geometry

The last window size and position are saved (debounced, only when changed)
and restored on the next start; sizes below the minimum are ignored.

### 5.4 Savings goal

A monthly amount, edited in a modal from the statistics box; 0 removes it.

---

## 6. Persistence

### 6.1 Data file

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
- `language` and `theme` are empty until the user chose (empty means German
  and light). `dueMonth`, `paused` and `notes` are omitted at their zero
  value.

### 6.2 Versioning and migration

The file carries a `version`. The application migrates older files step by
step when loading and refuses files with a newer version than it knows.

| Version | Change |
|---|---|
| 1 | initial structure (categories, entries with monthly/yearly period) |
| 2 | `settings.language` |
| 3 | periods quarterly/half-yearly; entry `dueMonth`, `paused`, `notes`; `settings.savingsGoalCents`, `settings.window`; later `settings.theme` (added without a bump, an absent value is valid) |

A file without a `version` field is treated as version 1.

Loading validates referential integrity (unique ids, entries reference
existing categories, known kinds, periods and themes, non-negative amounts
and savings goal, due month 0–12, valid language code) and rejects corrupt
files with a clear error.

### 6.3 Backup folder

`backups/` next to `data.json`, files named `data-<timestamp>.json`, newest
20 kept (see 4.3).

### 6.4 Error codes

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

## 7. Technical implementation

### 7.1 Stack

- **Wails v3** (Go) for the desktop shell, native dialogs and the service
  layer; **Svelte 5** with **TypeScript** and Vite for the frontend.
- No additional npm packages beyond the Wails Svelte template. Charts are
  inline SVG, drag & drop uses the native HTML5 drag events, tests use Node's
  built-in test runner. Anything that has to be installed is asked for first.

### 7.2 Architecture

- The Go **service** (`planner` package) owns the data: every mutating method
  validates, changes the data, saves the file and returns the complete new
  view state (categories with entries, derived values, statistics). The
  frontend replaces its state with the result, so business logic exists only
  once. Bulk operations (move, pause, delete, restore several entries) are
  single service calls.
- Service API (all mutating methods return the new `State` unless noted):

  | Method | Purpose |
  |---|---|
  | `GetState()` | current state |
  | `AddCategory(kind, name)`, `RenameCategory(id, name)`, `DeleteCategory(id)`, `RestoreCategory(category, index)` | categories |
  | `SetCategoryCollapsed(id, bool)`, `SetAllCollapsed(kind, bool)`, `MoveCategory(id, index)` | collapse and order |
  | `AddEntry(input)`, `UpdateEntry(id, input)`, `DeleteEntry(id)`, `RestoreEntry(entry, index)`, `MoveEntry(id, categoryId, index)`, `SetEntryPaused(id, bool)` | entries; `input` = {categoryId, name, amountCents, period, dueMonth, paused, notes} |
  | `MoveEntries(ids, categoryId)`, `SetEntriesPaused(ids, bool)`, `DeleteEntries(ids)`, `RestoreEntries([{entry, index}])` | bulk actions |
  | `SetLanguage(code)`, `SetTheme(name)`, `SetSavingsGoal(cents)`, `SetWindow(geometry)` → error only | settings |
  | `ExportData()` → path (dialog), `ExportTo(path)`, `ChooseImportFile()` → preview (dialog), `PreviewImport(path)`, `ImportData(path, mode)` → {state, categoriesAdded, categoriesReused, entriesAdded, entriesSkipped, backupPath} | JSON exchange |
  | `ExportCSV()` → path (dialog), `ExportCSVTo(path)` | CSV |
  | `ListBackups()` → [{path, time, categories, entries}], `RestoreBackup(path)` | backups |
  | `LoadSampleData(lang)` → {state, categories, entries} | sample data |

- View model (`State`): `version`, `dataPath`, `settings`, `income[]` and
  `spending[]` (category views: the category, its `entries[]` as entry views
  with `monthlyCents` and `yearlyCents`, the subtotals `monthlyCents` /
  `yearlyCents`, and `shareOfIncome`), and `stats`: income and spending per
  month and year, saldo per month and year, `toBankMonthlyCents`,
  `toSavingsMonthlyCents`, `savingsGoalCents`, `remainingAfterGoalCents`,
  `goalReachable`, `timeline[12]` of {`dueCents`, `savedCents`},
  `peakBufferCents`, `unscheduledCount`, `pausedCount`, `topSpendings[≤5]`
  of {id, name, categoryName, monthlyCents, yearlyCents, shareOfSpending,
  shareOfIncome}.
- Go structs are exposed to TypeScript through generated **bindings**
  (`wails3 generate bindings -ts -i -clean=true`), regenerated whenever a
  service signature changes.
- The frontend keeps one reactive **store** (state, open dialog, toasts,
  filter, selection, print flag), a **theme** store and an **i18n** layer,
  plus small components: application shell, blocks, category groups, entry
  rows, statistics panel, timeline, charts, selection toolbar, command
  palette and one component per dialog.

### 7.3 Project layout

```
harness/
  specs.md                 this specification
  en.ts, de.ts             reference copies of the language files (see 7.10)
main.go                    window setup, service registration, window geometry
planner/
  model.go                 data structures, validation, language and theme checks
  calc.go                  conversions, statistics, timeline, levers, views
  store.go                 data.json loading, versioning/migration, atomic saving
  backup.go                automatic backups
  csv.go                   CSV export
  sample.go                example data
  errors.go                coded errors and their JSON marshalling
  service.go               the service used by the frontend (incl. bulk operations)
  planner_test.go          tests against this specification
frontend/src/
  App.svelte               shell: top bar, search, shortcuts, blocks, statistics, dialogs
  i18n/en.ts, de.ts        language files (en.ts defines the key set)
  i18n/index.ts            language registry, interpolation, plural helper
  lib/i18n.svelte.ts       reactive t()/plural()/formatting for components
  lib/theme.svelte.ts      appearance (light/dark/system) handling
  lib/store.svelte.ts      application state, service calls, filter, selection, undo
  lib/dnd.svelte.ts        drag & drop state and drop handling
  lib/money.ts             € parsing and formatting (cents based)
  lib/reorder.ts           drag index arithmetic
  components/              Block, CategoryGroup, EntryRow, StatsPanel, Timeline,
                           Charts, SelectionBar, CommandPalette, Toasts, Modal,
                           Icon, and the dialogs (Entry, Category, SavingsGoal,
                           Import, Backups, Shortcuts, Confirm, Alert)
```

### 7.4 Persistence details

- Amounts are integer cents everywhere; the frontend converts to and from a
  human-readable € string.
- Saving writes a temporary file and renames it, so a crash never leaves a
  truncated data file.
- Window geometry is written debounced (500 ms) and only when it changed.

### 7.5 Internationalisation

- One TypeScript file per language exporting a flat object of dot-namespaced
  keys (`"stats.toBank": "To bank account"`). The English file defines the
  key set; every other file is typed against it, so a missing or unknown key
  is a compile error. A unit test additionally checks identical key sets and
  identical placeholders.
- Placeholders are written `{name}`. Plural forms use `<key>.one` /
  `<key>.other` and a `plural(key, count)` helper. Tooltip texts live under
  `tip.*`.
- A registry lists every language with its code, its native label and the
  BCP 47 tag used for number formatting (`de-DE`, `en-IE`).
- Backend error codes are translated under `errors.<code>`; kind names inside
  messages are translated too.

### 7.6 Tests

- **Go**: every backend requirement of sections 2, 4, 5 and 6 (category
  rules, validation, conversions and rounding, statistics, timeline, levers,
  ordering and moves, bulk operations, restore, collapse state, persistence,
  versioning and migration, import merge/replace rules, backups, CSV, sample
  data, language, theme and window settings, coded errors).
- **Node**: money parsing and formatting per locale, drag index arithmetic,
  language file completeness and interpolation, a compile-level test that
  dialog props in the shell are not bound directly to the mutable dialog
  state, and the harness sync test (see 7.10).
- `svelte-check` must report no errors and no warnings.

### 7.7 Build, run and workflow

```sh
wails3 dev          # run with hot reload
wails3 build        # production build -> bin/finance-planner
wails3 task test    # Go tests + frontend unit tests
```

- `wails3 task test` runs `go test ./planner/...` and `npm test`
  (`node --test "src/**/*.test.ts"`). Node runs the TypeScript tests
  directly, so relative imports inside `frontend/src/i18n` carry the `.ts`
  extension; `tsconfig.json` sets `allowImportingTsExtensions` and `noEmit`
  and excludes `*.test.ts` from `svelte-check`.
- `data.json` is git-ignored; `bin/` holds the built binary and, in
  development, its data file and backups.
- Work happens on the `main` branch, structured in commits after stages that
  make sense; every commit builds and passes all tests.
- Regenerate bindings before building when the Go service changed.

### 7.8 Packaging and icon

- `build/config.yml`: product name "Finance Planner", identifier
  `de.scheer.financeplanner`, company and copyright "Nicolai Scheer",
  description "Plan monthly and yearly income and spendings". The Wails
  application uses the same name and description; the window background
  colour is `rgb(245, 246, 250)` (the light page background).
- The application icon is a blue rounded square (`#2f5fd6`) with a white
  € sign. Geometry on a 1024 × 1024 canvas: square inset 40 px with corner
  radius 200; the € ring is centred at (560, 512) with outer radius 300 and
  inner radius 218, open on the right by ±48°; two horizontal bars 66 px
  high, centred 62 px above and below the ring centre, spanning x = 250 to
  585. For 16–32 px a bolder variant is used: inset 20, ring radius 330 with
  stroke 118, opening ±52°, bars 96 px high at ±100, x = 200 to 600.
  `build/appicon.png` is that rendering; `build/appicon.icon/Assets/wails_icon_vector.svg`
  is the vector form for the macOS icon composer.
- The Windows `icon.ico` embedded into the executable contains 16, 24, 32,
  48 and 64 px as uncompressed 32-bit bitmaps and 128 and 256 px as PNG; the
  macOS `.icns` is generated from the PNG. The icon generation task is
  skipped while both files exist, so the hand-built `.ico` is never
  overwritten.

### 7.9 Implementation notes

Non-obvious decisions that must survive a rewrite:

- **Dialog props**: Svelte 5 passes props as getters. The open dialog is
  bound to a local constant in the shell, so a dialog can close itself
  (`app.dialog = null`) and still read its props afterwards.
- **Drag & drop cursor**: `dragenter` is cancelled at block level for
  compatible drags, because WebKit otherwise shows "not allowed" for a moment
  at every element boundary; drop indicators are absolutely positioned
  overlays without transitions so the table never shifts while dragging.
- **Escape**: only text fields block the shortcuts; a focused checkbox or
  button must not swallow *Esc* (otherwise selection mode could not be left
  by keyboard after clicking a checkbox).
- **Sticky statistics**: the 20 px top gap is a margin on both columns, not
  padding on the scroll container, so WebKit keeps the sticky box level with
  the income block.
- **Button padding**: 1 px less on top than at the bottom, because system
  fonts such as Segoe UI render low in their line box.
- **Errors**: coded errors are marshalled through the Wails service option
  `MarshalError` and arrive in the frontend as `error.cause = {code, params}`.

### 7.10 Specification harness

The folder `harness/` holds this specification and the **reference copies
of the language files** `en.ts` and `de.ts`. They are the complete UI copy
in both languages and are kept as real TypeScript files rather than as a
table, so they can be used directly.

Sync mechanism: the copies must be byte-identical to
`frontend/src/i18n/en.ts` and `frontend/src/i18n/de.ts`. The Node test
`frontend/src/i18n/harness.test.ts` (part of `npm test` and therefore of
`wails3 task test`) compares the files and fails with the name of the file
that differs. Whenever a text is added or changed in the application, copy
the changed language file into `harness/` in the same commit:

```sh
cp frontend/src/i18n/en.ts frontend/src/i18n/de.ts harness/
```

---

## 8. Design and styling

### 8.1 Principles

- Calm, light, table-first interface: clear fonts, tabular numbers, few
  colors with a fixed meaning (green = income, rust = spending, blue =
  actions, violet = savings), generous but compact spacing.
- Color never carries information alone: badges have text, chart slices have
  a legend with names and amounts, status banners have icons.
- Dark mode is an explicit choice with its own token values (not an inverted
  palette); chart colors are re-stepped for the dark surface.
- Every visual value is a CSS custom property in `frontend/public/style.css`;
  components use tokens only.

### 8.2 Layout and dimensions

| Element | Value |
|---|---|
| Window default / minimum | 1440 × 900 / 1200 × 700 |
| Page max / min width | 1600 px / 1140 px, centered, 24 px side padding |
| Content grid | the full-width main area is the scroll container (scrollbar at the window edge); inside it a centered page wrapper holds two columns: tables `minmax(0, 1fr)`, statistics 320 px, 20 px gap |
| Statistics box | sticky; both columns start 20 px below the top bar (margin on the columns) |
| Top bar | 10 px 24 px padding; title left; search box (200–420 px) and the period filter dropdown as separate controls in the middle; buttons, icon buttons, appearance and language dropdowns right |
| Table columns | `28px minmax(150px, 1fr) 120px 110px 125px 125px 120px` (handle, name, frequency, due, per month, per year, actions) |
| Row height | 38 px min; category header 40 px |
| Block header | 28 px gap between the block header (title, totals, buttons) and the column titles (20 px margin + 8 px padding) |
| Corner radius | 10 px cards and dialogs, 6 px buttons, badges and inputs |
| Shadows | cards `0 1px 2px rgba(20,26,40,.06), 0 4px 16px rgba(20,26,40,.06)`; dialogs and the selection toolbar `0 12px 40px rgba(20,26,40,.22)` |

### 8.3 Typography

- Font stack: `Inter, system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif`; monospace for key caps.
- Base 14 px, line height 1.45, antialiased, `font-variant-numeric: tabular-nums` everywhere.
- Sizes: app title 18 px, block title 17 px, dialog title 17 px, category
  title 15 px, statistics heading 15 px, section labels 11 px uppercase with
  0.06 em letter spacing, hints and secondary text 11.5–12.5 px, badges 11 px
  uppercase.
- Weights: 400 for entry names, 500 for buttons, 600 for headings and master
  values, 700 for totals and category titles. Category titles stay in the
  primary text color; only the thin accent bar of the header carries the
  block color. Derived values use the muted text color.
- Text selection and the default cursor are disabled on the page (desktop
  application feel); inputs and error texts remain selectable.

### 8.4 Color tokens

Light mode (default) and dark mode (applied as `data-theme="dark"` on the
root element; "system" is resolved to one of the two in the frontend):

| Token | Light | Dark | Use |
|---|---|---|---|
| `--bg` | `#f5f6fa` | `#12151c` | page background |
| `--surface` | `#ffffff` | `#1b1f2a` | cards, rows, dialogs, selection toolbar |
| `--surface-2` | `#f8f9fc` | `#202533` | hover, search box, tiles |
| `--surface-3` | `#eef0f5` | `#2a3040` | category headers, pressed state, paused badge |
| `--border` | `#dfe3ea` | `#2f3646` | dividers |
| `--border-strong` | `#c8cdd8` | `#3d4557` | inputs, buttons, selection toolbar |
| `--text` | `#1c2130` | `#e8ebf2` | primary text |
| `--text-2` | `#5b6474` | `#aab2c3` | secondary text |
| `--text-3` | `#8a93a5` | `#7c8598` | muted text, icons, derived values |
| `--accent` / `-soft` / `-strong` | `#2f5fd6` / `#e6edfb` / `#1f47ad` | `#5b8def` / `#1f2c47` / `#7fa6f5` | primary buttons, focus, drop indicators, monthly badge, selected rows |
| `--income` / `-soft` / `-strong` | `#1e8a5a` / `#e5f4ec` / `#166b45` | `#3cbf82` / `#17302a` / `#5ed39c` | income accent bar and button, positive values |
| `--spending` / `-soft` / `-strong` | `#c2542c` / `#fbeae3` / `#9c4222` | `#e07a55` / `#3a241c` / `#f0956f` | spending accent bar and button, share bars, due bars |
| `--positive` / `--negative` | `#1e8a5a` / `#c9302c` | `#3cbf82` / `#ef6b6b` | signed saldo values |
| `--danger` / `-soft` | `#c9302c` / `#fbe7e6` | `#ef6b6b` / `#3a1f21` | delete buttons, error banner and toast |
| `--warn` / `-soft` | `#8a5a10` / `#fff4dc` | `#e0b25c` / `#3a2f16` | savings goal warning |
| `--savings` / `-soft` | `#6b3fa0` / `#f3ecfb` | `#a98ae6` / `#2b2340` | savings tile, timeline line, yearly badge |
| `--quarterly` / `-soft` | `#1b6f7d` / `#e6f4f6` | `#5fc0cf` / `#172e33` | quarterly badge |
| `--halfyearly` / `-soft` | `#8a5a10` / `#fbf0e0` | `#e0b25c` / `#3a2f16` | half-yearly badge |
| `--stripe` | `rgba(20,26,40,.07)` | `rgba(255,255,255,.07)` | stripes of paused rows |
| `--toast-bg` | `#1c2130` | `#2a3040` | info toasts |

Print forces the light palette with pure white surfaces and black text.

### 8.5 Chart palette

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
(55 % of the column) and a 2 px step line.

### 8.6 Components

- **Buttons**: padding 6 px top / 8 px bottom / 12 px sides (small: 3 / 5 /
  9 px, 13 px text), line height 1.2 (see 7.9 for the asymmetry), 1 px
  strong border, white surface, hover darkens the surface; primary in
  accent, income and spending buttons in their color, danger in red; icon
  buttons are 28 × 28 px transparent squares with a surface-3 background on
  hover.
- **Inputs and selects**: 8 × 10 px padding (selects 30 px on the right so
  the text keeps the same distance from the native arrow as from the left
  edge; small selects 4 × 8 px with 28 px right), strong border, accent
  border on focus, 2 px accent outline for keyboard focus; the amount input
  carries a trailing € sign; the period choice is a segmented control
  (accent-soft background for the active segment).
- **Badges**: pill, 11 px uppercase, semantic color pairs (monthly = accent,
  quarterly, half-yearly, yearly = savings). A row never shows more than one
  badge, so the frequency column never wraps.
- **Category header**: surface-3 background with a 3 px left accent bar in
  the block color (the left padding is reduced by 3 px so the columns stay
  aligned with the rows), title 15 px bold in the primary text color,
  chevron that rotates 90° when expanded, count pill in secondary text,
  share bar with a 24 px gap to the due column, subtotals right-aligned in
  secondary text, actions on hover. The bar is the only colored element.
- **Entry row**: separated by 1 px borders, hover surface-2, actions on
  hover, dragged rows at 35 % opacity, selected rows in accent-soft. Paused
  rows: `repeating-linear-gradient(135deg, var(--stripe) 0 3px, transparent
  3px 10px)` over the normal background; name, amounts and due month in
  muted text; the period badge in surface-3 with muted text.
- **Due column**: text centered, so it sits midway between the period badge
  and the right-aligned amounts.
- **Header checkbox**: native checkbox centered in the 28 px handle column
  of the column-title row, at 55 % opacity until selection mode is on,
  indeterminate while only some entries of the block are selected.
- **Selection toolbar**: fixed at the bottom center, 20 px from the bottom,
  sized by its content and wrapping before it could exceed the window width
  minus 40 px; surface background, strong border, large shadow, 10 px
  radius; the count as an accent-soft pill; slides up over 160 ms. The page
  gets 96 px bottom padding while it is visible.
- **Drop indicators**: 3 px accent line between rows and 3 px line in the
  block color between categories, both absolutely positioned overlays; a
  target category gets an accent border and a soft accent ring; empty and
  collapsed targets show a soft accent area with "Drop here".
- **Statistics tiles**: surface-2 with a 4 px left border in accent (bank)
  or savings color; value 22 px semibold.
- **Levers list**: numbered rows (rank in muted bold 11 px), name and
  category stacked, yearly amount and share stacked right; hover surface-3.
- **Banners**: 10 × 14 px padding, icon plus text, danger or warn colors.
- **Dialogs**: centered panel on a dark translucent backdrop
  (`rgba(10,13,20,.55)`), 12 px radius, title row with close icon, body,
  footer with right-aligned buttons; short fade and pop-in animation.
- **Command palette**: dialog with the search input on top, results list
  (max 360 px, scrolls) grouped by uppercase section labels, active item in
  accent-soft, hint line with the key legend.
- **Toasts**: lower right, dark surface with white text, success in
  income-strong, error in danger, optional outlined action button, close
  icon; slide-in animation; toasts with an action have a 3 px white
  (70 % opacity) countdown bar along the bottom edge.
- **Key caps** in the shortcut list: bordered inline blocks in the monospace
  font.

### 8.7 Interaction details

- Actions in rows and category headers are hidden until hover to keep the
  table calm; hovering never changes the handle cell.
- The current month is printed bold in the timeline axis; the readout below
  the timeline always uses two fixed lines so hovering never reflows the box.
- Every money value and percentage has an explanatory tooltip (see 3.3).
- Focus is visible everywhere (2 px accent outline).

---

## 9. Working agreement

- Read the whole specification before starting. Ask open questions first;
  once implementation has started, do not interrupt with questions until it
  is finished.
- Keep this document in sync with the implementation: every new or changed
  behaviour is specified here first or documented here afterwards.
- Build reusable components and clean data structures; comment where it
  helps; keep tests that cover this specification.
- Do not install anything without asking.
- Commit in meaningful stages on `main`; every commit builds and passes the
  tests.
