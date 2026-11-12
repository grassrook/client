<script lang="ts">
  import TriangleAlertIcon from "@lucide/svelte/icons/triangle-alert";
  import { type Snippet, untrack } from "svelte";

  import * as Alert from "$src/components/ui/alert";
  import Empty from "$src/components/ui/empty/empty.svelte";
  import EmptyDescription from "$src/components/ui/empty/empty-description.svelte";
  import EmptyHeader from "$src/components/ui/empty/empty-header.svelte";
  import EmptyMedia from "$src/components/ui/empty/empty-media.svelte";
  import EmptyTitle from "$src/components/ui/empty/empty-title.svelte";
  import Spinner from "$src/components/ui/spinner/spinner.svelte";
  import type { Judgement } from "$src/lib/judging.svelte";
  import type { Stage } from "$src/lib/stage.svelte";
  import type { entrant, repository, tournament } from "$wails/go/models";

  import CommonCriteria from "./common-criteria.svelte";
  import CriteriaWrapper from "./criteria-wrapper.svelte";

  let {
    tournament,
    judgement,
    isQualifying,
    stage,
    children,
    entrants,
  }: {
    tournament: tournament.Tournament;
    judgement: Judgement;
    isQualifying: boolean;
    stage: Stage;
    children: Snippet<[]>;
    entrants: entrant.Entrant[];
  } = $props();

  const criteriaConfig = $derived(
    JSON.parse(
      tournament.status === "qualifying"
        ? (tournament.qualifyingCriteriaConfig ?? tournament.criteriaConfig)
        : tournament.criteriaConfig,
    ) as repository.CriteriaCategories,
  );

  const categories = $derived(Object.entries(criteriaConfig));

  const common = $derived(categories.find(([id]) => id === "common"));
  const lead = $derived(categories.find(([id]) => id === "lead"));
  const chase = $derived(categories.find(([id]) => id === "chase"));

  const weighMaxScore = $derived(
    categories.find(([id]) => id === "weighMaxScore")?.[1] === true,
  );

  const scoreWeighs = $derived(
    categories.find(([id]) => id === "scoreWeighs")?.[1] as
      | repository.ScoreWeighs
      | undefined,
  );

  $effect(() => {
    const config = {
      weighMaxScore,
      scoreWeighs,
      categories: {
        lead: lead?.[1],
        chase: chase?.[1],
        common: common?.[1],
      },
    } as repository.CriteriaConfig;

    untrack(() => {
      judgement.setConfig(config);
    });
  });
</script>

<div class="relative mt-2" aria-busy={judgement.processing}>
  {#if isQualifying || stage.matchId}
    <div class="rounded-xl border text-card-foreground">
      <div class="p-2">
        {#if judgement.edit}
          {@const isElimination = tournament.status === "elimination"}
          {@const leader = entrants.find(
            (e) => e.participantId === stage.getLeader(),
          )}
          {@const chaser = entrants.find(
            (e) => e.participantId === stage.getChaser(),
          )}

          <Alert.Root
            class="mb-2 w-full border-amber-200 bg-amber-50/40 text-amber-900 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-50"
          >
            <TriangleAlertIcon />
            <Alert.Title>You are currently in edit mode</Alert.Title>
            <Alert.Description>
              Changes made in this mode will affect the qualifying or
              elimination results and may alter the final standings.
            </Alert.Description>

            {#if isElimination}
              <div class="col-start-2 mt-3 flex flex-wrap items-center gap-2">
                <span class="text-xs font-medium opacity-70">Editing</span>

                <div
                  class="inline-flex min-w-0 items-stretch overflow-hidden rounded-md border border-sky-300/70 bg-background/60 dark:border-sky-300/25"
                >
                  <span
                    class="flex items-center bg-sky-100 px-2 text-[11px] font-bold tracking-wider text-sky-800 uppercase dark:bg-sky-300/15 dark:text-sky-200"
                  >
                    Lead
                  </span>
                  <span class="truncate px-2.5 py-1 text-sm font-medium">
                    {leader?.name ?? "Unknown"}
                  </span>
                </div>

                <div
                  class="inline-flex min-w-0 items-stretch overflow-hidden rounded-md border border-violet-300/70 bg-background/60 dark:border-violet-300/25"
                >
                  <span
                    class="flex items-center bg-violet-100 px-2 text-[11px] font-bold tracking-wider text-violet-800 uppercase dark:bg-violet-300/15 dark:text-violet-200"
                  >
                    Chase
                  </span>
                  <span class="truncate px-2.5 py-1 text-sm font-medium">
                    {chaser?.name ?? "Unknown"}
                  </span>
                </div>
              </div>
            {/if}
          </Alert.Root>
        {/if}

        <div class="grid min-w-0 grid-cols-2 items-stretch gap-3 md:gap-2">
          {#if common}
            {@const [, category] = common}
            <div class={`h-full ${!isQualifying ? "col-span-2" : ""}`}>
              <CommonCriteria {category} {judgement} {isQualifying} />
            </div>
          {/if}

          {#if lead}
            {@const [, category] = lead}
            <div class="h-full">
              <CriteriaWrapper {category} role="lead" {judgement} />
            </div>
          {/if}

          {#if chase && !isQualifying}
            {@const [, category] = chase}
            <div class="h-full">
              <CriteriaWrapper {category} role="chase" {judgement} />
            </div>
          {/if}
        </div>
        <div class="mt-2">
          {@render children()}
        </div>
      </div>
    </div>
  {/if}

  {#if judgement.processing}
    <div
      class="absolute inset-0 z-50 flex items-center justify-center rounded-xl bg-background/95 px-4"
    >
      <Empty class="w-full max-w-md border-0 bg-transparent shadow-none ring-0">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Spinner />
          </EmptyMedia>

          <EmptyTitle>Processing the scores</EmptyTitle>
          <EmptyDescription>
            Please wait while we process your request. Do not refresh the page.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    </div>
  {/if}
</div>
