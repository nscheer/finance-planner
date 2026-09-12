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
} from "../../bindings/finance-planner/planner";
import type { MessageKey } from "../i18n";

import { hasKey, type Params } from "../i18n";
import { t, plural, applyLocale } from "./i18n.svelte";

export { Service, Kind, Period, ImportMode };
export type { State, CategoryView, EntryView, ImportPreview, Stats };

// ---- dialogs ---------------------------------------------------------------

/** Every user input happens in a modal; this union describes the open one. */
export type Dialog =
  | { type: "entry"; kind: Kind; entry?: EntryView; categoryId?: string }
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
  | { type: "import"; preview: ImportPreview };

export interface Toast {
  id: number;
  kind: "success" | "info" | "error";
  message: string;
}

// ---- state -------------------------------------------------------------------

export const app = $state({
  state: null as State | null,
  loadError: "" as string,
  busy: false,
  dialog: null as Dialog | null,
  toasts: [] as Toast[],
});

let toastSeq = 0;

/** Shows a short popup notification. Errors stay longer. */
export function notify(kind: Toast["kind"], message: string): void {
  const id = ++toastSeq;
  app.toasts.push({ id, kind, message });
  setTimeout(() => dismissToast(id), kind === "error" ? 8000 : 3500);
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
    // Apply the saved language before rendering the data.
    applyLocale(state.settings?.language);
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

export function setCollapsed(categoryId: string, collapsed: boolean): Promise<boolean> {
  return applyOrAlert(Service.SetCategoryCollapsed(categoryId, collapsed));
}

/** "Expand all" / "collapse all" for one block (income or spending). */
export function setAllCollapsed(kind: Kind, collapsed: boolean): Promise<boolean> {
  return applyOrAlert(Service.SetAllCollapsed(kind, collapsed));
}

export function confirmDeleteEntry(entry: EntryView): void {
  openDialog({
    type: "confirm",
    title: t("confirm.deleteEntry.title"),
    message: t("confirm.deleteEntry.message", { name: entry.name }),
    confirmLabel: t("dialog.delete"),
    onConfirm: async () => {
      if (await applyOrAlert(Service.DeleteEntry(entry.id), "alert.deleteFailed")) {
        notify("success", t("toast.entryDeleted", { name: entry.name }));
      }
    },
  });
}

export function confirmDeleteCategory(category: CategoryView): void {
  const count = category.entries?.length ?? 0;
  if (count > 0) {
    alert(t("confirm.categoryInUse.title"), plural("confirm.categoryInUse.message", count, { name: category.name }));
    return;
  }
  openDialog({
    type: "confirm",
    title: t("confirm.deleteCategory.title"),
    message: t("confirm.deleteCategory.message", { name: category.name }),
    confirmLabel: t("dialog.delete"),
    onConfirm: async () => {
      if (await applyOrAlert(Service.DeleteCategory(category.id), "alert.deleteFailed")) {
        notify("success", t("toast.categoryDeleted", { name: category.name }));
      }
    },
  });
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

export async function importData(path: string, mode: ImportMode): Promise<void> {
  if (await applyOrAlert(Service.ImportData(path, mode), "alert.importFailed")) {
    notify("success", t(mode === ImportMode.ImportReplace ? "toast.importedReplace" : "toast.importedMerge"));
  }
}
