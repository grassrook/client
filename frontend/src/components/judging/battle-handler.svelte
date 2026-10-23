<script lang="ts">
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";
  import { onDestroy } from "svelte";

  import type { Judgement } from "$src/lib/judging.svelte";
  import type { Stage } from "$src/lib/stage.svelte";
  import {
    ChangeLeader,
    GetNote,
    UpdateMatchSingleParticipant,
    UpdateNote,
  } from "$wails/go/main/App";

  import Button from "../ui/button/button.svelte";
  import Card from "../ui/card/card.svelte";
  import CardContent from "../ui/card/card-content.svelte";
  import CardHeader from "../ui/card/card-header.svelte";
  import CardTitle from "../ui/card/card-title.svelte";
  import Label from "../ui/label/label.svelte";
  import Textarea from "../ui/textarea/textarea.svelte";

  let {
    tournamentID,
    judgement,
    stage,
  }: {
    tournamentID: string;
    judgement: Judgement;
    stage: Stage;
  } = $props();

  const queryClient = useQueryClient();
  let submitting = $state(false);
  const currentScore = $derived(judgement.finalizedScore());

  async function onSubmitJudgement() {
    if (submitting) return;

    judgement.processing = true;
    submitting = true;

    try {
      const scores = judgement.finalizedScore();
      stage.setScore(scores);
      const [p1, p2] = stage.getScores();

      await Promise.all([
        UpdateMatchSingleParticipant({
          TournamentID: tournamentID,
          MatchID: stage.matchId!,
          ParticipantID: p1.participantId,
          AppendScore: true,
          Score: p1.score,
        }),
        UpdateMatchSingleParticipant({
          TournamentID: tournamentID,
          MatchID: stage.matchId!,
          ParticipantID: p2.participantId,
          AppendScore: true,
          Score: p2.score,
        }),
      ]);

      await ChangeLeader(tournamentID, stage.matchId!);

      await queryClient.invalidateQueries({
        queryKey: ["activeBattles", tournamentID],
      });

      await queryClient.invalidateQueries({
        queryKey: ["matches", tournamentID],
      });

      stage.switchLeader();
      judgement.clear();
    } finally {
      submitting = false;
      judgement.processing = false;
    }
  }

  let note = $state("");
  let debounceTimer: ReturnType<typeof setTimeout> | undefined;

  const noteQuery = createQuery(() => ({
    queryKey: ["note", tournamentID, stage.matchId],
    queryFn: () => GetNote(stage.matchId!),
  }));

  const updateNoteMutation = createMutation(() => ({
    mutationFn: (text: string) => UpdateNote(stage.matchId!, text),
    onError: (error) => {
      console.error("Failed to update note:", error);
    },
  }));

  $effect(() => {
    const data = noteQuery.data;
    if (data !== undefined) {
      note = data;
    }
  });

  function handleNoteChange(value: string) {
    note = value;

    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }

    const matchId = stage.matchId;
    if (!matchId) return;

    debounceTimer = setTimeout(() => {
      updateNoteMutation.mutate(value);
    }, 500);
  }

  onDestroy(() => {
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }
  });
</script>

<Card class="gap-0 border-border/70 shadow-none">
  <CardHeader class="pb-4">
    <CardTitle>Judgement</CardTitle>
  </CardHeader>

  <CardContent class="space-y-3 pt-0">
    <div
      class="flex items-center justify-between gap-4 rounded-md border bg-muted/30 px-3 py-2.5"
    >
      <span class="text-sm font-medium text-muted-foreground"
        >Current score</span
      >
      <div class="flex items-center gap-3 text-sm tabular-nums">
        <span>Lead <strong>{currentScore.lead}</strong></span>
        <span class="text-muted-foreground">/</span>
        <span>Chase <strong>{currentScore.chase}</strong></span>
      </div>
    </div>

    <div class="flex flex-col gap-2">
      <Label>Quick note</Label>
      <Textarea
        class="h-24"
        value={note}
        oninput={(event) => handleNoteChange(event.currentTarget.value)}
      />
    </div>

    <Button
      class="h-11 w-full"
      onclick={onSubmitJudgement}
      disabled={submitting}
    >
      {#if submitting}
        Submitting...
      {:else}
        Submit Judgment
      {/if}
    </Button>

    <Button
      variant="outline"
      type="reset"
      class="w-full"
      onclick={() => judgement.clear()}
      disabled={submitting}
    >
      Reset
    </Button>
  </CardContent>
</Card>
