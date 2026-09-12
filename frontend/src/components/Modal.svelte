<script lang="ts">
  /**
   * Generic modal dialog. Renders a backdrop, a titled panel and a footer
   * slot for buttons. Escape closes it via `onclose`.
   */
  import type { Snippet } from "svelte";
  import Icon from "./Icon.svelte";

  let {
    title,
    width = 480,
    onclose,
    children,
    footer,
  }: {
    title: string;
    width?: number;
    onclose: () => void;
    children: Snippet;
    footer?: Snippet;
  } = $props();

  function onKeydown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      event.preventDefault();
      onclose();
    }
  }

  /** Focus the first form field, else the primary button, when the dialog opens. */
  function autofocus(node: HTMLElement) {
    const el =
      node.querySelector<HTMLElement>("input, select, textarea") ??
      node.querySelector<HTMLElement>(".btn-primary") ??
      node.querySelector<HTMLElement>("footer button");
    el?.focus();
  }
</script>

<svelte:window onkeydown={onKeydown} />

<div class="backdrop" role="presentation" onmousedown={(e) => e.target === e.currentTarget && onclose()}>
  <div class="panel" role="dialog" aria-modal="true" aria-label={title} style:width="{width}px" use:autofocus>
    <header>
      <h2>{title}</h2>
      <button class="icon-btn" type="button" aria-label="Close" onclick={onclose}><Icon name="close" /></button>
    </header>
    <div class="body">
      {@render children()}
    </div>
    {#if footer}
      <footer>{@render footer()}</footer>
    {/if}
  </div>
</div>

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(20, 26, 40, 0.45);
    animation: fade 0.12s ease-out;
  }
  .panel {
    max-width: calc(100vw - 40px);
    max-height: calc(100vh - 40px);
    display: flex;
    flex-direction: column;
    background: var(--surface);
    border-radius: 12px;
    box-shadow: var(--shadow-lg);
    animation: pop 0.14s ease-out;
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px 12px;
  }
  h2 {
    font-size: 17px;
  }
  .body {
    padding: 4px 20px 8px;
    overflow: auto;
  }
  footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 12px 20px 16px;
    border-top: 1px solid var(--border);
  }
  @keyframes fade {
    from { opacity: 0; }
  }
  @keyframes pop {
    from { transform: translateY(6px) scale(0.98); opacity: 0; }
  }
</style>
