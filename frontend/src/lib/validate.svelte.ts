/**
 * Reactive money input field with live validation. The value is validated
 * on every change; the error is shown 400 ms after the last change or at
 * once when the field is left (touch), and disappears immediately when the
 * value becomes valid.
 */
import { amountFieldState, type AmountErrorKey } from "./amountField";
import { decimalMark, t } from "./i18n.svelte";

const SHOW_DELAY_MS = 400;

export class AmountField {
  value = $state("");
  private shown = $state(false);
  private timer: ReturnType<typeof setTimeout> | undefined;

  constructor(
    initial: string,
    private readonly allowEmpty = false,
  ) {
    this.value = initial;
  }

  private get state() {
    return amountFieldState(this.value, decimalMark(), this.allowEmpty);
  }

  /** Parsed value in cents (null while invalid). */
  get cents(): number | null {
    return this.state.errorKey === null ? this.state.cents : null;
  }

  get valid(): boolean {
    return this.state.errorKey === null;
  }

  /** The translated error to display, "" while hidden or valid. */
  get error(): string {
    const key: AmountErrorKey = this.state.errorKey;
    return this.shown && key ? t(key) : "";
  }

  /** Call from the input's oninput: hides the error and restarts the delay. */
  changed(): void {
    this.shown = false;
    clearTimeout(this.timer);
    this.timer = setTimeout(() => (this.shown = true), SHOW_DELAY_MS);
  }

  /** Clears the field for the next entry, without leaving a stale error. */
  reset(value = ""): void {
    clearTimeout(this.timer);
    this.shown = false;
    this.value = value;
  }

  /** Call on blur or submit: shows the error immediately if invalid. */
  touch(): void {
    clearTimeout(this.timer);
    this.shown = true;
  }

  dispose(): void {
    clearTimeout(this.timer);
  }
}
