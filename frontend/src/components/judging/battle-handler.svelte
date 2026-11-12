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
    UpdateEliminationScore,
    UpdateEliminationScoreByRound,
    UpdateMatchSingleParticipant,
    UpdateNote,
  } from "$wails/go/main/App";
  import { eliminationscore, type repository } from "$wails/go/models";

  import Button from "../ui/button/button.svelte";
  import Label from "../ui/label/label.svelte";
  import Textarea from "../ui/textarea/textarea.svelte";

  let {
    tournamentID,
    judgement,
    stage,
    activeBattles,
  }: {
    tournamentID: string;
    judgement: Judgement;
    stage: Stage;
    activeBattles: repository.ActiveBattle[];
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

      if (!judgement.edit) {
        const paramp1 = new eliminationscore.UpdateEliminationScoreParams({
          participantID: p1.participantId,
          score: p1.score,
          leadParticipantID: stage.getLeader(),
          matchId: stage.matchId!,
          results: judgement.results,
        });
        const paramp2 = new eliminationscore.UpdateEliminationScoreParams({
          participantID: p2.participantId,
          score: p2.score,
          leadParticipantID: stage.getLeader(),
          matchId: stage.matchId!,
          results: judgement.results,
        });

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
          UpdateEliminationScore(paramp1),
          UpdateEliminationScore(paramp2),
        ]);

        await ChangeLeader(tournamentID, stage.matchId!);

        await queryClient.invalidateQueries({
          queryKey: ["activeBattles", tournamentID],
        });

        await queryClient.invalidateQueries({
          queryKey: ["logs", tournamentID],
        });

        await queryClient.invalidateQueries({
          queryKey: ["matches", tournamentID],
        });

        const isFinished = activeBattles
          .filter((battle) => battle.matchId === stage.matchId)
          .every((battle) => battle.leadStatus === -1);

        // if round 2, set both run finished to close the panel
        if (isFinished) {
          stage.bothRunFinished = true;
        } else {
          stage.switchLeader();
          judgement.clear();
        }
      } else {
        const paramp1 =
          new eliminationscore.UpdateEliminationScoreByRoundParams({
            participantID: p1.participantId,
            score: p1.score,
            round: judgement.edit.round,
            matchId: stage.matchId!,
            results: judgement.results,
          });

        const paramp2 =
          new eliminationscore.UpdateEliminationScoreByRoundParams({
            participantID: p2.participantId,
            score: p2.score,
            round: judgement.edit.round,
            matchId: stage.matchId!,
            results: judgement.results,
          });
        const [scorep1, scorep2] = await Promise.all([
          await UpdateEliminationScoreByRound(paramp1),
          await UpdateEliminationScoreByRound(paramp2),
        ]);

        await Promise.all([
          UpdateMatchSingleParticipant({
            TournamentID: tournamentID,
            MatchID: stage.matchId!,
            ParticipantID: p1.participantId,
            AppendScore: false,
            Score: scorep1,
          }),
          UpdateMatchSingleParticipant({
            TournamentID: tournamentID,
            MatchID: stage.matchId!,
            ParticipantID: p2.participantId,
            AppendScore: false,
            Score: scorep2,
          }),
        ]);

        await queryClient.invalidateQueries({
          queryKey: ["activeBattles", tournamentID],
        });

        await queryClient.invalidateQueries({
          queryKey: ["logs", tournamentID],
        });

        await queryClient.invalidateQueries({
          queryKey: ["matches", tournamentID],
        });

        stage.clear();
        judgement.onSubmitJudgementEdit();
      }
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

<div
  class="gap-0 rounded-xl border border-border/70 text-card-foreground shadow-none p-4"
>
  <div class="space-y-3">
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
  </div>
</div>
