/**
 * Shared drag & drop state. Native HTML5 drag events are used; this module
 * only tracks what is being dragged and where it would land, so that the
 * components can render a clear drop indicator.
 *
 * Two things can be dragged:
 *  - a category (reordered among the categories of the same kind)
 *  - an entry (reordered within its category or moved to another category
 *    of the same kind)
 */
import { Service, Kind, type CategoryView, type EntryView } from "../../bindings/finance-planner/planner";
import { applyOrAlert, filterActive } from "./store.svelte";
import { resolveMoveIndex } from "./reorder";

export type DragSource =
  | { type: "category"; id: string; kind: Kind; index: number }
  | { type: "entry"; id: string; kind: Kind; categoryId: string; index: number };

export type DropTarget =
  /** Insert the dragged category before the category at `index` of `kind`. */
  | { type: "category"; kind: Kind; index: number }
  /** Insert the dragged entry before the entry at `index` of the category. */
  | { type: "entry"; categoryId: string; index: number };

export const dnd = $state({
  source: null as DragSource | null,
  target: null as DropTarget | null,
});

export function startDrag(event: DragEvent, source: DragSource): void {
  // With a filter active the visible indices don't match the stored order.
  if (filterActive()) {
    event.preventDefault();
    return;
  }
  dnd.source = source;
  dnd.target = null;
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = "move";
    // Some engines need data for the drag to start at all.
    event.dataTransfer.setData("text/plain", source.id);
  }
}

export function endDrag(): void {
  dnd.source = null;
  dnd.target = null;
}

/** Whether the pointer is in the lower half of the element. */
export function isLowerHalf(event: DragEvent, el: HTMLElement): boolean {
  const rect = el.getBoundingClientRect();
  return event.clientY - rect.top > rect.height / 2;
}

/** Checks if the dragged item may be dropped into the given category. */
export function canDropEntryInto(category: CategoryView): boolean {
  return dnd.source?.type === "entry" && dnd.source.kind === category.kind;
}

export function canDropCategoryIn(kind: Kind): boolean {
  return dnd.source?.type === "category" && dnd.source.kind === kind;
}

/** Sets the target and allows the drop (preventDefault) if compatible. */
export function hoverEntryTarget(event: DragEvent, category: CategoryView, index: number): void {
  if (!canDropEntryInto(category)) return;
  event.preventDefault();
  if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
  const t = dnd.target;
  if (!t || t.type !== "entry" || t.categoryId !== category.id || t.index !== index) {
    dnd.target = { type: "entry", categoryId: category.id, index };
  }
}

export function hoverCategoryTarget(event: DragEvent, kind: Kind, index: number): void {
  if (!canDropCategoryIn(kind)) return;
  event.preventDefault();
  if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
  const t = dnd.target;
  if (!t || t.type !== "category" || t.kind !== kind || t.index !== index) {
    dnd.target = { type: "category", kind, index };
  }
}

/**
 * Fallback for areas inside a block that are not drop positions themselves
 * (block header, padding, gaps): keep the current target and allow the
 * drop, so the cursor never flips to "not allowed" while moving across the
 * block. Does nothing when no target has been chosen yet.
 */
export function hoverKeepTarget(event: DragEvent, kind: Kind): void {
  if (event.defaultPrevented) return;
  const source = dnd.source;
  if (!source || source.kind !== kind || !dnd.target) return;
  event.preventDefault();
  if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
}

/** Whether an indicator should be shown before entry `index` of the category. */
export function isEntryTarget(categoryId: string, index: number): boolean {
  const t = dnd.target;
  return !!t && t.type === "entry" && t.categoryId === categoryId && t.index === index;
}

export function isCategoryTarget(kind: Kind, index: number): boolean {
  const t = dnd.target;
  return !!t && t.type === "category" && t.kind === kind && t.index === index;
}

/** Whether the category is the current drop target for an entry. */
export function isCategoryHighlighted(categoryId: string): boolean {
  const t = dnd.target;
  return !!t && t.type === "entry" && t.categoryId === categoryId;
}

export function isDraggedEntry(entry: EntryView): boolean {
  return dnd.source?.type === "entry" && dnd.source.id === entry.id;
}

export function isDraggedCategory(category: CategoryView): boolean {
  return dnd.source?.type === "category" && dnd.source.id === category.id;
}

/**
 * Performs the move for the current source/target pair. The backend expects
 * the index in the list *after* the dragged item was removed, so moving an
 * item downwards inside the same list shifts the index by one.
 */
export async function drop(event: DragEvent): Promise<void> {
  event.preventDefault();
  const source = dnd.source;
  const target = dnd.target;
  endDrag();
  if (!source || !target) return;

  if (source.type === "category" && target.type === "category") {
    if (source.kind !== target.kind) return;
    const index = resolveMoveIndex(true, source.index, target.index);
    if (index === null) return;
    await applyOrAlert(Service.MoveCategory(source.id, index), "alert.moveFailed");
    return;
  }

  if (source.type === "entry" && target.type === "entry") {
    const index = resolveMoveIndex(target.categoryId === source.categoryId, source.index, target.index);
    if (index === null) return;
    await applyOrAlert(Service.MoveEntry(source.id, target.categoryId, index), "alert.moveFailed");
  }
}
