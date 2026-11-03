<script lang="ts">
  type Partition = {
    id: string;
    label: string;
    size: number;
    minSize?: number;
    color?: string;
  };

  let {
    parts = $bindable(),
    onResize,
  }: {
    parts: Partition[];
    onResize?: (parts: Partition[]) => void;
  } = $props();

  let containerWidth = $state(0);
  let dragging = $state<number | null>(null);

  const totalSize = $derived(parts.reduce((sum, part) => sum + part.size, 0));

  function startDrag(index: number, event: PointerEvent) {
    if (!event.isPrimary) return;

    dragging = index;

    const target = event.currentTarget as HTMLElement;
    target.setPointerCapture(event.pointerId);

    target.dataset.startX = String(event.clientX);
    target.dataset.startLeft = String(parts[index].size);
    target.dataset.startRight = String(parts[index + 1].size);
  }

  function moveDrag(event: PointerEvent) {
    if (dragging === null || containerWidth <= 0) return;

    const target = event.currentTarget as HTMLElement;
    const startX = Number(target.dataset.startX);
    const startLeft = Number(target.dataset.startLeft);
    const startRight = Number(target.dataset.startRight);

    const delta = ((event.clientX - startX) / containerWidth) * totalSize;

    const left = parts[dragging];
    const right = parts[dragging + 1];

    if (!left || !right) return;

    const minLeft = left.minSize ?? 1;
    const minRight = right.minSize ?? 1;

    const actualDelta = Math.round(
      Math.max(minLeft - startLeft, Math.min(startRight - minRight, delta)),
    );

    parts = parts.map((part, i) => {
      if (i === dragging) {
        return { ...part, size: startLeft + actualDelta };
      }

      if (dragging && i === dragging + 1) {
        return { ...part, size: startRight - actualDelta };
      }

      return part;
    });

    onResize?.(parts);
  }

  function endDrag() {
    dragging = null;
  }

  const partitionColors = [
    "hsl(var(--muted-foreground) / 0.35)",
    "hsl(var(--muted-foreground) / 0.65)",
    "hsl(var(--muted-foreground) / 0.2)",
  ];
</script>

<div class="w-full space-y-3">
  <div
    class="relative flex h-12 w-full select-none overflow-hidden rounded-md border bg-muted"
    bind:clientWidth={containerWidth}
  >
    {#each parts as part, i (i)}
      <div
        class="relative flex h-full min-w-0 items-center justify-center overflow-hidden text-xs font-medium text-white"
        style="
width: {totalSize > 0 ? (part.size / totalSize) * 100 : 0}%;
          background-color: {part.color ??
          partitionColors[i % partitionColors.length]};
        "
        title={`${part.label}: ${part.size} GB`}
      >
        <span class="truncate px-1">
          {part.label}
        </span>
      </div>

      {#if i < parts.length - 1}
        <div class="group relative z-10 w-0 shrink-0">
          <div
            class="absolute -left-3 top-0 flex h-full w-6 cursor-col-resize touch-none items-center justify-center"
            role="separator"
            aria-orientation="vertical"
            aria-label={`Resize ${part.label} and ${parts[i + 1].label}`}
            onpointerdown={(e) => startDrag(i, e)}
            onpointermove={moveDrag}
            onpointerup={endDrag}
            onpointercancel={endDrag}
          >
            <div
              class="flex h-9 w-3 items-center justify-center rounded-full border border-border bg-card shadow-sm transition-all group-hover:border-muted-foreground/50 group-hover:bg-muted group-active:scale-95"
            >
              <div class="flex flex-col gap-[3px]">
                <span class="h-[2px] w-[2px] rounded-full bg-muted-foreground"
                ></span>
                <span class="h-[2px] w-[2px] rounded-full bg-muted-foreground"
                ></span>
                <span class="h-[2px] w-[2px] rounded-full bg-muted-foreground"
                ></span>
              </div>
            </div>
          </div>
        </div>
      {/if}
    {/each}
  </div>

  <div class="flex flex-wrap gap-x-4 gap-y-2">
    {#each parts as part, i (i)}
      <div class="flex items-center gap-2 text-sm">
        <span
          class="size-2.5 rounded-sm"
          style="background-color: {part.color ??
            partitionColors[i % partitionColors.length]}"
        ></span>
        <span class="text-muted-foreground">{part.label}</span>
        <span class="font-medium tabular-nums"
          >{Math.floor(Math.max(0, part.size))}</span
        >
      </div>
    {/each}
  </div>
</div>
