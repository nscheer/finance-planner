<script lang="ts">
  /** Modal notification for errors and other things the user must see. */
  import Modal from "./Modal.svelte";
  import Icon from "./Icon.svelte";
  import { closeDialog } from "../lib/store.svelte";

  let { title, message }: { title: string; message: string } = $props();
</script>

<Modal {title} width={440} onclose={closeDialog}>
  <div class="alert">
    <span class="icon"><Icon name="alert" size={20} /></span>
    <p>{message}</p>
  </div>
  {#snippet footer()}
    <button class="btn btn-primary" type="button" onclick={closeDialog}>OK</button>
  {/snippet}
</Modal>

<style>
  .alert {
    display: flex;
    gap: 12px;
    margin: 4px 0 12px;
    color: var(--text-2);
  }
  .icon {
    color: var(--danger);
    flex-shrink: 0;
    margin-top: 1px;
  }
  p {
    margin: 0;
    user-select: text;
  }
</style>
