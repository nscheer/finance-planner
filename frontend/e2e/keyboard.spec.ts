/**
 * Keyboard shortcuts, the command palette and the data menu
 * (specification 3.2 and 3.9).
 */
import { test, expect } from "@playwright/test";
import { openApp, row, block, en } from "./app.ts";

test("the single-key shortcuts open what they promise", async ({ page }) => {
  await openApp(page, { sample: true });

  await page.keyboard.press("n");
  await expect(page.getByRole("dialog", { name: en["entryDialog.titleNew.spending"] })).toBeVisible();
  await page.keyboard.press("Escape");

  await page.keyboard.press("i");
  await expect(page.getByRole("dialog", { name: en["entryDialog.titleNew.income"] })).toBeVisible();
  await page.keyboard.press("Escape");

  await page.keyboard.press("c");
  await expect(page.getByRole("dialog", { name: en["categoryDialog.titleNew.spending"] })).toBeVisible();
  await page.keyboard.press("Escape");

  await page.keyboard.press("?");
  await expect(page.getByRole("dialog", { name: en["app.shortcuts"] })).toBeVisible();
  await page.keyboard.press("Escape");
});

test("typing in a field never triggers a shortcut", async ({ page }) => {
  await openApp(page, { sample: true });
  const search = page.getByRole("textbox", { name: en["shortcuts.search"] });
  await search.fill("n i c");
  await expect(page.getByRole("dialog")).toHaveCount(0);
  await expect(search).toHaveValue("n i c");
});

test("slash focuses the search and Escape clears it", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.keyboard.press("/");
  const search = page.getByRole("textbox", { name: en["shortcuts.search"] });
  await expect(search).toBeFocused();

  await search.fill("Rent");
  await page.keyboard.press("Escape");
  await expect(search).toHaveValue("");
});

test("the command palette finds an entry and opens it", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.keyboard.press("Control+k");

  const palette = page.getByRole("textbox", { name: en["palette.placeholder"] });
  await expect(palette).toBeFocused();
  await palette.fill("Holiday");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("Enter");

  const edit = page.getByRole("dialog", { name: en["entryDialog.titleEdit.spending"] });
  await expect(edit).toBeVisible();
  // Scoped to the dialog: "Rename: …" on every category header contains "Name".
  await expect(edit.getByLabel(en["entryDialog.name"], { exact: true })).toHaveValue("Holiday");
});

test("Ctrl+Enter saves from inside the entry dialog", async ({ page }) => {
  await openApp(page, { sample: true });
  await block(page, "spending").getByRole("button", { name: en["block.addEntry.spending"] }).click();

  const dialog = page.getByRole("dialog");
  await dialog.getByLabel(en["entryDialog.name"]).fill("Entry A");
  await dialog.getByLabel(en["entryDialog.amount"]).fill("9.99");
  await dialog.getByLabel(en["entryDialog.notes"]).press("Control+Enter");

  // While adding, Ctrl+Enter behaves as "save and add another".
  await expect(dialog.getByLabel(en["entryDialog.name"])).toHaveValue("");
  await dialog.getByRole("button", { name: en["dialog.cancel"] }).click();
  await expect(row(page, "Entry A")).toBeVisible();
});

test("the data menu opens, runs an item and closes again", async ({ page }) => {
  await openApp(page, { sample: true });
  const menu = page.getByRole("button", { name: en["app.data"], exact: true });

  await menu.click();
  await expect(page.getByRole("menu")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("menu")).toBeHidden();
  await expect(menu).toBeFocused();

  await menu.click();
  await page.getByRole("menuitem", { name: en["app.backups"] }).click();
  await expect(page.getByRole("dialog", { name: en["backups.title"] })).toBeVisible();
});

test("an open menu swallows the single-key shortcuts", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.getByRole("button", { name: en["app.data"], exact: true }).click();
  await expect(page.getByRole("menu")).toBeVisible();

  await page.keyboard.press("n");
  await expect(page.getByRole("dialog")).toHaveCount(0);
});

test("a press outside closes the menu", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.getByRole("button", { name: en["app.data"], exact: true }).click();
  await expect(page.getByRole("menu")).toBeVisible();

  await page.getByRole("heading", { name: en["app.title"], level: 1 }).click();
  await expect(page.getByRole("menu")).toBeHidden();
});
