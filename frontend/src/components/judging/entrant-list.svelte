<script lang="ts">
  /* eslint-disable no-useless-assignment */
  import { ArrowUpToLine, Pencil, RotateCcw, UserX } from "@lucide/svelte";
  import EllipsisVertical from "@lucide/svelte/icons/ellipsis-vertical";
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";

  import { Badge } from "$src/components/ui/badge";
  import { Button } from "$src/components/ui/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "$src/components/ui/dropdown-menu";
  import type { Judgement } from "$src/lib/judging.svelte";
  import {
    ConcludeQualifers,
    CreateTournamentActivityLog,
    GetAllQualifyingScores,
    SyncTournament,
    UpdateEntrant,
    UpdateQualifyingScore,
    UpdateTournament,
    UpdateTournamentState,
  } from "$wails/go/main/App";
  import {
    type entrant,
    qualifyingscore,
    type tournament,
  } from "$wails/go/models";

  import AlertDialog from "../ui/alert-dialog/alert-dialog.svelte";
  import AlertDialogAction from "../ui/alert-dialog/alert-dialog-action.svelte";
  import AlertDialogCancel from "../ui/alert-dialog/alert-dialog-cancel.svelte";
  import AlertDialogContent from "../ui/alert-dialog/alert-dialog-content.svelte";
  import AlertDialogDescription from "../ui/alert-dialog/alert-dialog-description.svelte";
  import AlertDialogFooter from "../ui/alert-dialog/alert-dialog-footer.svelte";
  import AlertDialogHeader from "../ui/alert-dialog/alert-dialog-header.svelte";
  import AlertDialogTitle from "../ui/alert-dialog/alert-dialog-title.svelte";
  import AlertDialogTrigger from "../ui/alert-dialog/alert-dialog-trigger.svelte";
  import DropdownMenuSeparator from "../ui/dropdown-menu/dropdown-menu-separator.svelte";
  import Empty from "../ui/empty/empty.svelte";
  import EmptyDescription from "../ui/empty/empty-description.svelte";
  import EmptyHeader from "../ui/empty/empty-header.svelte";
  import EmptyTitle from "../ui/empty/empty-title.svelte";
  import Separator from "../ui/separator/separator.svelte";

  let {
    tournament,
    entrants,
    hidePanel = $bindable(false),
    judgement = $bindable(),
  }: {
    tournament: tournament.Tournament;
    entrants: entrant.Entrant[];
    hidePanel?: boolean;
    judgement: Judgement;
  } = $props();

  const queryClient = useQueryClient();

  const scores = createQuery(() => ({
    queryKey: ["scores", tournament.id],
    queryFn: () => GetAllQualifyingScores(tournament.id),
  }));

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

  function moveToFirst(idx: number) {
    const [entrant] = orderedEntrants.entrants.splice(idx, 1);

    if (!entrant) return;

    const reordered = [entrant, ...orderedEntrants.entrants];
    updateTournamentMutation.mutate({
      qualifyingOrder: reordered.map((entrant) => entrant.participantId),
    });
  }

  type QualifierMutationReq = {
    qualifierId: string;
    status: entrant.Entrant["status"];
    score?: number;
  };

  const updateQualifierMutation = createMutation(() => ({
    mutationFn: async ({ qualifierId, status, score }: QualifierMutationReq) =>
      UpdateEntrant({
        TournamentID: tournament.id,
        EntrantID: qualifierId,
        Status: status,
        Score: score,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["entrants", tournament.id],
      });
      await queryClient.query({
        queryKey: ["scores", tournament.id],
        queryFn: () => GetAllQualifyingScores(tournament.id),
      });
      await queryClient.invalidateQueries({
        queryKey: ["logs", tournament.id],
      });
    },
  }));

  const resetStatuses = ["pending", "not_qualified"];

  async function markAs(
    id: string,
    status: entrant.Entrant["status"],
    participantID: string,
  ) {
    try {
      const current = entrants.find(
        (entrant) => entrant.participantId === participantID,
      );
      if (!current) return;

      if (status === "pending") {
        await handlePending(participantID);
        await CreateTournamentActivityLog({
          tournament_id: tournament.id,
          content: `${current.name} set status to PENDING`,
          logType: "QUALIFYING_RESET",
        });
      } else if (status === "not_qualified") {
        await handleNotQualified(participantID);
        await CreateTournamentActivityLog({
          tournament_id: tournament.id,
          content: `${current.name} set status to NOT_QUALIFIED`,
          logType: "NOT_QUALIFIED",
        });
      }

      const req: QualifierMutationReq = {
        qualifierId: id,
        status,
      };

      if (resetStatuses.includes(status)) {
        req.score = 0;
      }

      await updateQualifierMutation.mutateAsync(req);
    } catch (error) {
      console.error("Failed to update qualifier status:", error);
    }
  }

  async function handleNotQualified(participantID: string) {
    const qualifyingRounds = tournament.qualifyingRounds;

    const qualifyingOrder = tournament.qualifying_order
      .split(",")
      .map((id) => id.trim());

    const filteredQualifyingOrder = qualifyingOrder.filter(
      (id) => id !== participantID,
    );

    await Promise.all(
      Array.from({ length: qualifyingRounds }, (_, i) =>
        UpdateQualifyingScore(
          new qualifyingscore.UpdateQualifyingScoreParams({
            TournamentID: tournament.id,
            ParticipantID: participantID,
            Round: i + 1,
            Score: 0,
            Status: "scored",
            Results: {},
          }),
        ),
      ),
    );

    updateTournamentMutation.mutate({
      qualifyingOrder: filteredQualifyingOrder,
    });
  }

  async function handlePending(participantID: string) {
    const qualifyingRounds = tournament.qualifyingRounds;

    await Promise.all(
      Array.from({ length: qualifyingRounds }, (_, i) =>
        UpdateQualifyingScore(
          new qualifyingscore.UpdateQualifyingScoreParams({
            TournamentID: tournament.id,
            ParticipantID: participantID,
            Round: i + 1,
            Score: 0,
            Status: "pending",
            Results: {},
          }),
        ),
      ),
    );

    let qualifyingOrder = tournament.qualifying_order
      .split(",")
      .map((id) => id.trim())
      .filter(Boolean);

    // only do this if this participant id was previously not qualified
    if (!qualifyingOrder.includes(participantID)) {
      qualifyingOrder.push(participantID);
      updateTournamentMutation.mutate({
        qualifyingOrder,
      });
    }
  }

  let concludeDialogOpen = $state(false);
  const concludeQualifiersMutation = createMutation(() => ({
    mutationFn: async () => ConcludeQualifers(tournament.id),
    onSuccess: async () => {
      await UpdateTournamentState(tournament.id, "start");
      await SyncTournament(tournament.id);

      await queryClient.invalidateQueries({
        queryKey: ["matches", tournament.id],
      });
      await queryClient.invalidateQueries({
        queryKey: ["activeBattles", tournament.id],
      });

      await queryClient.invalidateQueries({
        queryKey: ["tournament", tournament.id],
      });

      await queryClient.invalidateQueries({
        queryKey: ["tournaments"],
      });

      await queryClient.invalidateQueries({
        queryKey: ["logs", tournament.id],
      });

      concludeDialogOpen = false;
      hidePanel = false;
    },
    onError: (error) => {
      console.error("Failed to conclude qualifiers:", error);
    },
  }));

  const orderedEntrants = $derived.by(() => {
    const order = tournament.qualifying_order
      .split(",")
      .map((id) => id.trim())
      .filter(Boolean);

    const orderedIds = new Set(order);

    const byParticipantId = new Map(
      entrants.map((entrant) => [entrant.participantId, entrant]),
    );

    const scoresByParticipantId = new Map(
      (scores.data ?? []).map((data) => [data.participantId, data.scores]),
    );

    const ordered = order
      .map((id) => {
        const entrant = byParticipantId.get(id);

        if (!entrant) return undefined;

        return {
          ...entrant,
          scores: scoresByParticipantId.get(entrant.participantId) ?? [],
        };
      })
      .filter((entrant) => entrant !== undefined);

    const uniqueOrdered = [
      ...new Map(
        ordered.map((entrant) => [entrant.participantId, entrant]),
      ).values(),
    ];

    const unlistedEntrants = entrants
      .filter((entrant) => !orderedIds.has(entrant.participantId))
      .map((entrant) => ({
        ...entrant,
        scores: scoresByParticipantId.get(entrant.participantId) ?? [],
      }));

    let finalOrdered = [...uniqueOrdered, ...unlistedEntrants];

    const allEntrantsProcessed = finalOrdered.every(
      (entrant) =>
        entrant.scores.length > 0 &&
        entrant.scores.every((score) => score.status === "scored"),
    );

    if (tournament.status === "elimination" || allEntrantsProcessed) {
      finalOrdered.sort((a, b) => (b.score ?? 0) - (a.score ?? 0));
    }

    return {
      entrants: finalOrdered,
    };
  });

  const front = $derived(orderedEntrants.entrants[0]?.participantId);

  const concluding = $derived(
    orderedEntrants.entrants.length > 0 &&
      orderedEntrants.entrants.every(
        (entrant) =>
          entrant.scores.length > 0 &&
          entrant.scores.every((score) => score.status === "scored"),
      ),
  );

  $effect(() => {
    hidePanel = !judgement.edit && concluding;
  });
