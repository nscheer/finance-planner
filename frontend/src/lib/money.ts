/**
 * Money helpers. Amounts are handled as integer cents everywhere (matching the
 * Go backend) and only formatted for display.
 */

const formatter = new Intl.NumberFormat("de-DE", {
  style: "currency",
  currency: "EUR",
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});

/** Formats cents as "1.234,56 €". */
export function formatEuro(cents: number): string {
  return formatter.format(cents / 100);
}

/** Formats cents with an explicit sign, e.g. "+12,00 €" / "-3,50 €". */
export function formatEuroSigned(cents: number): string {
  const s = formatEuro(Math.abs(cents));
  return cents < 0 ? `-${s}` : `+${s}`;
}

/** Formats cents as a plain editable number ("1234,56") for input fields. */
export function centsToInput(cents: number): string {
  const abs = Math.abs(cents);
  const euros = Math.floor(abs / 100);
  const rest = abs % 100;
  return `${cents < 0 ? "-" : ""}${euros},${rest.toString().padStart(2, "0")}`;
}

/**
 * Parses user input into cents. Accepts "12", "12,5", "12.50", "1.234,56",
 * "1,234.56" and "1 234,56 €". When both "." and "," occur, the last one is
 * the decimal separator. A single separator followed by exactly three digits
 * is treated as a thousands separator ("1.234" = 1234 €).
 * Returns null for input that is not a number.
 */
export function parseEuro(input: string): number | null {
  let s = input.trim().replace(/[€\s]/g, "");
  if (s === "") return null;
  const negative = s.startsWith("-");
  if (negative) s = s.slice(1);

  const lastDot = s.lastIndexOf(".");
  const lastComma = s.lastIndexOf(",");
  let decimalSep: "." | "," | null = null;
  if (lastDot >= 0 && lastComma >= 0) {
    decimalSep = lastDot > lastComma ? "." : ",";
  } else if (lastDot >= 0 || lastComma >= 0) {
    const sep = lastDot >= 0 ? "." : ",";
    const parts = s.split(sep);
    const groupsOfThree = parts.slice(1).every((p) => p.length === 3);
    decimalSep = groupsOfThree ? null : sep;
    if (decimalSep !== null && parts.length > 2) return null; // "1,2,3"
  }

  let integer = s;
  let fraction = "";
  if (decimalSep !== null) {
    const idx = s.lastIndexOf(decimalSep);
    integer = s.slice(0, idx);
    fraction = s.slice(idx + 1);
  }
  integer = integer.replace(/[.,]/g, "");
  if (integer === "") integer = "0";
  if (!/^\d+$/.test(integer) || !/^\d*$/.test(fraction)) return null;
  if (fraction.length > 2) return null;
  fraction = fraction.padEnd(2, "0");

  const cents = parseInt(integer, 10) * 100 + parseInt(fraction, 10);
  if (!Number.isSafeInteger(cents)) return null;
  return negative ? -cents : cents;
}
