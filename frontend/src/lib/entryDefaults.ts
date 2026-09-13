/**
 * Which category a new entry should start in. Pure and testable: the rule
 * has to survive a category being deleted or the data being replaced by an
 * import, where the remembered id no longer exists.
 */
export function pickCategory(categories: { id: string }[], rememberedId?: string): string {
  if (rememberedId && categories.some((c) => c.id === rememberedId)) return rememberedId;
  return categories.length > 0 ? categories[0].id : "";
}
