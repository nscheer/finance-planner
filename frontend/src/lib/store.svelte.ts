/**
 * Application state. The Go service is the single source of truth: every
 * mutating call returns the complete new State, which replaces `app.state`.
 * This keeps the frontend free of duplicated business logic.
 */
import {
  Service,
  Kind,
  Period,
  ImportMode,
  type State,
  type CategoryView,
  type EntryView,
  type ImportPreview,
  type Stats,
  type BackupInfo,
} from "../../bindings/finance-planner/planner";
import type { MessageKey } from "../i18n";

import { hasKey, type Params } from "../i18n";
import { t, plural, applyLocale, currentLanguage } from "./i18n.svelte";
import { applyTheme } from "./theme.svelte";

export { Service, Kind, Period, ImportMode };
export type { State, CategoryView, EntryView, ImportPreview, Stats, BackupInfo };

// ---- dialogs ---------------------------------------------------------------

/** Every user input happens in a modal; this union describes the open one. */
export type Dialog =
  /** entry: edit this entry; duplicateOf: prefill from this entry but create a new one. */
  | { type: "entry"; kind: Kind; entry?: EntryView; categoryId?: string; duplicateOf?: EntryView }
  /** returnToEntry: reopen the "new entry" dialog after the category was added. */
  | { type: "category"; kind: Kind; category?: CategoryView; returnToEntry?: boolean }
  | {
      type: "confirm";
      title: string;
      message: string;
      confirmLabel: string;
      onConfirm: () => Promise<void> | void;
    }
  | { type: "alert"; title: string; message: string }
  | { type: "goal" }
  | { type: "backups" }
  | { type: "shortcuts" }
  | { type: "import"; preview: ImportPreview };

export interface Toast {
  id: number;
  kind: "success" | "info" | "error";
  message: string;
  /** Optional action button, e.g. "Undo". */
  action?: { label: string; run: () => void | Promise<void> };
  /** How long the toast stays, in milliseconds (drives the countdown bar). */
  durationMs: number;
}

/** Period filter values of the search bar. */
export type PeriodFilter = "all" | Period | "paused";

// ---- state -------------------------------------------------------------------

export const app = $state({
  state: null as State | null,
  loadError: "" as string,
  busy: false,
  dialog: null as Dialog | null,
  toasts: [] as Toast[],
  /** Search text and period filter of the top bar. */
  filter: { query: "", period: "all" as PeriodFilter },
  /** True while the print layout is active (all categories expanded). */
  printing: false,
  /** Ids of the selected entries (multi-select for bulk actions). */
  selection: {} as Record<string, true>,
  /** Last entry selected by click; anchor for Shift+click ranges. */
  selectionAnchor: null as string | null,
});

let toastSeq = 0;

/**
 * Shows a short popup notification. Errors and toasts with an action stay
 * longer so the user has time to react.
 */
export function notify(kind: Toast["kind"], message: string, action?: Toast["action"]): void {
  const id = ++toastSeq;
  const durationMs = kind === "error" || action ? 9000 : 3500;
  app.toasts.push({ id, kind, message, action, durationMs });
  setTimeout(() => dismissToast(id), durationMs);
}

// ---- search / filter ----------------------------------------------------

export function filterActive(): boolean {
  return app.filter.query.trim() !== "" || app.filter.period !== "all";
}

export function clearFilter(): void {
  app.filter.query = "";
  app.filter.period = "all";
}

/** Drops selected ids that no longer exist (after deletes, imports, restores). */
export function pruneSelection(): void {
  const existing = new Set<string>();
  for (const kind of [Kind.KindIncome, Kind.KindSpending]) {
    for (const c of categoriesOf(kind)) for (const e of c.entries ?? []) existing.add(e.id);
  }
  for (const id of Object.keys(app.selection)) if (!existing.has(id)) delete app.selection[id];
}

