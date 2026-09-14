/**
 * Shared helpers for the end-to-end tests.
 *
 * The visible copy comes from the language files instead of being repeated
 * here, so a reworded label moves the tests with it rather than breaking
 * them. English is the default language of the application, so the tests
 * speak English unless one asks for another language — i18n.spec.ts does,
 * and drives the whole interface in German.
 */
import { expect, type Page, type Locator } from "@playwright/test";
import { de } from "../src/i18n/de.ts";
import { en } from "../src/i18n/en.ts";

export { de, en };

export type Lang = "en" | "de";

/** The copy of a language. */
export function dict(lang: Lang) {
  return lang === "de" ? de : en;
}

/**
 * Starts a fresh planner and opens it. `sample: true` preloads the example
 * plan of 4.4 in the same language, so its category names match the
 * interface; `lang` saves that language first, the way a user would have
 * chosen it. English is the default, so it is saved as "no choice made".
 */
export async function openApp(page: Page, options: { sample?: boolean; lang?: Lang } = {}): Promise<void> {
  const lang = options.lang ?? "en";
  await page.request.post("/test/reset", {
    data: { sample: options.sample ? lang : "", language: lang === "en" ? "" : lang },
  });
  await page.goto("/");
  await expect(page.getByRole("heading", { name: dict(lang)["app.title"], level: 1 })).toBeVisible();
  // The first state has arrived once the status bar knows the data file.
  await expect(page.getByTestId("status-path")).toBeVisible();
}

/**
 * Waits until the animations of an element have finished. Dialogs pop in
 * over 140 ms (8.6), and measuring during that scales every rectangle by a
 * percent or two — which looks exactly like a layout bug.
 */
export async function settled(locator: Locator): Promise<void> {
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

/**
 * The collapsible group of a category, found by its name. The name is looked
 * up in the header only: a plain text filter over the whole group would also
 * match a category whose *entries* mention the name — "Insurance" matches
 * the entry "Dental insurance" under "Reserves".
 */
export function group(page: Page, name: string) {
  return page.getByTestId("category-group").filter({ has: page.locator(".head .title", { hasText: name }) });
}

/** The block of a kind; its accessible name is the translated kind. */
export function block(page: Page, kind: "income" | "spending", lang: Lang = "en") {
  const t = dict(lang);
  return page.getByRole("table", { name: kind === "income" ? t["kind.income"] : t["kind.spending"] });
}

/** The segmented control names a period by how it is paid, not by its adverb. */
export function periodLabels(lang: Lang = "en") {
  const t = dict(lang);
  return {
    monthly: t["entryDialog.perMonth"],
    quarterly: t["entryDialog.perQuarter"],
    halfyearly: t["entryDialog.perHalfYear"],
    yearly: t["entryDialog.perYear"],
  } as const;
}

/**
 * Fills the entry dialog and saves it. Only the fields a test cares about
 * are passed; the rest keeps the dialog's own defaults.
 */
export async function addEntry(
  page: Page,
  values: {
    name: string;
    amount: string;
    period?: keyof ReturnType<typeof periodLabels>;
    category?: string;
    due?: string;
    notes?: string;
    lang?: Lang;
  },
): Promise<void> {
  const lang = values.lang ?? "en";
  const t = dict(lang);
  const periods = periodLabels(lang);
  const dialog = page.getByRole("dialog");
  await settled(dialog); // typing into a dialog that is still popping in is racy
  await dialog.getByLabel(t["entryDialog.name"]).fill(values.name);
  await dialog.getByLabel(t["entryDialog.amount"]).fill(values.amount);
  if (values.period) await dialog.getByRole("radio", { name: periods[values.period] }).click();
  if (values.category) await dialog.getByLabel(t["entryDialog.category"]).selectOption({ label: values.category });
  if (values.due) await dialog.getByLabel(t["entryDialog.dueMonth"]).selectOption({ label: values.due });
  if (values.notes) await dialog.getByLabel(t["entryDialog.notes"]).fill(values.notes);
  await dialog.getByRole("button", { name: t["dialog.add"], exact: true }).click();
  await expect(dialog).toBeHidden();
}
