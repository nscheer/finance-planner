/**
 * Shared helpers for the end-to-end tests.
 *
 * The visible copy comes from the language files instead of being repeated
 * here, so a reworded label moves the tests with it rather than breaking
 * them. German is the default language of the application, so `de` is what
 * the tests look for unless a test switches the language itself.
 */
import { expect, type Page } from "@playwright/test";
import { de } from "../src/i18n/de.ts";
import { en } from "../src/i18n/en.ts";

export { de, en };

/**
 * Starts a fresh planner and opens it. `sample: true` preloads the example
 * plan of 4.4, which every test that needs data uses instead of clicking one
 * together.
 */
export async function openApp(page: Page, options: { sample?: boolean } = {}): Promise<void> {
  await page.request.post("/test/reset", { data: { sample: options.sample ? "de" : "" } });
  await page.goto("/");
  await expect(page.getByRole("heading", { name: de["app.title"], level: 1 })).toBeVisible();
  // The first state has arrived once the status bar knows the data file.
  await expect(page.getByTestId("status-path")).toBeVisible();
}

/**
 * Waits until the animations of an element have finished. Dialogs pop in
 * over 140 ms (8.6), and measuring during that scales every rectangle by a
 * percent or two — which looks exactly like a layout bug.
 */
export async function settled(locator: import("@playwright/test").Locator): Promise<void> {
  await locator.evaluate((el) => Promise.all(el.getAnimations().map((a) => a.finished)));
}

/** Queues the answer of the next native file dialog (see cmd/e2e-host). */
export async function nextDialog(page: Page, answer: { path?: string; cancel?: boolean }): Promise<void> {
  await page.request.post("/test/dialog", { data: answer });
}

/** The row of an entry, found by its name. */
export function row(page: Page, name: string) {
  return page.getByTestId("entry-row").filter({ hasText: name });
}

/** The collapsible group of a category, found by its name. */
export function group(page: Page, name: string) {
  return page.getByTestId("category-group").filter({ hasText: name });
}

/** The block of a kind; its accessible name is the translated kind. */
export function block(page: Page, kind: "income" | "spending", lang: "de" | "en" = "de") {
  const dict = lang === "de" ? de : en;
  return page.getByRole("table", { name: kind === "income" ? dict["kind.income"] : dict["kind.spending"] });
}

/**
 * Fills the entry dialog and saves it. Only the fields a test cares about
 * are passed; the rest keeps the dialog's own defaults.
 */
export async function addEntry(
  page: Page,
  values: { name: string; amount: string; period?: keyof typeof periodLabels; category?: string; due?: string; notes?: string },
): Promise<void> {
  const dialog = page.getByRole("dialog");
  await settled(dialog); // typing into a dialog that is still popping in is racy
  await dialog.getByLabel(de["entryDialog.name"]).fill(values.name);
  await dialog.getByLabel(de["entryDialog.amount"]).fill(values.amount);
  if (values.period) await dialog.getByRole("radio", { name: periodLabels[values.period] }).click();
  if (values.category) await dialog.getByLabel(de["entryDialog.category"]).selectOption({ label: values.category });
  if (values.due) await dialog.getByLabel(de["entryDialog.dueMonth"]).selectOption({ label: values.due });
  if (values.notes) await dialog.getByLabel(de["entryDialog.notes"]).fill(values.notes);
  await dialog.getByRole("button", { name: de["dialog.add"], exact: true }).click();
  await expect(dialog).toBeHidden();
}

/** The segmented control names a period by how it is paid, not by its adverb. */
export const periodLabels = {
  monthly: de["entryDialog.perMonth"],
  quarterly: de["entryDialog.perQuarter"],
  halfyearly: de["entryDialog.perHalfYear"],
  yearly: de["entryDialog.perYear"],
} as const;