function entryMatches(e: EntryView): boolean {
  const q = app.filter.query.trim().toLowerCase();
  if (q && !e.name.toLowerCase().includes(q) && !(e.notes ?? "").toLowerCase().includes(q)) return false;
  const p = app.filter.period;
  if (p === "paused") return !!e.paused;
  if (p !== "all" && e.period !== p) return false;
  return true;
}

/**
 * Categories of a kind as shown in the table: with an active filter only
 * the matching entries are listed and categories without matches are
 * hidden. Subtotals stay those of the whole category.
 */
export function visibleCategories(kind: Kind): CategoryView[] {
  const all = categoriesOf(kind);
  if (!filterActive()) return all;
  return all
    .map((c) => ({ ...c, entries: (c.entries ?? []).filter(entryMatches) }))
    .filter((c) => (c.entries?.length ?? 0) > 0);
}

/** Count of entries shown vs. all entries (for the filter hint). */
export function filterCounts(): { shown: number; total: number } {
  let shown = 0;
  let total = 0;
  for (const kind of [Kind.KindIncome, Kind.KindSpending]) {
    for (const c of categoriesOf(kind)) {
      for (const e of c.entries ?? []) {
        total++;
        if (entryMatches(e)) shown++;
      }
    }
  }
  return { shown, total };
}

export function dismissToast(id: number): void {
  app.toasts = app.toasts.filter((t) => t.id !== id);
}

export function openDialog(dialog: Dialog): void {
  app.dialog = dialog;
}

export function closeDialog(): void {
  app.dialog = null;
}

/** Shows an error in a modal (for things that are not tied to a form). */
export function alert(title: string, message: string): void {
  app.dialog = { type: "alert", title, message };
}

/**
 * Turns an error thrown by a binding call into a message in the current
 * language. Coded backend errors (planner/errors.go) arrive as `cause`
 * = {code, params} and are looked up under "errors.<code>"; anything else
 * is shown as-is.
 */
export function errorMessage(err: unknown): string {
  const cause = (err as { cause?: unknown } | null)?.cause;
  if (cause && typeof cause === "object" && "code" in cause) {
    const { code, params = {} } = cause as { code: string; params?: Record<string, unknown> };
    return translateError(code, params);
  }
  if (err instanceof Error) return err.message;
  if (typeof err === "string") return err;
  return String(err);
}

function translateError(code: string, raw: Record<string, unknown>): string {
  const params: Params = {};
  for (const [k, v] of Object.entries(raw)) {
    // Kinds are shown in the user's language, e.g. "spending" -> "Ausgaben".
    if ((k === "kind" || k === "from" || k === "to") && hasKey(`kindInline.${v}`)) {
      params[k] = t(`kindInline.${v}` as "kindInline.income" | "kindInline.spending");
    } else {
      params[k] = typeof v === "number" ? v : String(v);
    }
  }
  const key = `errors.${code}`;
  if (hasKey(`${key}.one`) && typeof params.count === "number") {
    return plural(key as "errors.category.inUse", params.count, params);
  }
  if (hasKey(key)) return t(key, params);
  return code;
}

// ---- service calls -------------------------------------------------------------

/**
 * Runs a service call that returns the new state. Throws on error so callers
 * inside dialogs can display the message inline.
 */
export async function apply(call: Promise<State>): Promise<State> {
  app.busy = true;
  try {
    const next = await call;
    app.state = next;
    pruneSelection();
    return next;
  } finally {
    app.busy = false;
  }
}

/** Like apply(), but reports errors in an alert modal instead of throwing. */
export async function applyOrAlert(call: Promise<State>, titleKey: MessageKey = "alert.generic"): Promise<boolean> {
  try {
    await apply(call);
    return true;
  } catch (err) {
    alert(t(titleKey), errorMessage(err));
    return false;
  }
}

export async function loadState(): Promise<void> {
  try {
    const state = await Service.GetState();
    // Apply the saved language and theme before rendering the data.
    applyLocale(state.settings?.language);
    applyTheme(state.settings?.theme);
    app.state = state;
    app.loadError = "";
  } catch (err) {
    app.loadError = errorMessage(err);
  }
}

