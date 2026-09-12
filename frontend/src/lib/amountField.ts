/**
 * Pure validation rules for money input fields (no Svelte, testable with
 * Node). The reactive wrapper with the debounce lives in validate.svelte.ts.
 */
import { parseEuro } from "./money.ts";

export type AmountErrorKey = "entryDialog.invalidAmount" | "errors.entry.amountPositive" | "errors.settings.savingsGoalNegative" | null;

export interface AmountFieldState {
  /** Parsed value in cents, or null when the text is not a number. */
  cents: number | null;
  /** Translation key of the error, null when the value is acceptable. */
  errorKey: AmountErrorKey;
}

/**
 * @param value      the typed text
 * @param decimalMark decimal mark of the current language
 * @param allowEmpty empty text is acceptable (means 0, e.g. "no savings goal")
 */
export function amountFieldState(value: string, decimalMark: "," | ".", allowEmpty: boolean): AmountFieldState {
  if (value.trim() === "") {
    return allowEmpty ? { cents: 0, errorKey: null } : { cents: null, errorKey: "entryDialog.invalidAmount" };
  }
  const cents = parseEuro(value, decimalMark);
  if (cents === null) return { cents: null, errorKey: "entryDialog.invalidAmount" };
  if (allowEmpty) {
    // Savings goal: 0 is fine, negative is not.
    return { cents, errorKey: cents < 0 ? "errors.settings.savingsGoalNegative" : null };
  }
  return { cents, errorKey: cents <= 0 ? "errors.entry.amountPositive" : null };
}
