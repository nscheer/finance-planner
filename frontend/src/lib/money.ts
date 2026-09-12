/**
 * Money helpers. Amounts are handled as integer cents everywhere (matching the
 * Go backend) and only formatted for display.
 */

const formatters = new Map<string, Intl.NumberFormat>();

function formatter(locale: string): Intl.NumberFormat {
  let f = formatters.get(locale);
  if (!f) {
    f = new Intl.NumberFormat(locale, {
      style: "currency",
      currency: "EUR",
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
    formatters.set(locale, f);
  }
  return f;
}

/** Formats cents as a € amount in the given locale, e.g. "1.234,56 €" (de-DE). */
export function formatEuroIn(cents: number, locale: string): string {
  return formatter(locale).format(cents / 100);
}

/** Formats cents with an explicit sign, e.g. "+12,00 €" / "-3,50 €". */
export function formatEuroSignedIn(cents: number, locale: string): string {
  const s = formatEuroIn(Math.abs(cents), locale);
  return cents < 0 ? `-${s}` : `+${s}`;
}

/**
 * Formats cents as a plain editable number for input fields, e.g. "1234,56"
 * with the German decimal mark or "1234.56" with the English one.
 */
export function centsToInput(cents: number, decimalMark: "," | "." = ","): string {
  const abs = Math.abs(cents);
  const euros = Math.floor(abs / 100);
  const rest = abs % 100;
  return `${cents < 0 ? "-" : ""}${euros}${decimalMark}${rest.toString().padStart(2, "0")}`;
}

/**
 * Parses user input into cents according to the language's decimal mark.
 * With "," (German): "12", "12,5", "1.234,56", "1 234,56 €" and "10.123"
 * (= 10123 €). With "." (English): "12.5", "1,234.56" and "10,123".
 * The other character is accepted only as a thousands separator, i.e. in
 * front of exactly three digits; at most two decimals. Returns null for
 * input that is not a number.
 */
export function parseEuro(input: string, decimalMark: "," | "." = ","): number | null {
  let s = input.trim().replace(/[€\s]/g, "");
  if (s === "") return null;
  const negative = s.startsWith("-");
  if (negative) s = s.slice(1);

  const groupMark = decimalMark === "," ? "." : ",";
  const idx = s.indexOf(decimalMark);
  if (idx >= 0 && s.indexOf(decimalMark, idx + 1) >= 0) return null; // two decimal marks
  let integer = idx >= 0 ? s.slice(0, idx) : s;
  let fraction = idx >= 0 ? s.slice(idx + 1) : "";

  // Thousands separators only in the integer part, each followed by 3 digits.
  if (integer.includes(groupMark)) {
    const groups = integer.split(groupMark);
    if (groups.slice(1).some((g) => g.length !== 3) || !/^\d{1,3}$/.test(groups[0])) return null;
    integer = groups.join("");
  }
  if (fraction.includes(groupMark)) return null;
  if (integer === "") integer = "0";
  if (!/^\d+$/.test(integer) || !/^\d*$/.test(fraction)) return null;
  if (fraction.length > 2) return null;
  fraction = fraction.padEnd(2, "0");

  const cents = parseInt(integer, 10) * 100 + parseInt(fraction, 10);
  if (!Number.isSafeInteger(cents)) return null;
  return negative ? -cents : cents;
}