// ---- derived helpers ------------------------------------------------------------

export function categoriesOf(kind: Kind): CategoryView[] {
  if (!app.state) return [];
  return (kind === Kind.KindIncome ? app.state.income : app.state.spending) ?? [];
}

export function kindLabel(kind: Kind): string {
  return t(kind === Kind.KindIncome ? "kind.income" : "kind.spending");
}

/** Helper for keys that exist per kind, e.g. "block.empty" -> "block.empty.income". */
export function kindKey<T extends string>(base: T, kind: Kind): `${T}.income` | `${T}.spending` {
  return `${base}.${kind === Kind.KindIncome ? "income" : "spending"}`;
}

// ---- actions used by several components ----------------------------------------

/** Translation key of a period badge. */
export function periodKey(period: Period): MessageKey {
  switch (period) {
    case Period.PeriodQuarterly:
      return "entry.quarterly";
    case Period.PeriodHalfYearly:
      return "entry.halfyearly";
    case Period.PeriodYearly:
      return "entry.yearly";
    default:
      return "entry.monthly";
  }
}

/** Months between two payments of a period (1, 3, 6, 12). */
export function periodMonths(period: Period): number {
  switch (period) {
    case Period.PeriodQuarterly:
      return 3;
    case Period.PeriodHalfYearly:
      return 6;
    case Period.PeriodYearly:
      return 12;
    default:
      return 1;
  }
}

export async function pauseEntry(entry: EntryView, paused: boolean): Promise<void> {
  if (await applyOrAlert(Service.SetEntryPaused(entry.id, paused))) {
    notify("info", t(paused ? "toast.entryPaused" : "toast.entryResumed", { name: entry.name }));
  }
}

export function setCollapsed(categoryId: string, collapsed: boolean): Promise<boolean> {
  return applyOrAlert(Service.SetCategoryCollapsed(categoryId, collapsed));
}

/** "Expand all" / "collapse all" for one block (income or spending). */
export function setAllCollapsed(kind: Kind, collapsed: boolean): Promise<boolean> {
  return applyOrAlert(Service.SetAllCollapsed(kind, collapsed));
}

export function confirmDeleteEntry(entry: EntryView, index: number): void {
  // Snapshot for the undo action: the view object may be gone after the delete.
  const restore = {
    id: entry.id,
    categoryId: entry.categoryId,
    name: entry.name,
    amountCents: entry.amountCents,
    period: entry.period,
    dueMonth: entry.dueMonth ?? 0,
    paused: entry.paused ?? false,
    notes: entry.notes ?? "",
  };
  openDialog({
    type: "confirm",
    title: t("confirm.deleteEntry.title"),
    message: t("confirm.deleteEntry.message", { name: entry.name }),
    confirmLabel: t("dialog.delete"),
    onConfirm: async () => {
      if (await applyOrAlert(Service.DeleteEntry(entry.id), "alert.deleteFailed")) {
        notify("success", t("toast.entryDeleted", { name: entry.name }), {
          label: t("toast.undo"),
          run: async () => {
            if (await applyOrAlert(Service.RestoreEntry(restore, index))) notify("info", t("toast.restored"));
          },
        });
      }
    },
  });
}

export function confirmDeleteCategory(category: CategoryView, index: number): void {
  const count = category.entries?.length ?? 0;
  if (count > 0) {
    alert(t("confirm.categoryInUse.title"), plural("confirm.categoryInUse.message", count, { name: category.name }));
    return;
  }
  const restore = { id: category.id, name: category.name, kind: category.kind, collapsed: category.collapsed };
  openDialog({
    type: "confirm",
    title: t("confirm.deleteCategory.title"),
    message: t("confirm.deleteCategory.message", { name: category.name }),
    confirmLabel: t("dialog.delete"),
    onConfirm: async () => {
      if (await applyOrAlert(Service.DeleteCategory(category.id), "alert.deleteFailed")) {
        notify("success", t("toast.categoryDeleted", { name: category.name }), {
          label: t("toast.undo"),
          run: async () => {
            if (await applyOrAlert(Service.RestoreCategory(restore, index))) notify("info", t("toast.restored"));
          },
        });
      }
    },
  });
}

