<script lang="ts">
  import type { Judgement } from "$src/lib/judging.svelte";
  import type { repository } from "$wails/go/models";

  import Separator from "../ui/separator/separator.svelte";
  import CriterionRow from "./criterion-row.svelte";

  let {
    category,
    role,
    judgement,
  }: {
    category: repository.CriteriaCategory;
    role: "lead" | "chase";
    judgement: Judgement;
  } = $props();
</script>

<div class="gap-0 rounded-xl border border-border/70 text-card-foreground">
  <div class="flex flex-row items-baseline justify-between gap-3 p-4">
    <div
      class="text-sm min-w-0 truncate font-semibold leading-none tracking-tight"
    >
      {category.label}
    </div>
    <!-- <span class="shrink-0 text-xs font-medium text-muted-foreground capitalize"
      >{role}</span
    > -->
  </div>
  <Separator />
  <div class="space-y-1 px-6 pt-2 pb-6">
    {#each category.tags as tag, index (tag.id)}
      <CriterionRow
        {tag}
        value={judgement.get(tag.id)?.[role]}
        showSeparator={index > 0}
        onValueChange={(value) => judgement.set(tag.id, role, value)}
      />
    {/each}
  </div>
</div>
