<script lang="ts">
  /** Popup notifications in the lower right corner. */
  import Icon from "./Icon.svelte";
  import { app, dismissToast } from "../lib/store.svelte";
  import { t } from "../lib/i18n.svelte";
</script>

<div class="toasts" aria-live="polite">
  {#each app.toasts as toast (toast.id)}
    <div class="toast {toast.kind}">
      <Icon name={toast.kind === "error" ? "alert" : toast.kind === "success" ? "check" : "info"} />
      <span>{toast.message}</span>
      {#if toast.action}
        <button class="action" type="button" onclick={() => { const a = toast.action; dismissToast(toast.id); a?.run(); }}>{toast.action.label}</button>
      {/if}
      <button class="icon-btn" type="button" aria-label={t("dialog.dismiss")} onclick={() => dismissToast(toast.id)}><Icon name="close" size={14} /></button>
    </div>
  {/each}
</div>

<style>
  .toasts {
    position: fixed;
    right: 20px;
    bottom: 20px;
    z-index: 200;
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-width: 420px;
  }
  .toast {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 8px 10px 14px;
    border-radius: var(--radius-sm);
    background: var(--toast-bg);
    color: #fff;
    box-shadow: var(--shadow-lg);
    animation: slide 0.16s ease-out;
  }
  .toast span {
    flex: 1;
    user-select: text;
  }
  .action {
    padding: 4px 10px;
    border: 1px solid rgba(255, 255, 255, 0.5);
    border-radius: var(--radius-sm);
    background: rgba(255, 255, 255, 0.12);
    color: #fff;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
  }
  .action:hover {
    background: rgba(255, 255, 255, 0.24);
  }
  .toast .icon-btn {
    color: rgba(255, 255, 255, 0.7);
  }
  .toast .icon-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #fff;
  }
  .toast.success {
    background: var(--income-strong);
  }
  .toast.error {
    background: var(--danger);
  }
  @keyframes slide {
    from { transform: translateY(8px); opacity: 0; }
  }
</style>
