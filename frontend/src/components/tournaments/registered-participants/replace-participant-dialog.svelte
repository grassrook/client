<script lang="ts">
  import { createMutation, useQueryClient } from "@tanstack/svelte-query";

  import { Button } from "$src/components/ui/button/index.js";
  import * as Dialog from "$src/components/ui/dialog/index.js";
  import * as Select from "$src/components/ui/select/index.js";
  import { ReplaceParticipant } from "$wails/go/main/App";
  import type { entrant } from "$wails/go/models";

  type Props = {
    tournamentId: string;
    /** the participant being replaced; the dialog is open while this is set */
    target: entrant.Entrant | null;
    /** participants that can take the target's place (not in the bracket) */
    options: entrant.Entrant[];
  };

  let { tournamentId, target = $bindable(null), options }: Props = $props();

  const queryClient = useQueryClient();

  let replacementId = $state("");

  const replacement = $derived(
    options.find((e) => e.participantId === replacementId),
  );

  const replaceParticipantMutation = createMutation(() => ({
    mutationFn: async ({ from, to }: { from: string; to: string }) =>
      ReplaceParticipant(from, to),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["matches", tournamentId] }),
        queryClient.invalidateQueries({
          queryKey: ["activeBattles", tournamentId],
        }),
        queryClient.invalidateQueries({ queryKey: ["entrants", tournamentId] }),
        queryClient.invalidateQueries({ queryKey: ["logs", tournamentId] }),
      ]);
    },
  }));

  function close() {
    target = null;
    replacementId = "";
  }

  function confirm() {
    if (!target || !replacement) return;

    replaceParticipantMutation.mutate(
      { from: target.participantId, to: replacement.participantId },
      { onSuccess: close },
    );
  }
</script>

<Dialog.Root
  open={target !== null}
  onOpenChange={(open) => {
    if (!open) close();
  }}
>
  <Dialog.Content class="sm:max-w-sm">
    <Dialog.Header>
      <Dialog.Title>Replace participant</Dialog.Title>
      <Dialog.Description>
        Replace
        <span class="font-medium text-foreground">{target?.name}</span>
        in the bracket with another participant. To undo this later, replace them
        back the same way.
      </Dialog.Description>
    </Dialog.Header>

    <Select.Root type="single" bind:value={replacementId}>
      <Select.Trigger class="w-full">
        {replacement?.name ?? "Select a participant"}
      </Select.Trigger>
      <Select.Content>
        {#each options as option (option.participantId)}
          <Select.Item value={option.participantId}>{option.name}</Select.Item>
        {/each}
      </Select.Content>
    </Select.Root>

    {#if replaceParticipantMutation.isError}
      <p class="text-sm text-destructive">
        {replaceParticipantMutation.error?.message ?? "Failed to replace."}
      </p>
    {/if}

    <Dialog.Footer>
      <Button
        variant="outline"
        onclick={close}
        disabled={replaceParticipantMutation.isPending}
      >
        Cancel
      </Button>
      <Button
        onclick={confirm}
        disabled={!replacement || replaceParticipantMutation.isPending}
      >
        {replaceParticipantMutation.isPending ? "Replacing..." : "Replace"}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
