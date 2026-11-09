<script lang="ts">
  import { createMutation, useQueryClient } from "@tanstack/svelte-query";

  import { Button } from "$src/components/ui/button/index.js";
  import * as Dialog from "$src/components/ui/dialog/index.js";
  import { DeleteParticipant } from "$wails/go/main/App";
  import type { entrant } from "$wails/go/models";

  type Props = {
    tournamentId: string;
    target: entrant.Entrant | null;
  };

  let { tournamentId, target = $bindable(null) }: Props = $props();

  const queryClient = useQueryClient();

  const deleteParticipantMutation = createMutation(() => ({
    mutationFn: async ({ participantId }: { participantId: string }) =>
      DeleteParticipant(tournamentId, participantId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["entrants", tournamentId],
      });

      // the participant is also removed from the qualifying order
      await queryClient.invalidateQueries({
        queryKey: ["tournament", tournamentId],
      });
    },
  }));

  function confirm() {
    if (!target) return;

    deleteParticipantMutation.mutate(
      { participantId: target.participantId },
      { onSuccess: () => (target = null) },
    );
  }
</script>

<Dialog.Root
  open={target !== null}
  onOpenChange={(open) => {
    if (!open) target = null;
  }}
>
  <Dialog.Content class="sm:max-w-sm">
    <Dialog.Header>
      <Dialog.Title>Delete participant</Dialog.Title>
      <Dialog.Description>
        Remove <span class="font-medium text-foreground">{target?.name}</span>
        from this tournament? This can't be undone.
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer>
      <Button
        variant="outline"
        onclick={() => (target = null)}
        disabled={deleteParticipantMutation.isPending}
      >
        Cancel
      </Button>
      <Button
        variant="destructive"
        onclick={confirm}
        disabled={deleteParticipantMutation.isPending}
      >
        {deleteParticipantMutation.isPending ? "Deleting..." : "Delete"}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
