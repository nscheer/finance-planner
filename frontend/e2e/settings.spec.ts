/**
 * Settings, the top bar and the status bar (specification 3.2, 3.11, 5).
 */
import { test, expect } from "@playwright/test";
import { openApp, row, block, de, en } from "./app.ts";

test("German is the language until another one is chosen", async ({ page }) => {
  await openApp(page, { sample: true });
  await expect(page.getByRole("button", { name: de["app.print"] })).toBeVisible();

  await page.getByRole("combobox", { name: de["app.language"] }).selectOption("en");
  await expect(page.getByRole("button", { name: en["app.print"] })).toBeVisible();

  // The choice is saved, so a reload keeps English.
  await page.reload();
  await expect(page.getByRole("button", { name: en["app.print"] })).toBeVisible();
});

test("the language decides how an amount is read", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.getByRole("combobox", { name: de["app.language"] }).selectOption("en");

  await block(page, "spending", "en").getByRole("button", { name: en["block.addEntry.spending"] }).click();
  const dialog = page.getByRole("dialog");
  await dialog.getByLabel(en["entryDialog.name"]).fill("Entry A");
  await dialog.getByLabel(en["entryDialog.amount"]).fill("12.50"); // valid in English
  await expect(dialog.getByText(en["entryDialog.invalidAmount"])).toBeHidden();
  await dialog.getByRole("button", { name: en["dialog.add"], exact: true }).click();

  await expect(row(page, "Entry A")).toContainText("€12.50");
});

test("the appearance choice is applied and survives a reload", async ({ page }) => {
  await openApp(page, { sample: true });
  await expect(page.locator("html")).not.toHaveAttribute("data-theme", "dark");

  await page.getByRole("combobox", { name: de["app.theme"] }).selectOption("dark");
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");

  await page.reload();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
});

test("the savings goal shows up in the statistics and can be removed", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.getByRole("button", { name: de["stats.goalEdit"] }).click();

  const dialog = page.getByRole("dialog");
  await dialog.getByLabel(de["goalDialog.amount"]).fill("300,00");
  await dialog.getByRole("button", { name: de["dialog.save"], exact: true }).click();
  await expect(page.getByText("300,00 €").first()).toBeVisible();

  await page.getByRole("button", { name: de["stats.goalEdit"] }).click();
  await dialog.getByLabel(de["goalDialog.amount"]).fill("0");
  await dialog.getByRole("button", { name: de["dialog.save"], exact: true }).click();
  await expect(page.getByText(de["stats.goalNone"])).toBeVisible();
});

test("the status bar names the file, the counts and the version", async ({ page }) => {
  await openApp(page, { sample: true });
  await expect(page.getByTestId("status-path")).toContainText("data.json");
  await expect(page.getByTestId("status-counts")).toContainText("7");
  await expect(page.getByTestId("status-counts")).toContainText("15");
  await expect(page.getByTestId("status-version")).toContainText(de["app.title"]);
  await expect(page.getByTestId("status-version")).toContainText(/\d+\.\d+\.\d+/);
});

test("clicking the path copies it", async ({ page, context }) => {
  await context.grantPermissions(["clipboard-read", "clipboard-write"]);
  await openApp(page, { sample: true });

  const shown = await page.getByTestId("status-path").textContent();
  await page.getByTestId("status-path").click();
  await expect(page.getByText(de["status.pathCopied"])).toBeVisible();

  const copied = await page.evaluate(() => navigator.clipboard.readText());
  expect(shown).toContain(copied);
});

test("the top bar does not overlap itself at the smallest window width", async ({ page }) => {
  await openApp(page, { sample: true });
  await page.setViewportSize({ width: 1200, height: 700 });

  const search = await page.locator(".search").boundingBox();
  const actions = await page.locator(".topbar .actions").boundingBox();
  expect(search!.x + search!.width).toBeLessThanOrEqual(actions!.x);

  // And the period filter stays reachable rather than being covered.
  await expect(page.getByRole("combobox", { name: de["app.filter.all"] })).toBeVisible();
});
