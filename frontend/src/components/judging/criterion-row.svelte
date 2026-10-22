<script lang="ts">
  import { Badge } from "$src/components/ui/badge";
  import type { repository } from "$wails/go/models";

  import Criteriacontrol from "./criteria-control.svelte";

  let {
    tag,
    value,
    onValueChange,
  }: {
    tag: repository.CriteriaTag;
    value: number | undefined;
    onValueChange: (value: number) => void;
    showSeparator?: boolean;
  } = $props();

  function getBadgeVariant(type: string) {
    if (type === "bonus") return "default";
    if (type === "deduction") return "destructive";
    return "secondary";
  }
</script>

<div
  class="flex gap-3 py-3"
  class:flex-col={tag.valueType !== "switch"}
  class:justify-between={tag.valueType === "switch"}
  class:items-center={tag.valueType === "switch"}
>
  <div class="min-w-0">
    <div class="flex flex-wrap items-center gap-x-3 gap-y-1">
      <span class="text-sm leading-5 font-medium">
        {tag.label}
      </span>

      {#if tag.max !== undefined}
        <span class="shrink-0 text-xs text-muted-foreground tabular-nums">
          {Math.abs((value ?? 0) / (tag.multiplier ?? 1))}/{(tag.max ?? 0) /
            (tag.multiplier ?? 1)}
        </span>
      {/if}

      {#if tag.type !== "neutral"}
        <Badge
          variant={getBadgeVariant(tag.type)}
          class="shrink-0 text-xs capitalize"
        >
          {tag.type}
        </Badge>
      {/if}
    </div>
  </div>

  <Criteriacontrol
    {tag}
    {value}
    onValueChange={(value) => onValueChange(Number(value))}
  />
</div>
