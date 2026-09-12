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
import { formatEuroIn, formatEuroSignedIn } from "./money";

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
