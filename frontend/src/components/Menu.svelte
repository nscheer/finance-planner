<script lang="ts">
  /**
   * A button that opens a small drop-down menu of actions.
   *
   * The top bar cannot hold every action as its own button: with German
   * labels the row is wider than the window (see the specification, 3.2), and
   * the buttons then overlap the search box. The rarely used file actions
   * live in here instead.
   */
  import { tick } from "svelte";
  import Icon, { type IconName } from "./Icon.svelte";

  type MenuItem = { label: string; icon: IconName; run: () => void };

  let {
    label,
    title,
    items,
    onOpenChange,
  }: {
    label: string;
    title?: string;
    items: MenuItem[];
    /** Lets the shell suppress the single-key shortcuts while the menu is open. */
    onOpenChange?: (open: boolean) => void;
  } = $props();

  let open = $state(false);
  let active = $state(0);
  let root: HTMLDivElement | undefined = $state();
  let button: HTMLButtonElement | undefined = $state();
  let itemEls: (HTMLButtonElement | undefined)[] = $state([]);

  function setOpen(value: boolean) {
    open = value;
    onOpenChange?.(value);
  }

  async function toggle() {
    if (open) {
      close();
      return;
    }
    active = 0;
    setOpen(true);
    await tick(); // the list exists only after this update
    itemEls[0]?.focus();
  }

  function close(focusButton = true) {
    if (!open) return;
    setOpen(false);
    if (focusButton) button?.focus();
  }

  function choose(item: MenuItem) {
    close();
    item.run();
  }

  function focusItem(index: number) {
    const count = items.length;
    active = ((index % count) + count) % count;
    itemEls[active]?.focus();
  }

  function onMenuKeydown(event: KeyboardEvent) {
    switch (event.key) {
      case "Escape":
        close();
        break;
      case "ArrowDown":
        focusItem(active + 1);
        break;
      case "ArrowUp":
        focusItem(active - 1);
        break;
      case "Home":
        focusItem(0);
        break;
      case "End":
        focusItem(items.length - 1);
        break;
      case "Tab":
        // Leaving the menu by keyboard closes it, like a click outside does.
        close(false);
        return;
      default:
        return;
    }
    event.preventDefault();
  }

  function onButtonKeydown(event: KeyboardEvent) {
    if (event.key !== "ArrowDown" || open) return;
    event.preventDefault();
    toggle();
  }

  /** A press anywhere outside closes the menu. */
  function onPointerDown(event: PointerEvent) {
    if (!open || !root) return;
    if (!root.contains(event.target as Node)) close(false);
  }
</script>

<svelte:window onpointerdown={onPointerDown} />

<div class="menu" bind:this={root}>
  <button
    class="btn btn-sm"
    class:open
    type="button"
    {title}
    aria-haspopup="menu"
    aria-expanded={open}
    bind:this={button}
    onclick={toggle}
    onkeydown={onButtonKeydown}
  >
    {label}
    <span class="chevron" class:open><Icon name="chevron" size={13} /></span>
  </button>
  {#if open}
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div class="list" role="menu" tabindex="-1" onkeydown={onMenuKeydown}>
      {#each items as item, i (item.label)}
        <button
          class="item"
          type="button"
          role="menuitem"
          bind:this={itemEls[i]}
          onclick={() => choose(item)}
          onmouseenter={() => (active = i)}
        >
          <Icon name={item.icon} size={14} />
          {item.label}
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .menu {
    position: relative;
    display: flex;
  }
  .chevron {
    display: flex;
    color: var(--text-3);
    transform: rotate(90deg);
    transition: transform 120ms ease;
  }
  .chevron.open {
    transform: rotate(-90deg);
  }
  .btn.open {
    background: var(--surface-3);
  }
  .list {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    z-index: 30;
    display: flex;
    flex-direction: column;
    min-width: 200px;
    padding: 4px;
    background: var(--surface);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    box-shadow: var(--shadow-lg);
  }
  .item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 7px 10px 8px;
    border: 0;
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--text);
    font: inherit;
    text-align: left;
    white-space: nowrap;
    cursor: pointer;
  }
  .item:hover,
  .item:focus-visible {
    background: var(--accent-soft);
    color: var(--accent-strong);
    outline: none;
  }
</style>
