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

/** Extracts a readable message from an error thrown by a binding call. */
export function errorMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  if (typeof err === "string") return err;
  return String(err);
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
export async function applyOrAlert(call: Promise<State>, title = "Something went wrong"): Promise<boolean> {
  try {
    await apply(call);
    return true;
  } catch (err) {
    alert(title, errorMessage(err));
    return false;
  }
}

export async function loadState(): Promise<void> {
  try {
    app.state = await Service.GetState();
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
  return kind === Kind.KindIncome ? "Income" : "Spending";
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
    title: "Delete entry",
    message: `Delete "${entry.name}"? This can't be undone.`,
    confirmLabel: "Delete",
    onConfirm: async () => {
      if (await applyOrAlert(Service.DeleteEntry(entry.id), "Delete failed")) {
        notify("success", `Deleted "${entry.name}".`);
      }
    },
  });
}

export function confirmDeleteCategory(category: CategoryView): void {
  const count = category.entries?.length ?? 0;
  if (count > 0) {
    alert(
      "Category in use",
      `"${category.name}" still contains ${count} ${count === 1 ? "entry" : "entries"}. ` +
        "Move or delete them first, then delete the category.",
    );
    return;
  }
  openDialog({
    type: "confirm",
    title: "Delete category",
    message: `Delete the category "${category.name}"?`,
    confirmLabel: "Delete",
    onConfirm: async () => {
      if (await applyOrAlert(Service.DeleteCategory(category.id), "Delete failed")) {
        notify("success", `Deleted category "${category.name}".`);
      }
    },
  });
}

export async function exportData(): Promise<void> {
  try {
    const path = await Service.ExportData();
    if (path) notify("success", `Exported to ${path}`);
  } catch (err) {
    alert("Export failed", errorMessage(err));
  }
}

export async function startImport(): Promise<void> {
  try {
    const preview = await Service.ChooseImportFile();
    if (!preview.path) return; // cancelled
    openDialog({ type: "import", preview });
  } catch (err) {
    alert("Import failed", errorMessage(err));
  }
}

export async function importData(path: string, mode: ImportMode): Promise<void> {
  if (await applyOrAlert(Service.ImportData(path, mode), "Import failed")) {
    notify("success", mode === ImportMode.ImportReplace ? "Data replaced by import." : "Imported data added.");
  }
}