</script>

<div class="space-4">
  <div class="flex items-end justify-between gap-4 pb-4">
    <div>
      <h3 class="font-semibold tracking-tight">Entrants</h3>
      <p class="text-sm text-muted-foreground">Current qualification scores</p>
    </div>
    {#if orderedEntrants.entrants.length > 0}
      <span class="text-xs font-medium text-muted-foreground">
        {orderedEntrants.entrants.length}
        {orderedEntrants.entrants.length === 1 ? "entrant" : "entrants"}
      </span>
    {/if}
  </div>

  {#if orderedEntrants.entrants.length === 0}
    <Empty class="min-h-40 border-dashed py-10">
      <EmptyHeader>
        <EmptyTitle>No entrants yet</EmptyTitle>
        <EmptyDescription>
          Entrants will appear here once they are added to this tournament.
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  {:else}
    <div class="divide-y overflow-hidden rounded-lg border border-border/70">
      {#each orderedEntrants.entrants as entrant, idx (entrant.participantId)}
        <button
          type="button"
          class={[
            "flex w-full items-center justify-between gap-4 px-3 py-3.5 text-left transition-colors sm:px-4",
            judgement.edit?.participantID === entrant.participantId
              ? "border border-amber-200 bg-amber-50/40 text-amber-900 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-50"
              : !judgement.edit && front === entrant.participantId
                ? "bg-accent/50 text-accent-foreground"
                : "hover:bg-muted/50",
          ]}
        >
          <div class="flex min-w-0 items-center gap-3">
            <span
              class="flex size-7 shrink-0 items-center justify-center rounded-md border bg-muted/50 text-xs font-semibold tabular-nums text-muted-foreground"
              aria-label={`Rank ${idx + 1}`}
            >
              {idx + 1}
            </span>

            <div class="min-w-0 space-y-1">
              <div class="items-center flex gap-2">
                <p class="truncate font-medium">{entrant.name}</p>
                {#if entrant.status === "pending"}
                  {#if idx === 0}
                    <Badge variant="default">On Going</Badge>
                  {:else}
                    <Badge variant="outline">Pending</Badge>
                  {/if}
                {:else if entrant.status === "qualified"}
                  <Badge>Qualified</Badge>
                {:else if entrant.status === "not_qualified"}
                  <Badge variant="destructive">Not Qualified</Badge>
                {:else if entrant.status === "scored"}
                  <Badge variant="secondary">Scored</Badge>
                {/if}
              </div>
              {#if tournament.qualifyingRounds > 1}
                <small class="flex gap-1 items-center text-muted-foreground">
                  {#each entrant.scores as data, idx (idx)}
                    <span
                      >Round {idx + 1}: {data.status === "pending"
                        ? "—"
                        : data.score}</span
                    >
                    {#if idx + 1 != entrant.scores.length}
                      <Separator orientation="vertical" />
                    {/if}
                  {/each}
                </small>
              {/if}
            </div>
          </div>

          <div class="flex shrink-0 items-center gap-2 sm:gap-3">
            <span
              class="min-w-10 text-right text-lg font-semibold tabular-nums"
            >
              {entrant.status === "pending" ? "—" : entrant.score}
            </span>

            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div onclick={(event) => event.stopPropagation()}>
              <DropdownMenu>
                <DropdownMenuTrigger>
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label="Qualifier actions"
                  >
                    <EllipsisVertical />
                  </Button>
                </DropdownMenuTrigger>

                <DropdownMenuContent align="end" class="w-56">
                  <DropdownMenuItem
                    onclick={() => moveToFirst(idx)}
                    disabled={entrant.status === "not_qualified" ||
                      updateTournamentMutation.isPending ||
                      idx === 0 ||
                      !tournament.qualifying_order.includes(
                        entrant.participantId,
                      )}
                  >
                    <ArrowUpToLine class="mr-2 h-4 w-4" />
                    Move to front
                  </DropdownMenuItem>

                  {#if entrant.scores?.length}
                    <DropdownMenuSeparator />

                    {#each entrant.scores as score, i (i)}
                      <DropdownMenuItem
                        disabled={score.status === "pending"}
                        onclick={() => {
                          judgement.onJudgementEdit(
                            {
                              id: entrant.id,
                              participantID: entrant.participantId,
                              round: i + 1,
                            },
                            score.results,
                          );
                        }}
                      >
                        <Pencil class="mr-2 h-4 w-4" />

                        <span class="flex-1">
                          Round {i + 1}
                        </span>

                        <span class="text-muted-foreground">
                          {score.status === "pending" ? "Pending" : score.score}
                        </span>
                      </DropdownMenuItem>
                    {/each}
                  {/if}

                  <DropdownMenuSeparator />

                  <DropdownMenuItem
                    onclick={() =>
                      markAs(
                        entrant.id,
                        "not_qualified",
                        entrant.participantId,
                      )}
                    disabled={updateQualifierMutation.isPending ||
                      entrant.status === "not_qualified"}
                    class="text-destructive focus:text-destructive"
                  >
                    <UserX class="mr-2 h-4 w-4" />
                    Set as not qualified
                  </DropdownMenuItem>

                  <DropdownMenuItem
                    onclick={() =>
                      markAs(entrant.id, "pending", entrant.participantId)}
                    disabled={updateQualifierMutation.isPending ||
                      entrant.status === "pending"}
                  >
                    <RotateCcw class="mr-2 h-4 w-4" />
                    Reset
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        </button>
      {/each}
    </div>
  {/if}

  {#if concluding}
    <AlertDialog bind:open={concludeDialogOpen}>
      <AlertDialogTrigger class="w-full pt-4">
        <Button
          class="w-full"
          variant="default"
          disabled={concludeQualifiersMutation.isPending}
        >
          Conclude
        </Button>
      </AlertDialogTrigger>

      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Conclude qualifying?</AlertDialogTitle>

          <AlertDialogDescription>
            After concluding qualifying, the qualified entrants will be
            finalized and will be seeded to the tournament bracket. This action
            cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>

        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>

          <AlertDialogAction
            onclick={() => {
              concludeQualifiersMutation.mutate();
            }}
            disabled={concludeQualifiersMutation.isPending}
          >
            {concludeQualifiersMutation.isPending
              ? "Concluding..."
              : "Conclude"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  {/if}
</div>
