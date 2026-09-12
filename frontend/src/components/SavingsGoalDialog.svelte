<script lang="ts">
  /** Sets the monthly savings goal shown in the statistics box. */
  import Modal from "./Modal.svelte";
  import { Service, app, apply, closeDialog, errorMessage, notify } from "../lib/store.svelte";
  import { t, amountInput, amountPlaceholder } from "../lib/i18n.svelte";
  import { AmountField } from "../lib/validate.svelte";
  import { onDestroy, untrack } from "svelte";

  const initial = untrack(() => app.state?.stats.savingsGoalCents ?? 0);
  const amount = new AmountField(initial > 0 ? amountInput(initial) : "", true);
  let amountEl: HTMLInputElement | undefined = $state();
  onDestroy(() => amount.dispose());
  let error = $state("");
  let working = $state(false);

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    error = "";
    amount.touch();
    const cents = amount.cents;
    if (cents === null) {
      amountEl?.focus();
      return;
    }
    working = true;
    try {
      await apply(Service.SetSavingsGoal(cents));
      notify("success", t("toast.goalSaved"));
      closeDialog();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      working = false;
    }
  }
</script>

<Modal title={t("goalDialog.title")} width={420} onclose={closeDialog}>
  <form id="goal-form" onsubmit={submit}>
    {#if error}<p class="form-error">{error}</p>{/if}
    <div class="field">
      <label for="goal-amount">{t("goalDialog.amount")}</label>
      <div class="input-suffix">
        <input
          id="goal-amount"
          class="input"
          class:invalid={!!amount.error}
          type="text"
          inputmode="decimal"
          bind:value={amount.value}
          bind:this={amountEl}
          oninput={() => amount.changed()}
          onblur={() => amount.touch()}
          aria-invalid={!!amount.error}
          placeholder={amountPlaceholder()}
          autocomplete="off"
        />
        <span>€</span>
      </div>
      <span class="hint" class:error={!!amount.error}>{amount.error || t("goalDialog.hint")}</span>
    </div>
  </form>
  {#snippet footer()}
    <button class="btn" type="button" onclick={closeDialog}>{t("dialog.cancel")}</button>
    <button class="btn btn-primary" type="submit" form="goal-form" disabled={working}>{t("dialog.save")}</button>
  {/snippet}
</Modal>
