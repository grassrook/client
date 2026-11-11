<script lang="ts">
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";

  import type { Judgement, Result } from "$src/lib/judging.svelte";
  import {
    CreateTournamentActivityLog,
    GetAllQualifyingScores,
    UpdateEntrant,
    UpdateQualifyingScore,
    UpdateTournament,
  } from "$wails/go/main/App";
  import {
    type entrant,
    qualifyingscore,
    type tournament,
  } from "$wails/go/models";

  import Button from "../ui/button/button.svelte";

  let {
    tournament,
    judgement,
    entrants,
  }: {
    tournament: tournament.Tournament;
    judgement: Judgement;
    entrants: entrant.Entrant[];
  } = $props();

  const queryClient = useQueryClient();

  let submitting = $state(false);
  const currentScore = $derived(judgement.finalizedScore());

  const updateTournamentMutation = createMutation(() => ({
    mutationFn: async ({ qualifyingOrder }: { qualifyingOrder: string[] }) =>
      UpdateTournament(tournament.id, { qualifyingOrder }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["tournament", tournament.id],
      });
      await queryClient.invalidateQueries({
        queryKey: ["logs", tournament.id],
      });
    },
  }));

  const updateEntrantMutation = createMutation(() => ({
    mutationFn: async ({
      entrantId,
      score,
      status,
    }: {
      entrantId: string;
      score: number;
      status: entrant.Entrant["status"];
    }) => {
      await UpdateEntrant({
        TournamentID: tournament.id,
        EntrantID: entrantId,
        Score: score,
        Status: status,
      });

      const current = entrants.find((entrant) => entrant.id === entrantId);
      if (!current) {
        console.error("no entrant found");
        return;
      }

      await CreateTournamentActivityLog({
        tournament_id: tournament.id,
        content: `${current.name} scored ${score} on qualifying`,
        logType: "QUALIFYING_SCORED",
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["entrants", tournament.id],
      });
      await queryClient.invalidateQueries({
        queryKey: ["tournament", tournament.id],
      });

      await queryClient.invalidateQueries({
        queryKey: ["logs", tournament.id],
      });

      const result = await queryClient.query({
        queryKey: ["scores", tournament.id],
        queryFn: () => GetAllQualifyingScores(tournament.id),
      });

      const nextFront = tournament.qualifying_order
        .split(",")
        .map((id) => id.trim())[0];

      const nextScores =
        result.find((data) => data.participantId === nextFront)?.scores ?? [];

      const nextRound = Array.from(
        { length: tournament.qualifyingRounds },
        (_, i) => i + 1,
      ).find((round) => nextScores[round - 1]?.status !== "scored");

      if (nextRound !== undefined) {
        qualifyingRound = String(nextRound) as QualifyingRound;
      }
    },
  }));

  const orderedEntrants = $derived.by(() => {
    const order = tournament.qualifying_order.split(",").map((id) => id.trim());
    const byParticipantId = new Map(
      entrants.map((entrant) => [entrant.participantId, entrant]),
    );

    return order
      .map((id) => byParticipantId.get(id))
      .filter((entrant): entrant is entrant.Entrant => entrant !== undefined);
  });

  async function onSubmitJudgement() {
    if (submitting) return;

    judgement.processing = true;
    submitting = true;

    const round = !judgement.edit
      ? Number(qualifyingRound)
      : judgement.edit.round;

    try {
      await handleQualifying(
        judgement.finalizedScore().lead, // lead score only since qualifying is most likely using lead criteria
        round,
        judgement.results,
      );
      judgement.clear();
    } catch (error) {
      console.error("onSubmitJudgement", error);
    } finally {
      submitting = false;
      judgement.processing = false;
    }
  }

  type QualifyingRound = "1" | "2";
  let qualifyingRound = $state<QualifyingRound>("1");

  async function handleQualifying(
    score: number,
    round: number,
    results: Result,
  ) {
    const front = orderedEntrants[0];
    let order = tournament.qualifying_order.split(",").map((id) => id.trim());

    if (!judgement.edit) {
      // move this current participant to the back of the order
      const first = order.shift();
      if (first) {
        order.push(first);
      }

      const onLastRound = round == tournament.qualifyingRounds;
      // remove participant from qualifying order if already scored all rounds
      if (onLastRound) {
        order.pop();
      }

      await updateTournamentMutation.mutateAsync({
        qualifyingOrder: order,
      });
    }

    const participantId = judgement.edit
      ? judgement.edit.participantID
      : front.participantId;

    await UpdateQualifyingScore(
      new qualifyingscore.UpdateQualifyingScoreParams({
        TournamentID: tournament.id,
        ParticipantID: participantId,
        Round: round,
        Score: score,
        Status: "scored",
        Results: results,
      }),
    );

    const entrantId = judgement.edit ? judgement.edit.id : front.id;
    await updateEntrantMutation.mutateAsync({
      entrantId,
      score,
      status: "scored",
    });

    if (judgement.edit) {
      judgement.onSubmitJudgementEdit();
    }
  }

  const scores = createQuery(() => ({
    queryKey: ["scores", tournament.id],
    queryFn: () => GetAllQualifyingScores(tournament.id),
  }));

  const frontScores = $derived(
    scores.data?.find(
      (data) => data.participantId === orderedEntrants[0]?.participantId,
    )?.scores ?? [],
  );

  // using initialized state instead of onMount because frontScores data
  // comes after the page is mounted
  let initialized = $state(false);
  $effect(() => {
    if (initialized || frontScores.length === 0) return;

    const nextRound = frontScores.findIndex(
      (score) => score.status !== "scored",
    );

    if (nextRound !== -1) {
      qualifyingRound = String(nextRound + 1) as QualifyingRound;
    }

    initialized = true;
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
      <strong class="text-sm tabular-nums">{currentScore.lead}</strong>
    </div>

    {#if !judgement.edit && tournament.qualifyingRounds > 1}
      <div class="flex flex-col gap-2 space-2">
        <!-- svelte-ignore a11y_label_has_associated_control -->
        <label class="text-sm font-medium">Qualifying Round </label>
        <div class="flex flex-wrap gap-2">
          {#each Array.from({ length: tournament.qualifyingRounds }, (_, i) => i + 1) as round (round)}
            {@const isScored = frontScores[round - 1]?.status === "scored"}

            <Button
              type="button"
              variant={isScored
                ? "ghost"
                : qualifyingRound === String(round)
                  ? "default"
                  : "outline"}
              disabled={isScored}
              onclick={() =>
                (qualifyingRound = String(round) as QualifyingRound)}
              aria-pressed={qualifyingRound === String(round)}
            >
              Round {round}
            </Button>
          {/each}
        </div>
      </div>
    {/if}

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
