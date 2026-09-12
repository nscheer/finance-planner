<script lang="ts">
  /** Confirmation modal for destructive actions (delete). */
  import Modal from "./Modal.svelte";
  import { closeDialog } from "../lib/store.svelte";
  import { t } from "../lib/i18n.svelte";

  let {
    title,
    message,
    confirmLabel,
    onConfirm,
  }: { title: string; message: string; confirmLabel: string; onConfirm: () => Promise<void> | void } = $props();

  let working = $state(false);

  async function confirm() {
    working = true;
    try {
      closeDialog();
      await onConfirm();
    } finally {
      working = false;
    }
  }
</script>

<Modal {title} width={420} onclose={closeDialog}>
  <p class="message">{message}</p>
  {#snippet footer()}
    <button class="btn" type="button" onclick={closeDialog}>{t("dialog.cancel")}</button>
    <button class="btn btn-danger" type="button" disabled={working} onclick={confirm}>{confirmLabel}</button>
  {/snippet}
</Modal>

<style>
  .message {
    margin: 4px 0 12px;
    color: var(--text-2);
  }
</style>
