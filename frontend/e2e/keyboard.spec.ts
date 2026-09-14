/**
 * Keyboard shortcuts, the command palette and the data menu
 * (specification 3.2 and 3.9).
 */
import { test, expect } from "@playwright/test";
import { openApp, row, block, de } from "./app.ts";

test("the single-key shortcuts open what they promise", async ({ page }) => {
  await openApp(page, { sample: true });

  await page.keyboard.press("n");
  await expect(page.getByRole("dialog", { name: de["entryDialog.titleNew.spending"] })).toBeVisible();
  await page.keyboard.press("Escape");

  await page.keyboard.press("i");
  await expect(page.getByRole("dialog", { name: de["entryDialog.titleNew.income"] })).toBeVisible();
  await page.keyboard.press("Escape");

  await page.keyboard.press("c");
  await expect(page.getByRole("dialog", { name: de["categoryDialog.titleNew.spending"] })).toBeVisible();
  await page.keyboard.press("Escape");

  await page.keyboard.press("?");
  await expect(page.getByRole("dialog", { name: de["app.shortcuts"] })).toBeVisible();
  await page.keyboard.press("Escape");
});

test("typing in a field never triggers a shortcut", async ({ page }) => {
  await openApp(page, { sample: true });
  const search = page.getByRole("textbox", { name: de["shortcuts.search"] });
  await search.fill("n i c");
  await expect(page.getByRole("dialog")).toHaveCount(0);
  await expect(search).toHaveValue("n i c");
});

test("slash focuses the search and Escape clears it", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.keyboard.press("/");
  const search = page.getByRole("textbox", { name: de["shortcuts.search"] });
  await expect(search).toBeFocused();

  await search.fill("Miete");
  await page.keyboard.press("Escape");
  await expect(search).toHaveValue("");
});

test("the command palette finds an entry and opens it", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.keyboard.press("Control+k");

  const palette = page.getByRole("textbox", { name: de["palette.placeholder"] });
  await expect(palette).toBeFocused();
  await palette.fill("Urlaub");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("Enter");

  await expect(page.getByRole("dialog", { name: de["entryDialog.titleEdit.spending"] })).toBeVisible();
  await expect(page.getByLabel(de["entryDialog.name"])).toHaveValue("Urlaub");
});

test("Ctrl+Enter saves from inside the entry dialog", async ({ page }) => {
  await openApp(page, { sample: true });
  await block(page, "spending").getByRole("button", { name: de["block.addEntry.spending"] }).click();

  const dialog = page.getByRole("dialog");
  await dialog.getByLabel(de["entryDialog.name"]).fill("Eintrag A");
  await dialog.getByLabel(de["entryDialog.amount"]).fill("9,99");
  await dialog.getByLabel(de["entryDialog.notes"]).press("Control+Enter");

  // While adding, Ctrl+Enter behaves as "save and add another".
  await expect(dialog.getByLabel(de["entryDialog.name"])).toHaveValue("");
  await dialog.getByRole("button", { name: de["dialog.cancel"] }).click();
  await expect(row(page, "Eintrag A")).toBeVisible();
});

test("the data menu opens, runs an item and closes again", async ({ page }) => {
  await openApp(page, { sample: true });
  const menu = page.getByRole("button", { name: de["app.data"] });

  await menu.click();
  await expect(page.getByRole("menu")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("menu")).toBeHidden();
  await expect(menu).toBeFocused();

  await menu.click();
  await page.getByRole("menuitem", { name: de["app.backups"] }).click();
  await expect(page.getByRole("dialog", { name: de["backups.title"] })).toBeVisible();
});

test("an open menu swallows the single-key shortcuts", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.getByRole("button", { name: de["app.data"] }).click();
  await expect(page.getByRole("menu")).toBeVisible();

  await page.keyboard.press("n");
  await expect(page.getByRole("dialog")).toHaveCount(0);
});

test("a press outside closes the menu", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.getByRole("button", { name: de["app.data"] }).click();
  await expect(page.getByRole("menu")).toBeVisible();

  await page.getByRole("heading", { name: de["app.title"], level: 1 }).click();
  await expect(page.getByRole("menu")).toBeHidden();
});
