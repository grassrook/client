<script lang="ts">
  import PlusIcon from "@lucide/svelte/icons/plus";
  import UsersIcon from "@lucide/svelte/icons/users";

  import { Button } from "$src/components/ui/button/index.js";

  type Props = {
    editable: boolean;
    onAdd: () => void;
  };

  let { editable, onAdd }: Props = $props();
</script>

<div
  class="flex flex-col items-center justify-center rounded-lg border border-dashed px-6 py-16 text-center"
>
  <div
    class="mb-4 flex size-12 items-center justify-center rounded-full bg-muted"
  >
    <UsersIcon class="size-6 text-muted-foreground" />
  </div>
  <h3 class="text-base font-semibold">No participants yet</h3>
  {#if editable}
    <p class="mt-1 max-w-sm text-sm text-muted-foreground">
      Add the people taking part in this tournament.
    </p>
    <Button class="mt-6" onclick={onAdd}>
      <PlusIcon class="mr-2 size-4" />
      Add participants
    </Button>
  {:else}
    <p class="mt-1 max-w-sm text-sm text-muted-foreground">
      This tournament has already started, so participants can no longer be
      added.
    </p>
  {/if}
</div>
