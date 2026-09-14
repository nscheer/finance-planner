/**
 * Categories and entries: the main view of the specification (3.3) and the
 * entry dialog (3.8), driven the way a user drives them.
 */
import { test, expect } from "@playwright/test";
import { openApp, addEntry, row, group, block, settled, de, periodLabels } from "./app.ts";

test("an empty planner offers the example plan and then shows it", async ({ page }) => {
  await openApp(page);
  await expect(page.getByText(de["app.getStarted.title"])).toBeVisible();

  await page.getByRole("button", { name: de["app.getStarted.sample"] }).click();
  await page.getByRole("dialog").getByRole("button", { name: de["confirm.sample.confirm"] }).click();

  await expect(page.getByText(de["app.getStarted.title"])).toBeHidden();
  await expect(row(page, "Miete")).toBeVisible();
  await expect(page.getByTestId("status-counts")).toContainText("7");
  await expect(page.getByTestId("status-counts")).toContainText("15");
});

test("a category has to exist before an entry can be added", async ({ page }) => {
  await openApp(page);
  await block(page, "spending").getByRole("button", { name: de["block.addEntry.spending"] }).click();

  const dialog = page.getByRole("dialog");
  await expect(dialog).toContainText(de["entryDialog.noCategories.spending"]);
  await dialog.getByRole("button", { name: de["entryDialog.addCategoryFirst"] }).click();

  // The category dialog opens, and creating the category reopens the entry dialog.
  await dialog.getByLabel(de["categoryDialog.name"]).fill("Kategorie A");
  // Opened from the entry dialog, the category dialog returns there by itself,
  // so it offers the plain submit button only.
  await dialog.getByRole("button", { name: de["categoryDialog.submit"] }).click();
  await expect(dialog.getByLabel(de["entryDialog.name"])).toBeVisible();

  await addEntry(page, { name: "Eintrag A", amount: "12,34" });
  await expect(row(page, "Eintrag A")).toContainText("12,34 €");
});

test("a new entry updates its category, its block and the statistics", async ({ page }) => {
  await openApp(page, { sample: true });
  const insurance = group(page, "Versicherungen");
  const before = await insurance.locator(".subtotal, .money").first().textContent();

  await insurance.getByRole("button", { name: `${de["category.addEntry"]}: Versicherungen` }).click();
  await addEntry(page, { name: "Rechtsschutz", amount: "240,00", period: "yearly", due: de["month.5"] });

  const added = row(page, "Rechtsschutz");
  await expect(added).toContainText("20,00 €"); // 240 / 12
  await expect(added).toContainText("240,00 €");
  await expect(added).toContainText(de["month.5"]);
  await expect(added).toContainText(de["entry.yearly"]); // the badge, not the dialog wording
  await expect(insurance).not.toContainText(before ?? "");
});

test("an invalid amount is refused without moving the form", async ({ page }) => {
  await openApp(page, { sample: true });
  await block(page, "spending").getByRole("button", { name: de["block.addEntry.spending"] }).click();

  const dialog = page.getByRole("dialog");
  await settled(dialog);
  const category = dialog.getByLabel(de["entryDialog.category"]);
  // Measured inside the panel: the dialog is centred in the window, so it
  // moves as a whole when its content grows. What must not change is where
  // the fields sit within the form.
  const offset = async () => (await category.boundingBox())!.y - (await dialog.boundingBox())!.y;
  const before = await offset();

  await dialog.getByLabel(de["entryDialog.amount"]).fill("12.50"); // a point is a thousands mark in German
  await expect(dialog.getByText(de["entryDialog.invalidAmount"])).toBeVisible();

  // The hint line is reserved, so the message must not push the form down.
  expect(await offset()).toBe(before);

  await dialog.getByLabel(de["entryDialog.amount"]).fill("12,50");
  await expect(dialog.getByText(de["entryDialog.invalidAmount"])).toBeHidden();
});

test("save and add another keeps the category and clears the name", async ({ page }) => {
  await openApp(page, { sample: true });
  await block(page, "spending").getByRole("button", { name: de["block.addEntry.spending"] }).click();

  const dialog = page.getByRole("dialog");
  await dialog.getByLabel(de["entryDialog.category"]).selectOption({ label: "Freizeit" });
  await dialog.getByRole("radio", { name: periodLabels.quarterly }).click();
  await dialog.getByLabel(de["entryDialog.name"]).fill("Eintrag A");
  await dialog.getByLabel(de["entryDialog.amount"]).fill("30,00");
  await dialog.getByRole("button", { name: de["entryDialog.saveAndNext"] }).click();

  await expect(dialog).toBeVisible();
  await expect(dialog.getByLabel(de["entryDialog.name"])).toHaveValue("");
  await expect(dialog.getByLabel(de["entryDialog.amount"])).toHaveValue("");
  await expect(dialog.getByLabel(de["entryDialog.category"])).toHaveValue(/.+/);
  await expect(dialog.getByRole("radio", { name: periodLabels.quarterly })).toHaveAttribute("aria-checked", "true");

  await addEntry(page, { name: "Eintrag B", amount: "15,00" });
  await expect(group(page, "Freizeit")).toContainText("Eintrag A");
  await expect(group(page, "Freizeit")).toContainText("Eintrag B");
});

test("deleting an entry can be undone", async ({ page }) => {
  await openApp(page, { sample: true });
  await row(page, "Strom").hover();
  await row(page, "Strom").getByRole("button", { name: `${de["entry.delete"]}: Strom` }).click();
  await page.getByRole("dialog").getByRole("button", { name: de["dialog.delete"] }).click();

  await expect(row(page, "Strom")).toHaveCount(0);
  await page.getByRole("button", { name: de["toast.undo"] }).click();
  await expect(row(page, "Strom")).toBeVisible();
});

test("an entry can be paused and counts nowhere afterwards", async ({ page }) => {
  await openApp(page, { sample: true });
  const monthly = block(page, "spending").locator(".totals").first();
  const before = await monthly.textContent();

  await row(page, "Fitnessstudio").getByRole("button", { name: `${de["entry.pause"]}: Fitnessstudio` }).click();
  await expect(row(page, "Fitnessstudio")).toHaveClass(/paused/);
  await expect(monthly).not.toHaveText(before ?? "");
});