export async function exportCSV(): Promise<void> {
  try {
    const path = await Service.ExportCSV();
    if (path) notify("success", t("toast.exportedCsv", { path }));
  } catch (err) {
    alert(t("alert.exportFailed"), errorMessage(err));
  }
}

export async function restoreBackup(path: string): Promise<boolean> {
  const ok = await applyOrAlert(Service.RestoreBackup(path), "alert.importFailed");
  if (ok) notify("success", t("toast.backupRestored"));
  return ok;
}

export function confirmLoadSampleData(): void {
  openDialog({
    type: "confirm",
    title: t("confirm.sample.title"),
    message: t("confirm.sample.message"),
    confirmLabel: t("confirm.sample.confirm"),
    onConfirm: async () => {
      try {
        const result = await Service.LoadSampleData(currentLanguage());
        app.state = result.state;
        notify(
          "success",
          t("toast.sampleLoaded", {
            categories: plural("importDialog.categories", result.categories),
            entries: plural("importDialog.entries", result.entries),
          }),
        );
      } catch (err) {
        alert(t("alert.generic"), errorMessage(err));
      }
    },
  });
}

// ---- multi-select ----------------------------------------------------------

export function isSelected(id: string): boolean {
  return app.selection[id] === true;
}

export function selectionCount(): number {
  return Object.keys(app.selection).length;
}

export function setSelected(id: string, on: boolean): void {
  if (on) app.selection[id] = true;
  else delete app.selection[id];
  app.selectionAnchor = on ? id : app.selectionAnchor;
}

export function toggleSelected(id: string): void {
  setSelected(id, !isSelected(id));
}

/** Selects all entries between the anchor and `id` inside one category. */
export function selectRange(category: CategoryView, id: string): void {
  const entries = category.entries ?? [];
  const to = entries.findIndex((e) => e.id === id);
  const from = entries.findIndex((e) => e.id === app.selectionAnchor);
  if (to < 0) return;
  if (from < 0) {
    setSelected(id, true);
    return;
  }
  const [a, b] = from < to ? [from, to] : [to, from];
  for (let i = a; i <= b; i++) app.selection[entries[i].id] = true;
}

export function clearSelection(): void {
  app.selection = {};
  app.selectionAnchor = null;
}

/** The selected entries with their category, in table order. */
export function selectedEntries(): { entry: EntryView; category: CategoryView; index: number }[] {
  const out: { entry: EntryView; category: CategoryView; index: number }[] = [];
  for (const kind of [Kind.KindIncome, Kind.KindSpending]) {
    for (const category of categoriesOf(kind)) {
      (category.entries ?? []).forEach((entry, index) => {
        if (isSelected(entry.id)) out.push({ entry, category, index });
      });
    }
  }
  return out;
}

/** Kind of the selection: one kind, or "mixed" when both kinds are selected. */
export function selectionKind(): Kind | "mixed" | null {
  let kind: Kind | null = null;
  for (const { category } of selectedEntries()) {
    if (kind === null) kind = category.kind;
    else if (kind !== category.kind) return "mixed";
  }
  return kind;
}

function selectedIds(): string[] {
  return selectedEntries().map((s) => s.entry.id);
}

export async function moveSelection(categoryId: string): Promise<void> {
  const ids = selectedIds();
  const target = [...categoriesOf(Kind.KindIncome), ...categoriesOf(Kind.KindSpending)].find((c) => c.id === categoryId);
  if (ids.length === 0 || !target) return;
  if (await applyOrAlert(Service.MoveEntries(ids, categoryId), "alert.moveFailed")) {
    clearSelection();
    notify("success", plural("toast.entriesMoved", ids.length, { name: target.name }));
  }
}

