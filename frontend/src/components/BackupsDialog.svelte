<script lang="ts">
  /** Lists the automatic backups and restores one after confirmation. */
  import { onMount } from "svelte";
  import Modal from "./Modal.svelte";
  import Icon from "./Icon.svelte";
  import { Service, type BackupInfo, closeDialog, errorMessage, openDialog, restoreBackup } from "../lib/store.svelte";
  import { t, plural, formatDateTime } from "../lib/i18n.svelte";

  let backups = $state<BackupInfo[]>([]);
  let error = $state("");
  let loading = $state(true);

  onMount(async () => {
    try {
      backups = (await Service.ListBackups()) ?? [];
    } catch (err) {
      error = errorMessage(err);
    } finally {
      loading = false;
    }
  });

  function restore(b: BackupInfo) {
    const path = b.path;
    const time = formatDateTime(b.time);
    openDialog({
      type: "confirm",
      title: t("confirm.restoreBackup.title"),
      message: t("confirm.restoreBackup.message", { time }),
      confirmLabel: t("confirm.restoreBackup.confirm"),
      onConfirm: async () => {
        await restoreBackup(path);
      },
    });
  }
</script>

<Modal title={t("backups.title")} width={520} onclose={closeDialog}>
  <p class="intro">{t("backups.intro")}</p>
  {#if error}<p class="form-error">{error}</p>{/if}
  {#if loading}
    <p class="muted">{t("app.loading")}</p>
  {:else if backups.length === 0}
    <p class="muted">{t("backups.empty")}</p>
  {:else}
    <ul class="list">
      {#each backups as b (b.path)}
        <li>
          <span class="icon"><Icon name="history" size={15} /></span>
          <span class="when">{formatDateTime(b.time)}</span>
          <span class="what muted">{t("backups.content", { categories: plural("importDialog.categories", b.categories), entries: plural("importDialog.entries", b.entries) })}</span>
          <button class="btn btn-sm" type="button" onclick={() => restore(b)}>{t("backups.restore")}</button>
        </li>
      {/each}
    </ul>
  {/if}
  {#snippet footer()}
    <button class="btn" type="button" onclick={closeDialog}>{t("dialog.close")}</button>
  {/snippet}
</Modal>

<style>
  .intro {
    margin: 4px 0 12px;
    font-size: 12.5px;
    color: var(--text-2);
  }
  .list {
    margin: 0 0 12px;
    padding: 0;
    list-style: none;
    max-height: 320px;
    overflow: auto;
  }
  li {
    display: grid;
    grid-template-columns: auto auto 1fr auto;
    align-items: center;
    gap: 10px;
    padding: 6px 4px;
    border-top: 1px solid var(--border);
  }
  li:first-child {
    border-top: 0;
  }
  .icon {
    display: flex;
    color: var(--text-3);
  }
  .when {
    font-weight: 500;
    white-space: nowrap;
  }
  .what {
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
