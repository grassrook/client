<script lang="ts">
  import type { Snippet } from "svelte";

  import { Card, CardContent } from "$src/components/ui/card";
  import Empty from "$src/components/ui/empty/empty.svelte";
  import EmptyDescription from "$src/components/ui/empty/empty-description.svelte";
  import EmptyHeader from "$src/components/ui/empty/empty-header.svelte";
  import EmptyMedia from "$src/components/ui/empty/empty-media.svelte";
  import EmptyTitle from "$src/components/ui/empty/empty-title.svelte";
  import Spinner from "$src/components/ui/spinner/spinner.svelte";
  import type { Judgement } from "$src/lib/judging.svelte";
  import type { Stage } from "$src/lib/stage.svelte";
  import type { repository, tournament } from "$wails/go/models";

  import CommonCriteria from "./common-criteria.svelte";
  import CriteriaWrapper from "./criteria-wrapper.svelte";

  let {
    tournament,
    judgement,
    isQualifying,
    stage,
    children,
  }: {
    tournament: tournament.Tournament;
    judgement: Judgement;
    isQualifying: boolean;
    stage: Stage;
    children: Snippet<[]>;
  } = $props();

  const criteriaConfig = $derived(
    JSON.parse(tournament.criteriaConfig) as repository.CriteriaCategories,
  );

  const categories = $derived(Object.entries(criteriaConfig));

  const common = $derived(categories.find(([id]) => id === "common"));
  const lead = $derived(categories.find(([id]) => id === "lead"));
  const chase = $derived(categories.find(([id]) => id === "chase"));
</script>

<div class="relative" aria-busy={judgement.processing}>
  {#if isQualifying || stage.matchId}
    <Card>
      <CardContent>
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
        </div></CardContent
      >
    </Card>
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
