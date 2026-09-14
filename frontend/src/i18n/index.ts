/**
 * Language registry and pure translation helpers (no Svelte, no Wails), so
 * they can be unit tested with plain Node.
 *
 * To add a language:
 *  1. create `<code>.ts` exporting a `Messages` object (TypeScript reports
 *     missing or unknown keys),
 *  2. add it to `locales` below with its label and number-format locale.
 */
import { en } from "./en.ts";
import { de } from "./de.ts";

/** All keys are defined by the English file. */
export type Messages = { [K in keyof typeof en]: string };
export type MessageKey = keyof Messages;

/** Base keys that have ".one" / ".other" plural variants. */
type PluralBase<K> = K extends `${infer Base}.one` ? (`${Base}.other` extends MessageKey ? Base : never) : never;
export type PluralKey = PluralBase<MessageKey>;

export type LocaleCode = "en" | "de";

export interface Locale {
  code: LocaleCode;
  /** Name shown in the language dropdown, in that language. */
  label: string;
  /** BCP 47 tag used for number and currency formatting. */
  numberLocale: string;
  messages: Messages;
}

export const locales: readonly Locale[] = [
  { code: "en", label: "English", numberLocale: "en-IE", messages: en },
  { code: "de", label: "Deutsch", numberLocale: "de-DE", messages: de },
];

/** Used until the user has chosen a language (empty language in data.json). */
export const defaultLocale: LocaleCode = "en";

export function findLocale(code: string | undefined | null): Locale {
  return locales.find((l) => l.code === code) ?? locales.find((l) => l.code === defaultLocale)!;
}

export type Params = Record<string, string | number>;

/** Replaces {name} placeholders. Unknown placeholders are left as they are. */
export function interpolate(template: string, params?: Params): string {
  if (!params) return template;
  return template.replace(/\{(\w+)\}/g, (match, key: string) =>
    key in params ? String(params[key]) : match,
  );
}

/** Looks up a key in the messages, falling back to English, then to the key. */
export function translate(messages: Messages, key: MessageKey, params?: Params): string {
  const template = messages[key] ?? en[key] ?? key;
  return interpolate(template, params);
}

/** Selects the ".one" or ".other" form of a plural key. */
export function translatePlural(messages: Messages, key: PluralKey, count: number, params?: Params): string {
  const full = (count === 1 ? `${key}.one` : `${key}.other`) as MessageKey;
  return translate(messages, full, { ...params, count });
}

/** Whether a key exists in the message catalogue. */
export function hasKey(key: string): key is MessageKey {
  return key in en;
}

/** Returns the placeholder names used in a template, e.g. ["name", "count"]. */
export function placeholders(template: string): string[] {
  return [...template.matchAll(/\{(\w+)\}/g)].map((m) => m[1]).sort();
}
