/**
 * The printed document (specification 3.10). Nothing here can be checked by
 * clicking: the print stylesheet only applies under print media, which is
 * why these rules used to break unnoticed.
 *
 * Two things have to line up. The stylesheet needs print media, and the
 * expanded categories come from the `printing` flag that the Print button
 * sets before it calls window.print(). So the tests press the real button,
 * with window.print() stubbed out — a print dialog would block the browser.
 */
import { test, expect } from "@playwright/test";
import { openApp, group, en } from "./app.ts";

/** Presses Print and puts the page into print media, as the printer sees it. */
async function print(page: import("@playwright/test").Page) {
  await page.getByRole("button", { name: en["app.print"] }).click();
  await page.emulateMedia({ media: "print" });
}

test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => {
    window.print = () => {};
  });
});

test("the screen furniture is gone and the document header is there", async ({ page }) => {
  await openApp(page, { sample: true });
  await print(page);

  await expect(page.locator(".topbar")).toBeHidden();
  await expect(page.locator(".statusbar")).toBeHidden();
  await expect(page.getByRole("button", { name: en["app.print"] })).toBeHidden();

  const header = page.locator(".print-header");
  await expect(header).toBeVisible();
  await expect(header).toContainText(en["app.title"]);
  await expect(header).toContainText("data.json");
});

test("every category prints expanded", async ({ page }) => {
  await openApp(page, { sample: true });
  await group(page, "Housing").getByRole("button", { expanded: true }).click(); // collapse it
  await expect(group(page, "Housing")).not.toContainText("Rent");

  await print(page);
  await expect(group(page, "Housing")).toContainText("Rent");
});

test("the timeline prints its figures as a table", async ({ page }) => {
  await openApp(page, { sample: true });
  await print(page);

  const values = page.locator(".print-values");
  await expect(values).toBeVisible();
  // Two half years side by side, six months each.
  await expect(values.locator("tbody tr")).toHaveCount(6);
  await expect(values).toContainText(en["month.1"].slice(0, 3));
  await expect(values).toContainText(en["month.12"].slice(0, 3));
});

test("the page fits the width of a sheet of paper", async ({ page }) => {
  await openApp(page, { sample: true });
  await print(page);

  // A4 at 96 dpi minus the 15 mm margins.
  await page.setViewportSize({ width: 737, height: 1000 });
  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
  );
  expect(overflow).toBeLessThanOrEqual(1); // sub-pixel rounding only
});

test("the parts that need a pointer are not printed", async ({ page }) => {
  await openApp(page, { sample: true });
  await print(page);

  await expect(page.locator(".handle").first()).toBeHidden();
  await expect(page.locator(".chevron").first()).toBeHidden();
  await expect(page.locator(".count").first()).toBeHidden();

  // The rows must stay aligned with the column titles even so (7.9): the
  // hidden cells keep their place in the grid instead of being removed.
  const title = await page.locator(".columns span").nth(4).boundingBox(); // per month
  const cell = await page.getByTestId("entry-row").first().locator(".amount").first().boundingBox();
  expect(Math.abs(title!.x + title!.width - (cell!.x + cell!.width))).toBeLessThan(2);
});
