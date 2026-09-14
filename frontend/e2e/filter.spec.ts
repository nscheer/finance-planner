/**
 * Search and period filter (specification 3.6).
 */
import { test, expect } from "@playwright/test";
import { openApp, addEntry, row, group, block, de } from "./app.ts";

test("the search matches names and notes and reports how much is shown", async ({ page }) => {
  await openApp(page, { sample: true });
  const search = page.getByRole("textbox", { name: de["shortcuts.search"] });

  await search.fill("Miete");
  await expect(row(page, "Miete")).toBeVisible();
  await expect(row(page, "Strom")).toHaveCount(0);
  await expect(group(page, "Freizeit")).toHaveCount(0); // no match, so the category is gone
  await expect(page.locator(".filter-result")).toHaveText(/\b1\b.*\b15\b/);

  await search.fill("nichts davon");
  await expect(page.getByText(de["block.noMatch"]).first()).toBeVisible();
});

test("the search also looks in the notes", async ({ page }) => {
  await openApp(page, { sample: true });
  await row(page, "Internet").getByRole("button", { name: `${de["entry.edit"]}: Internet` }).click();
  await page.getByRole("dialog").getByLabel(de["entryDialog.notes"]).fill("Vertrag 4711");
  await page.getByRole("dialog").getByRole("button", { name: de["dialog.save"], exact: true }).click();

  await page.getByRole("textbox", { name: de["shortcuts.search"] }).fill("4711");
  await expect(row(page, "Internet")).toBeVisible();
  await expect(row(page, "Miete")).toHaveCount(0);
});

test("the period filter restricts to one period and the reset clears everything", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.getByRole("combobox", { name: de["app.filter.all"] }).selectOption({ label: de["entry.yearly"] });

  await expect(row(page, "Kfz-Versicherung")).toBeVisible(); // yearly
  await expect(row(page, "Miete")).toHaveCount(0); // monthly
  await expect(row(page, "Zahnzusatzversicherung")).toHaveCount(0); // quarterly

  await page.getByRole("button", { name: de["app.filterReset"] }).click();
  await expect(row(page, "Miete")).toBeVisible();
});

test("subtotals stay those of the whole category while a filter is active", async ({ page }) => {
  await openApp(page, { sample: true });
  const subtotal = await group(page, "Wohnen").locator(".head .money").last().textContent();

  await page.getByRole("textbox", { name: de["shortcuts.search"] }).fill("Miete");
  await expect(group(page, "Wohnen").locator(".head .money").last()).toHaveText(subtotal ?? "");
});

test("drag and drop is off while a filter is active", async ({ page }) => {
  await openApp(page, { sample: true });
  // The rows stay draggable="true" for the browser; the drag is refused in
  // the handler and the row is marked as locked.
  await expect(row(page, "Miete")).not.toHaveClass(/locked/);

  await page.getByRole("textbox", { name: de["shortcuts.search"] }).fill("Miete");
  await expect(row(page, "Miete")).toHaveClass(/locked/);
});

test("a paused-only filter finds exactly the paused entries", async ({ page }) => {
  await openApp(page, { sample: true });
  await row(page, "Streaming").getByRole("button", { name: `${de["entry.pause"]}: Streaming` }).click();

  await page.getByRole("combobox", { name: de["app.filter.all"] }).selectOption({ label: de["app.filter.paused"] });
  await expect(row(page, "Streaming")).toBeVisible();
  await expect(page.getByTestId("entry-row")).toHaveCount(1);
});

test("adding an entry while nothing matches still works", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.getByRole("textbox", { name: de["shortcuts.search"] }).fill("zzz");
  await block(page, "spending").getByRole("button", { name: de["block.addEntry.spending"] }).click();
  await addEntry(page, { name: "zzz Eintrag", amount: "5,00" });
  await expect(row(page, "zzz Eintrag")).toBeVisible();
});
