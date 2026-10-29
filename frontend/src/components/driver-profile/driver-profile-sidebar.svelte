<script lang="ts">
  import { Trash2 } from "@lucide/svelte";

  import { Button } from "$src/components/ui/button";
  import { Input } from "$src/components/ui/input";
  import { Label } from "$src/components/ui/label";
  import Separator from "$src/components/ui/separator/separator.svelte";

  type Props = {
    driverId: string;
    onDelete: () => void;
  };

  let { driverId, onDelete }: Props = $props();
</script>

<aside class="space-y-6">
  <section class="rounded-xl border p-5">
    <h2 class="font-semibold">Driver information</h2>
    <p class="mt-1 text-sm text-muted-foreground">
      System details for this record.
    </p>

    <Separator class="my-4" />

    <div class="space-y-2">
      <Label for="driver-id">Driver ID</Label>
      <Input id="driver-id" value={driverId} readonly />
      <p class="text-xs text-muted-foreground">
        Unique identifier assigned to this driver.
      </p>
    </div>
  </section>

  <section class="rounded-xl border border-destructive/30 p-5">
    <div class="flex items-center gap-2">
      <Trash2 class="size-4 text-destructive" />
      <h2 class="font-semibold">Danger zone</h2>
    </div>

    <p class="mt-2 text-sm text-muted-foreground">
      Deleting this driver also deletes their registered cars and associated car
      photos.
    </p>

    <Button class="mt-4 w-full" variant="destructive" onclick={onDelete}>
      Delete driver
    </Button>
  </section>
</aside>