export async function pauseSelection(paused: boolean): Promise<void> {
  const ids = selectedIds();
  if (ids.length === 0) return;
  if (await applyOrAlert(Service.SetEntriesPaused(ids, paused))) {
    clearSelection();
    notify("info", plural(paused ? "toast.entriesPaused" : "toast.entriesResumed", ids.length));
  }
}

export function confirmDeleteSelection(): void {
  const items = selectedEntries();
  if (items.length === 0) return;
  // Snapshot for the undo action.
  const restore = items.map(({ entry, index }) => ({
    entry: {
      id: entry.id,
      categoryId: entry.categoryId,
      name: entry.name,
      amountCents: entry.amountCents,
      period: entry.period,
      dueMonth: entry.dueMonth ?? 0,
      paused: entry.paused ?? false,
      notes: entry.notes ?? "",
    },
    index,
  }));
  const ids = items.map((s) => s.entry.id);
  openDialog({
    type: "confirm",
    title: t("confirm.deleteEntries.title"),
    message: plural("confirm.deleteEntries.message", ids.length),
    confirmLabel: t("dialog.delete"),
    onConfirm: async () => {
      if (await applyOrAlert(Service.DeleteEntries(ids), "alert.deleteFailed")) {
        clearSelection();
        notify("success", plural("toast.entriesDeleted", ids.length), {
          label: t("toast.undo"),
          run: async () => {
            if (await applyOrAlert(Service.RestoreEntries(restore))) notify("info", t("toast.restored"));
          },
        });
      }
    },
  });
}

/**
 * Opens the system print dialog. While printing, every category is rendered
 * expanded (see CategoryGroup) and the print stylesheet lays the page out in
 * one column.
 */
export function printPlanner(): void {
  if (app.printing) return;
  app.printing = true;
  const done = () => {
    app.printing = false;
    window.removeEventListener("afterprint", done);
  };
  window.addEventListener("afterprint", done);
  // Let Svelte render the expanded categories before the dialog opens.
  setTimeout(() => window.print(), 50);
}

/** Whether the planner has no categories and no entries at all. */
export function isEmpty(): boolean {
  return categoriesOf(Kind.KindIncome).length === 0 && categoriesOf(Kind.KindSpending).length === 0;
}

export async function exportData(): Promise<void> {
  try {
    const path = await Service.ExportData();
    if (path) notify("success", t("toast.exported", { path }));
  } catch (err) {
    alert(t("alert.exportFailed"), errorMessage(err));
  }
}

export async function startImport(): Promise<void> {
  try {
    const preview = await Service.ChooseImportFile();
    if (!preview.path) return; // cancelled
    openDialog({ type: "import", preview });
  } catch (err) {
    alert(t("alert.importFailed"), errorMessage(err));
  }
}

/** Toast action that restores the backup written before an import. */
function undoImport(backupPath: string): Toast["action"] | undefined {
  if (!backupPath) return undefined;
  return {
    label: t("toast.undoImport"),
    run: async () => {
      await restoreBackup(backupPath);
    },
  };
}

export async function importData(path: string, mode: ImportMode): Promise<void> {
  app.busy = true;
  try {
    const result = await Service.ImportData(path, mode);
    app.state = result.state;
    if (mode === ImportMode.ImportReplace) {
      notify(
        "success",
        t("toast.importedReplace", {
          entries: plural("importDialog.entries", result.entriesAdded),
          categories: plural("importDialog.categories", result.categoriesAdded),
        }),
        undoImport(result.backupPath),
      );
    } else {
      notify(
        "success",
        t("toast.importedMerge", {
          entries: plural("toast.importedMerge.entries", result.entriesAdded),
          skipped: plural("toast.importedMerge.skipped", result.entriesSkipped),
          categories: plural("toast.importedMerge.categories", result.categoriesAdded),
        }),
        undoImport(result.backupPath),
      );
    }
  } catch (err) {
    alert(t("alert.importFailed"), errorMessage(err));
  } finally {
    app.busy = false;
  }
}
