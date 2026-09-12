/**
 * Reactive translation layer. Components call `t("key")`; because it reads
 * the reactive locale, everything re-renders when the language changes.
 * The chosen language is stored in data.json via the backend.
 */
import { Service } from "../../bindings/finance-planner/planner";
import {
  findLocale,
  locales,
  translate,
  translatePlural,
  type Locale,
  type LocaleCode,
  type MessageKey,
  type Params,
  type PluralKey,
} from "../i18n";
import { centsToInput, formatEuroIn, formatEuroSignedIn } from "./money";

export { locales };
export type { LocaleCode, MessageKey };

export const i18n = $state({ locale: findLocale(undefined) as Locale });

export function t(key: MessageKey, params?: Params): string {
  return translate(i18n.locale.messages, key, params);
}

export function plural(key: PluralKey, count: number, params?: Params): string {
  return translatePlural(i18n.locale.messages, key, count, params);
}

/** Formats cents in the current language's number format, e.g. "1.234,56 €". */
export function formatEuro(cents: number): string {
  return formatEuroIn(cents, i18n.locale.numberLocale);
}

export function formatEuroSigned(cents: number): string {
  return formatEuroSignedIn(cents, i18n.locale.numberLocale);
}

/** The decimal mark of the current language ("," in German, "." in English). */
export function decimalMark(): "," | "." {
  const parts = new Intl.NumberFormat(i18n.locale.numberLocale).formatToParts(1.5);
  return parts.find((p) => p.type === "decimal")?.value === "," ? "," : ".";
}

/** Amount as shown in an edit field, in the language's decimal notation. */
export function amountInput(cents: number): string {
  return centsToInput(cents, decimalMark());
}

/** Placeholder of amount fields: "0,00" or "0.00". */
export function amountPlaceholder(): string {
  return `0${decimalMark()}00`;
}

/** Formats a ratio (0..1) as a percentage without decimals, e.g. "23 %". */
export function formatPercent(ratio: number): string {
  return new Intl.NumberFormat(i18n.locale.numberLocale, { style: "percent", maximumFractionDigits: 0 }).format(ratio);
}

/** Full month name (1-12) in the current language. */
export function monthName(month: number): string {
  return t(`month.${month}` as MessageKey);
}

/** Three-letter month abbreviation (1-12). */
export function monthShort(month: number): string {
  return monthName(month).slice(0, 3);
}

/** The code of the current language, e.g. "de". */
export function currentLanguage(): LocaleCode {
  return i18n.locale.code;
}

/** Formats a date and time in the current language, e.g. "12.09.2026, 21:04". */
export function formatDateTime(value: string | Date): string {
  const date = typeof value === "string" ? new Date(value) : value;
  return new Intl.DateTimeFormat(i18n.locale.numberLocale, { dateStyle: "medium", timeStyle: "short" }).format(date);
}

/** Applies a language locally (used when the saved choice is loaded). */
export function applyLocale(code: string): void {
  i18n.locale = findLocale(code);
  document.documentElement.lang = i18n.locale.code;
}

/**
 * Switches the language and saves the choice. The switch is applied first so
 * the UI reacts immediately; the error (if any) is returned to the caller.
 */
export async function setLocale(code: LocaleCode): Promise<unknown> {
  applyLocale(code);
  try {
    await Service.SetLanguage(code);
    return null;
  } catch (err) {
    return err;
  }
}
