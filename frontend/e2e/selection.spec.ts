/**
 * Selection mode and bulk actions (specification 3.5).
 */
import { test, expect } from "@playwright/test";
import { openApp, row, group, block, de } from "./app.ts";

type Page = import("@playwright/test").Page;

const headerCheckbox = (page: Page, kind: "income" | "spending") =>
  block(page, kind).locator(".select-all input");

/**
 * Turns the mode on. The first click on the header checkbox only arms the
 * mode and selects nothing (3.5); it is a click, not a check, because the
 * checkbox is controlled and stays unchecked here.
 */
async function enterSelectMode(page: Page, kind: "income" | "spending") {
  await headerCheckbox(page, kind).click();
  await expect(page.getByRole("toolbar")).toBeVisible();
}

/** Arms the mode and then selects every entry of the block. */
async function selectAll(page: Page, kind: "income" | "spending") {
  await enterSelectMode(page, kind);
  await headerCheckbox(page, kind).click();
}

test("the header checkbox arms the mode, then selects and deselects the block", async ({ page }) => {
  await openApp(page, { sample: true });
  const bar = page.getByRole("toolbar");

  await enterSelectMode(page, "spending");
  await expect(bar).toContainText(de["selection.count.zero"]); // armed, nothing selected

  await headerCheckbox(page, "spending").click();
  await expect(bar).toContainText("12"); // the twelve spendings of the example plan
  await expect(headerCheckbox(page, "spending")).toBeChecked();

  await headerCheckbox(page, "spending").click();
  await expect(bar).toContainText(de["selection.count.zero"]);
});

test("the header checkbox is indeterminate while only some rows are selected", async ({ page }) => {
  await openApp(page, { sample: true });
  await row(page, "Miete").click({ modifiers: ["ControlOrMeta"] });

  const box = headerCheckbox(page, "spending");
  expect(await box.evaluate((el: HTMLInputElement) => el.indeterminate)).toBe(true);
});

test("ctrl+click enters the mode and shift+click takes a range", async ({ page }) => {
  await openApp(page, { sample: true });
  await row(page, "Miete").click({ modifiers: ["ControlOrMeta"] });

  const bar = page.getByRole("toolbar");
  await expect(bar).toContainText("1");

  await row(page, "Internet").click({ modifiers: ["Shift"] });
  await expect(bar).toContainText("3"); // Miete, Strom, Internet
});

test("the floating bar never covers the status bar", async ({ page }) => {
  await openApp(page, { sample: true });
  await enterSelectMode(page, "spending");

  const bar = await page.getByRole("toolbar").boundingBox();
  const status = await page.getByTestId("status-path").boundingBox();
  expect(bar!.y + bar!.height).toBeLessThanOrEqual(status!.y);
});

test("a bulk pause pauses every selected entry", async ({ page }) => {
  await openApp(page, { sample: true });
  await selectAll(page, "spending");
  await page.getByRole("toolbar").getByRole("button", { name: de["selection.pause"] }).click();

  await expect(row(page, "Miete")).toHaveClass(/paused/);
  await expect(row(page, "Urlaub")).toHaveClass(/paused/);
});

test("a bulk move puts the entries into another category", async ({ page }) => {
  await openApp(page, { sample: true });
  await row(page, "Tanken").click({ modifiers: ["ControlOrMeta"] });

  const bar = page.getByRole("toolbar");
  await bar.getByRole("combobox", { name: de["selection.moveTo"] }).selectOption({ label: "Wohnen" });
  await bar.getByRole("button", { name: de["selection.move"], exact: true }).click();

  await expect(group(page, "Wohnen")).toContainText("Tanken");
  await expect(group(page, "Mobilität")).not.toContainText("Tanken");
});

test("a bulk delete can be undone at the original positions", async ({ page }) => {
  await openApp(page, { sample: true });
  await row(page, "Miete").click({ modifiers: ["ControlOrMeta"] });
  await row(page, "Strom").click({ modifiers: ["ControlOrMeta"] });

  await page.getByRole("toolbar").getByRole("button", { name: de["selection.delete"] }).click();
  await page.getByRole("dialog").getByRole("button", { name: de["dialog.delete"] }).click();
  await expect(row(page, "Miete")).toHaveCount(0);

  await page.getByRole("button", { name: de["toast.undo"] }).click();
  await expect(row(page, "Miete")).toBeVisible();
  // Restored where they were: Miete is still the first row of its category.
  await expect(group(page, "Wohnen").getByTestId("entry-row").first()).toContainText("Miete");
});

test("Escape leaves selection mode even after a checkbox was clicked", async ({ page }) => {
  await openApp(page, { sample: true });
  await enterSelectMode(page, "spending");
  await expect(page.getByRole("toolbar")).toBeVisible();

  await page.keyboard.press("Escape");
  await expect(page.getByRole("toolbar")).toBeHidden();
});
