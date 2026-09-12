/**
 * Color scheme handling. The choice ("light", "dark" or "system") is saved
 * in data.json; "system" follows the operating system and is resolved here,
 * so the stylesheet only has to know data-theme="dark".
 */
import { Service } from "../../bindings/finance-planner/planner";

export type Theme = "light" | "dark" | "system";
export const themes: readonly Theme[] = ["light", "dark", "system"];

/** Light is the default until a choice was saved. */
export const defaultTheme: Theme = "light";

export const theme = $state({ current: defaultTheme as Theme });

const media = typeof window !== "undefined" ? window.matchMedia("(prefers-color-scheme: dark)") : null;
media?.addEventListener("change", () => render());

function render(): void {
  const resolved = theme.current === "system" ? (media?.matches ? "dark" : "light") : theme.current;
  document.documentElement.dataset.theme = resolved;
}

/** Applies a theme locally (used when the saved choice is loaded). */
export function applyTheme(value: string | undefined | null): void {
  theme.current = (themes as readonly string[]).includes(value ?? "") ? (value as Theme) : defaultTheme;
  render();
}

/**
 * Switches the theme and saves the choice. Applied first so the UI reacts
 * immediately; the error (if any) is returned to the caller.
 */
export async function setTheme(value: Theme): Promise<unknown> {
  applyTheme(value);
  try {
    await Service.SetTheme(value);
    return null;
  } catch (err) {
    return err;
  }
}
