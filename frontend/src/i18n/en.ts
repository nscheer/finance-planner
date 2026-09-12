/**
 * English messages. This file defines the set of translation keys: every
 * other language must provide exactly these keys (enforced by TypeScript via
 * the Messages type in index.ts and by the i18n unit test).
 *
 * Placeholders are written as {name}. Keys ending in ".one"/".other" are
 * plural forms selected by the `plural()` helper.
 */
export const en = {
  // ---- application shell ----
  "app.title": "Finance Planner",
  "app.language": "Language",
  "app.loading": "Loading…",
  "app.loadError": "Could not load the data file.",
  "app.retry": "Retry",
  "app.import": "Import",
  "app.export": "Export",

  // ---- kinds ----
  "kind.income": "Income",
  "kind.spending": "Spending",
  // used inside sentences ("a spending category")
  "kindInline.income": "income",
  "kindInline.spending": "spending",

  // ---- blocks ----
  "block.expandAll": "Expand all",
  "block.collapseAll": "Collapse all",
  "block.addCategory": "Category",
  "block.perMonth": "/ month",
  "block.perYear": "/ year",
  "block.column.name": "Name",
  "block.column.entered": "Entered",
  "block.column.perMonth": "Per month",
  "block.column.perYear": "Per year",
  "block.empty.income": "No income categories yet. Add a category first, then add entries to it.",
  "block.empty.spending": "No spending categories yet. Add a category first, then add entries to it.",

  // ---- categories ----
  "category.dragHint": "Drag to reorder categories",
  "category.addEntry": "Add entry",
  "category.rename": "Rename",
  "category.delete": "Delete",
  "category.inUse": "Category is in use",
  "category.noEntries": "No entries yet",
  "category.dropHere": "Drop here",
  "category.dropInto": "Drop to add to \"{name}\"",

  // ---- entries ----
  "entry.dragHint": "Drag to reorder or move to another category",
  "entry.edit": "Edit",
  "entry.delete": "Delete",
  "entry.monthly": "monthly",
  "entry.yearly": "yearly",

  // ---- statistics ----
  "stats.title": "Statistics",
  "stats.transfers": "Monthly transfers",
  "stats.toBank": "To bank account",
  "stats.toBankSub": "spendings paid per month",
  "stats.toSavings": "To savings account",
  "stats.toSavingsSub": "1/12 of spendings paid per year",
  "stats.overview": "Overview",
  "stats.incomePerMonth": "Income per month",
  "stats.avgCostPerMonth": "Average cost per month",
  "stats.saldoPerMonth": "Saldo per month",
  "stats.incomePerYear": "Income per year",
  "stats.costPerYear": "Cost per year",
  "stats.saldoPerYear": "Saldo per year",
  "stats.note":
    "Monthly spendings are paid from the bank account. Yearly spendings are saved up month by month on the savings account, so the money is available when they are due.",

  // ---- dialogs (shared) ----
  "dialog.cancel": "Cancel",
  "dialog.ok": "OK",
  "dialog.close": "Close",
  "dialog.save": "Save",
  "dialog.add": "Add",
  "dialog.delete": "Delete",
  "dialog.dismiss": "Dismiss",

  // ---- category dialog ----
  "categoryDialog.titleNew.income": "New income category",
  "categoryDialog.titleNew.spending": "New spending category",
  "categoryDialog.titleRename": "Rename category",
  "categoryDialog.name": "Name",
  "categoryDialog.namePlaceholder": "e.g. Housing",
  "categoryDialog.submit": "Add category",

  // ---- entry dialog ----
  "entryDialog.titleNew.income": "New income",
  "entryDialog.titleNew.spending": "New spending",
  "entryDialog.titleEdit.income": "Edit income",
  "entryDialog.titleEdit.spending": "Edit spending",
  "entryDialog.noCategories.income":
    "There are no income categories yet. Categories have to be added before entries can be entered.",
  "entryDialog.noCategories.spending":
    "There are no spending categories yet. Categories have to be added before entries can be entered.",
  "entryDialog.addCategoryFirst": "Add a category first",
  "entryDialog.name": "Name",
  "entryDialog.namePlaceholder": "e.g. Rent",
  "entryDialog.amount": "Amount",
  "entryDialog.paid": "Paid",
  "entryDialog.perMonth": "per month",
  "entryDialog.perYear": "per year",
  "entryDialog.category": "Category",
  "entryDialog.previewYear": "= {amount} per year",
  "entryDialog.previewMonth": "= {amount} per month",
  "entryDialog.invalidAmount": "Please enter a valid amount, e.g. 12.50.",

  // ---- confirmations ----
  "confirm.deleteEntry.title": "Delete entry",
  "confirm.deleteEntry.message": "Delete \"{name}\"? This can't be undone.",
  "confirm.deleteCategory.title": "Delete category",
  "confirm.deleteCategory.message": "Delete the category \"{name}\"?",
  "confirm.categoryInUse.title": "Category in use",
  "confirm.categoryInUse.message.one":
    "\"{name}\" still contains {count} entry. Move or delete it first, then delete the category.",
  "confirm.categoryInUse.message.other":
    "\"{name}\" still contains {count} entries. Move or delete them first, then delete the category.",

  // ---- import dialog ----
  "importDialog.title": "Import data",
  "importDialog.summary": "{file} contains {categories} and {entries} (file version {version}).",
  "importDialog.categories.one": "{count} category",
  "importDialog.categories.other": "{count} categories",
  "importDialog.entries.one": "{count} entry",
  "importDialog.entries.other": "{count} entries",
  "importDialog.question": "How should the data be imported?",
  "importDialog.merge.title": "Add to current data",
  "importDialog.merge.description": "Keeps everything you have. Categories with the same name are reused.",
  "importDialog.replace.title": "Replace current data",
  "importDialog.replace.description":
    "Deletes all current categories and entries and uses the imported file instead.",

  // ---- notifications ----
  "toast.entryAdded": "Added \"{name}\".",
  "toast.entrySaved": "Saved \"{name}\".",
  "toast.entryDeleted": "Deleted \"{name}\".",
  "toast.categoryAdded.income": "Added income category \"{name}\".",
  "toast.categoryAdded.spending": "Added spending category \"{name}\".",
  "toast.categoryRenamed": "Renamed category to \"{name}\".",
  "toast.categoryDeleted": "Deleted category \"{name}\".",
  "toast.exported": "Exported to {path}",
  "toast.importedMerge": "Imported data added.",
  "toast.importedReplace": "Data replaced by import.",

  // ---- error dialog titles ----
  "alert.generic": "Something went wrong",
  "alert.deleteFailed": "Delete failed",
  "alert.moveFailed": "Move failed",
  "alert.exportFailed": "Export failed",
  "alert.importFailed": "Import failed",

  // ---- backend error codes (planner/errors.go) ----
  "errors.kind.unknown": "Unknown kind \"{kind}\".",
  "errors.period.unknown": "Unknown period \"{period}\".",
  "errors.category.notFound": "The category was not found.",
  "errors.category.nameEmpty": "The category name must not be empty.",
  "errors.category.exists": "A {kind} category named \"{name}\" already exists.",
  "errors.category.inUse.one": "The category \"{name}\" is still used by {count} entry and can't be deleted.",
  "errors.category.inUse.other": "The category \"{name}\" is still used by {count} entries and can't be deleted.",
  "errors.entry.notFound": "The entry was not found.",
  "errors.entry.categoryRequired": "Please choose a category.",
  "errors.entry.nameEmpty": "The name must not be empty.",
  "errors.entry.amountPositive": "The amount must be greater than 0.",
  "errors.entry.kindMismatch": "An {from} entry can't be moved into a {to} category.",
  "errors.import.modeUnknown": "Unknown import mode \"{mode}\".",
  "errors.import.invalidFile": "{file} is not a valid planner file: {detail}",
  "errors.language.invalid": "Unsupported language \"{language}\".",
};
