<script lang="ts">
  /** Sets the monthly savings goal shown in the statistics box. */
  import Modal from "./Modal.svelte";
  import { Service, app, apply, closeDialog, errorMessage, notify } from "../lib/store.svelte";
  import { t, amountInput, amountPlaceholder, parseAmount } from "../lib/i18n.svelte";
  import { untrack } from "svelte";

  const initial = untrack(() => app.state?.stats.savingsGoalCents ?? 0);
  let amount = $state(initial > 0 ? amountInput(initial) : "");
  let error = $state("");
  let working = $state(false);

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    error = "";
    const cents = amount.trim() === "" ? 0 : parseAmount(amount);
    if (cents === null || cents < 0) {
      error = t("entryDialog.invalidAmount");
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
        <input id="goal-amount" class="input" type="text" inputmode="decimal" bind:value={amount} placeholder={amountPlaceholder()} autocomplete="off" />
        <span>€</span>
      </div>
      <span class="hint">{t("goalDialog.hint")}</span>
    </div>
  </form>
  {#snippet footer()}
    <button class="btn" type="button" onclick={closeDialog}>{t("dialog.cancel")}</button>
    <button class="btn btn-primary" type="submit" form="goal-form" disabled={working}>{t("dialog.save")}</button>
  {/snippet}
</Modal>
