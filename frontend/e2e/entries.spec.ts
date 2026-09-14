/**
 * Categories and entries: the main view of the specification (3.3) and the
 * entry dialog (3.8), driven the way a user drives them.
 */
import { test, expect } from "@playwright/test";
import { openApp, addEntry, row, group, block, settled, en, periodLabels } from "./app.ts";

test("an empty planner offers the example plan and then shows it", async ({ page }) => {
  await openApp(page);
  await expect(page.getByText(en["app.getStarted.title"])).toBeVisible();

  await page.getByRole("button", { name: en["app.getStarted.sample"] }).click();
  await page.getByRole("dialog").getByRole("button", { name: en["confirm.sample.confirm"] }).click();

  await expect(page.getByText(en["app.getStarted.title"])).toBeHidden();
  await expect(row(page, "Rent")).toBeVisible();
  await expect(page.getByTestId("status-counts")).toContainText("7");
  await expect(page.getByTestId("status-counts")).toContainText("15");
});

test("a category has to exist before an entry can be added", async ({ page }) => {
  await openApp(page);
  await block(page, "spending").getByRole("button", { name: en["block.addEntry.spending"] }).click();

  const dialog = page.getByRole("dialog");
  await expect(dialog).toContainText(en["entryDialog.noCategories.spending"]);
  await dialog.getByRole("button", { name: en["entryDialog.addCategoryFirst"] }).click();

  // The category dialog opens, and creating the category reopens the entry dialog.
  await dialog.getByLabel(en["categoryDialog.name"]).fill("Category A");
  // Opened from the entry dialog, the category dialog returns there by itself,
  // so it offers the plain submit button only.
  await dialog.getByRole("button", { name: en["categoryDialog.submit"] }).click();
  await expect(dialog.getByLabel(en["entryDialog.name"])).toBeVisible();

  await addEntry(page, { name: "Entry A", amount: "12.34" });
  await expect(row(page, "Entry A")).toContainText("€12.34");
});

test("a new entry updates its category, its block and the statistics", async ({ page }) => {
  await openApp(page, { sample: true });
  const insurance = group(page, "Insurance");
  const before = await insurance.locator(".subtotal, .money").first().textContent();

  await insurance.getByRole("button", { name: `${en["category.addEntry"]}: Insurance` }).click();
  await addEntry(page, { name: "Legal insurance", amount: "240.00", period: "yearly", due: en["month.5"] });

  const added = row(page, "Legal insurance");
  await expect(added).toContainText("€20.00"); // 240 / 12
  await expect(added).toContainText("€240.00");
  await expect(added).toContainText(en["month.5"]);
  await expect(added).toContainText(en["entry.yearly"]); // the badge, not the dialog wording
  await expect(insurance).not.toContainText(before ?? "");
});

test("an invalid amount is refused without moving the form", async ({ page }) => {
  await openApp(page, { sample: true });
  await block(page, "spending").getByRole("button", { name: en["block.addEntry.spending"] }).click();

  const dialog = page.getByRole("dialog");
  await settled(dialog);
  const category = dialog.getByLabel(en["entryDialog.category"]);
  // Measured inside the panel: the dialog is centred in the window, so it
  // moves as a whole when its content grows. What must not change is where
  // the fields sit within the form.
  const offset = async () => (await category.boundingBox())!.y - (await dialog.boundingBox())!.y;
  const before = await offset();

  await dialog.getByLabel(en["entryDialog.amount"]).fill("12,50"); // a comma is a thousands mark in English
  await expect(dialog.getByText(en["entryDialog.invalidAmount"])).toBeVisible();

  // The hint line is reserved, so the message must not push the form down.
  expect(await offset()).toBe(before);

  await dialog.getByLabel(en["entryDialog.amount"]).fill("12.50");
  await expect(dialog.getByText(en["entryDialog.invalidAmount"])).toBeHidden();
});

test("save and add another keeps the category and clears the name", async ({ page }) => {
  await openApp(page, { sample: true });
  await block(page, "spending").getByRole("button", { name: en["block.addEntry.spending"] }).click();

  const dialog = page.getByRole("dialog");
  await dialog.getByLabel(en["entryDialog.category"]).selectOption({ label: "Leisure" });
  await dialog.getByRole("radio", { name: periodLabels().quarterly }).click();
  await dialog.getByLabel(en["entryDialog.name"]).fill("Entry A");
  await dialog.getByLabel(en["entryDialog.amount"]).fill("30.00");
  await dialog.getByRole("button", { name: en["entryDialog.saveAndNext"] }).click();

  await expect(dialog).toBeVisible();
  await expect(dialog.getByLabel(en["entryDialog.name"])).toHaveValue("");
  await expect(dialog.getByLabel(en["entryDialog.amount"])).toHaveValue("");
  await expect(dialog.getByLabel(en["entryDialog.category"])).toHaveValue(/.+/);
  await expect(dialog.getByRole("radio", { name: periodLabels().quarterly })).toHaveAttribute("aria-checked", "true");

  await addEntry(page, { name: "Entry B", amount: "15.00" });
  await expect(group(page, "Leisure")).toContainText("Entry A");
  await expect(group(page, "Leisure")).toContainText("Entry B");
});

test("deleting an entry can be undone", async ({ page }) => {
  await openApp(page, { sample: true });
  await row(page, "Electricity").hover();
  await row(page, "Electricity").getByRole("button", { name: `${en["entry.delete"]}: Electricity` }).click();
  await page.getByRole("dialog").getByRole("button", { name: en["dialog.delete"] }).click();

  await expect(row(page, "Electricity")).toHaveCount(0);
  await page.getByRole("button", { name: en["toast.undo"] }).click();
  await expect(row(page, "Electricity")).toBeVisible();
});

test("an entry can be paused and counts nowhere afterwards", async ({ page }) => {
  await openApp(page, { sample: true });
  const monthly = block(page, "spending").locator(".totals").first();
  const before = await monthly.textContent();

  await row(page, "Gym").getByRole("button", { name: `${en["entry.pause"]}: Gym` }).click();
  await expect(row(page, "Gym")).toHaveClass(/paused/);
  await expect(monthly).not.toHaveText(before ?? "");
});
