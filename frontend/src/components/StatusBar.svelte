<script lang="ts">
  /**
   * The thin bar along the bottom edge: where the data lives on the left,
   * what it contains in the middle, which application version is running on
   * the right. Three fixed slots, so more can be added later without moving
   * what is already there.
   */
  import Icon from "./Icon.svelte";
  import { app, notify } from "../lib/store.svelte";
  import { t, plural } from "../lib/i18n.svelte";

  const categoryCount = $derived((app.state?.income ?? []).length + (app.state?.spending ?? []).length);
  const entryCount = $derived(
    [...(app.state?.income ?? []), ...(app.state?.spending ?? [])].reduce((sum, c) => sum + (c.entries ?? []).length, 0),
  );

  /** The path is long and awkward to retype, so a click puts it on the clipboard. */
  async function copyPath() {
    const path = app.state?.dataPath;
    if (!path) return;
    try {
      await navigator.clipboard.writeText(path);
      notify("success", t("status.pathCopied"));
    } catch {
      notify("error", t("status.pathCopyFailed"));
    }
  }
</script>

<footer class="statusbar">
  {#if app.state}
    <button class="path" type="button" title={t("status.pathTitle", { path: app.state.dataPath })} onclick={copyPath}>
      <Icon name="folder" size={13} />
      <span class="path-text">{app.state.dataPath}</span>
    </button>
    <span class="counts">
      {plural("status.categories", categoryCount)} · {plural("status.entries", entryCount)}
    </span>
  {:else}
    <span></span>
    <span></span>
  {/if}
  <span class="version">{t("app.title")} {app.state?.appVersion ?? ""}</span>
</footer>

<style>
  .statusbar {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
    align-items: center;
    gap: 16px;
    height: var(--statusbar-h);
    flex-shrink: 0;
    padding: 0 24px;
    background: var(--surface);
    border-top: 1px solid var(--border);
    font-size: 11.5px;
    color: var(--text-3);
  }
  .path {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    padding: 0;
    border: 0;
    border-radius: var(--radius-sm);
    background: transparent;
    color: inherit;
    font: inherit;
    cursor: pointer;
  }
  .path:hover {
    color: var(--text-2);
  }
  .path-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .counts {
    white-space: nowrap;
  }
  .version {
    justify-self: end;
    white-space: nowrap;
  }
</style>
