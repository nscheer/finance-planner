/**
 * Search and period filter (specification 3.6).
 */
import { test, expect } from "@playwright/test";
import { openApp, addEntry, row, group, block, en } from "./app.ts";

test("the search matches names and notes and reports how much is shown", async ({ page }) => {
  await openApp(page, { sample: true });
  const search = page.getByRole("textbox", { name: en["shortcuts.search"] });

  await search.fill("Rent");
  await expect(row(page, "Rent")).toBeVisible();
  await expect(row(page, "Electricity")).toHaveCount(0);
  await expect(group(page, "Leisure")).toHaveCount(0); // no match, so the category is gone
  await expect(page.locator(".filter-result")).toHaveText(/\b1\b.*\b15\b/);

  await search.fill("nothing like this");
  await expect(page.getByText(en["block.noMatch"]).first()).toBeVisible();
});

test("the search also looks in the notes", async ({ page }) => {
  await openApp(page, { sample: true });
  await row(page, "Internet").getByRole("button", { name: `${en["entry.edit"]}: Internet` }).click();
  await page.getByRole("dialog").getByLabel(en["entryDialog.notes"]).fill("Vertrag 4711");
  await page.getByRole("dialog").getByRole("button", { name: en["dialog.save"], exact: true }).click();

  await page.getByRole("textbox", { name: en["shortcuts.search"] }).fill("4711");
  await expect(row(page, "Internet")).toBeVisible();
  await expect(row(page, "Rent")).toHaveCount(0);
});

test("the period filter restricts to one period and the reset clears everything", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.getByRole("combobox", { name: en["app.filter.all"] }).selectOption({ label: en["entry.yearly"] });

  await expect(row(page, "Car insurance")).toBeVisible(); // yearly
  await expect(row(page, "Rent")).toHaveCount(0); // monthly
  await expect(row(page, "Dental insurance")).toHaveCount(0); // quarterly

  await page.getByRole("button", { name: en["app.filterReset"] }).click();
  await expect(row(page, "Rent")).toBeVisible();
});

test("subtotals stay those of the whole category while a filter is active", async ({ page }) => {
  await openApp(page, { sample: true });
  const subtotal = await group(page, "Housing").locator(".head .money").last().textContent();

  await page.getByRole("textbox", { name: en["shortcuts.search"] }).fill("Rent");
  await expect(group(page, "Housing").locator(".head .money").last()).toHaveText(subtotal ?? "");
});

test("drag and drop is off while a filter is active", async ({ page }) => {
  await openApp(page, { sample: true });
  // The rows stay draggable="true" for the browser; the drag is refused in
  // the handler and the row is marked as locked.
  await expect(row(page, "Rent")).not.toHaveClass(/locked/);

  await page.getByRole("textbox", { name: en["shortcuts.search"] }).fill("Rent");
  await expect(row(page, "Rent")).toHaveClass(/locked/);
});

test("a paused-only filter finds exactly the paused entries", async ({ page }) => {
  await openApp(page, { sample: true });
  await row(page, "Streaming").getByRole("button", { name: `${en["entry.pause"]}: Streaming` }).click();

  await page.getByRole("combobox", { name: en["app.filter.all"] }).selectOption({ label: en["app.filter.paused"] });
  await expect(row(page, "Streaming")).toBeVisible();
  await expect(page.getByTestId("entry-row")).toHaveCount(1);
});

test("adding an entry while nothing matches still works", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.getByRole("textbox", { name: en["shortcuts.search"] }).fill("zzz");
  await block(page, "spending").getByRole("button", { name: en["block.addEntry.spending"] }).click();
  await addEntry(page, { name: "zzz Entry", amount: "5.00" });
  await expect(row(page, "zzz Entry")).toBeVisible();
});
