<script lang="ts">
  /** Popup notifications in the lower right corner. */
  import Icon from "./Icon.svelte";
  import { app, dismissToast } from "../lib/store.svelte";
</script>

<div class="toasts" aria-live="polite">
  {#each app.toasts as toast (toast.id)}
    <div class="toast {toast.kind}">
      <Icon name={toast.kind === "error" ? "alert" : toast.kind === "success" ? "check" : "info"} />
      <span>{toast.message}</span>
      <button class="icon-btn" type="button" aria-label="Dismiss" onclick={() => dismissToast(toast.id)}><Icon name="close" size={14} /></button>
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
    background: #1c2130;
    color: #fff;
    box-shadow: var(--shadow-lg);
    animation: slide 0.16s ease-out;
  }
  .toast span {
    flex: 1;
    user-select: text;
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
