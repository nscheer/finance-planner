/**
 * Index arithmetic for drag & drop, kept free of Svelte/Wails imports so it
 * can be unit tested with plain Node.
 *
 * Drop targets are expressed as "insert before item `targetIndex`" of the
 * target list, while the backend expects the position in the list *after*
 * the dragged item was removed. Both are only different when the item moves
 * downwards inside the same list.
 *
 * Returns null when the drop would not change anything.
 */
export function resolveMoveIndex(sameList: boolean, sourceIndex: number, targetIndex: number): number | null {
  let index = targetIndex;
  if (sameList) {
    if (index > sourceIndex) index -= 1;
    if (index === sourceIndex) return null;
  }
  return index;
}
