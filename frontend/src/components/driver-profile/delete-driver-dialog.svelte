<script lang="ts">
  import { Trash2 } from "@lucide/svelte";

  import { Button } from "$src/components/ui/button";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
  } from "$src/components/ui/dialog";

  type Props = {
    open: boolean;
    driverName: string | undefined;
    error: string;
    deleting: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: () => void;
  };

  let { open, driverName, error, deleting, onOpenChange, onConfirm }: Props =
    $props();
</script>

<Dialog {open} {onOpenChange}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Delete driver?</DialogTitle>
      <DialogDescription>
        You're about to delete
        <strong class="text-foreground">{driverName}</strong>. Their registered
        cars and associated photos will also be deleted. This action cannot be
        undone.
      </DialogDescription>
    </DialogHeader>

    {#if error}
      <p class="text-sm text-destructive">{error}</p>
    {/if}

    <DialogFooter>
      <Button
        variant="outline"
        onclick={() => onOpenChange(false)}
        disabled={deleting}
      >
        Cancel
      </Button>
      <Button variant="destructive" onclick={onConfirm} disabled={deleting}>
        <Trash2 class="mr-2 size-4" />
        {deleting ? "Deleting..." : "Delete permanently"}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
