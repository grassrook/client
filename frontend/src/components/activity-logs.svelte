<script lang="ts">
  import { SvelteMap, SvelteSet } from "svelte/reactivity";

  import type { tournamentactivitylog } from "$wails/go/models";

  import Separator from "./ui/separator/separator.svelte";

  let { logs = [] }: { logs: tournamentactivitylog.TournamentActivityLog[] } =
    $props();

  // the log types the user has toggled on; empty means "show everything"
  const selectedFilters = new SvelteSet<string>();

  const typeOptions = $derived([...new Set(logs.map((log) => log.logType))]);

  // ignore selections for types that no longer exist in the current logs
  const activeFilters = $derived(
    typeOptions.filter((type) => selectedFilters.has(type)),
  );

  const filteredLogs = $derived(
    logs
      .filter(
        (log) =>
          activeFilters.length === 0 || activeFilters.includes(log.logType),
      )
      .toSorted((a, b) => b.createdAt - a.createdAt),
  );

  function toggleFilter(type: string) {
    if (selectedFilters.has(type)) {
      selectedFilters.delete(type);
    } else {
      selectedFilters.add(type);
    }
  }

  function clearFilters() {
    selectedFilters.clear();
  }

  const groupedLogs = $derived.by(() => {
    const groups = new SvelteMap<
      string,
      { label: string; logs: tournamentactivitylog.TournamentActivityLog[] }
    >();

    for (const log of filteredLogs) {
      const date = new Date(log.createdAt * 1000);
      const key = [date.getFullYear(), date.getMonth(), date.getDate()].join(
        "-",
      );

      if (!groups.has(key)) {
        groups.set(key, {
          label: date.toLocaleDateString(undefined, {
            weekday: "long",
            month: "long",
            day: "numeric",
            year: "numeric",
          }),
          logs: [],
        });
      }

      groups.get(key)!.logs.push(log);
    }

    return Array.from(groups.entries()).map(([key, group]) => ({
      key,
      ...group,
    }));
  });

  function formatType(type: string) {
    return type
      .toLowerCase()
      .split("_")
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(" ");
  }

  function formatTime(timestamp: number) {
    return new Date(timestamp * 1000).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    });
  }
</script>

<section class="w-full space-y-4">
  <div class="flex items-center justify-between px-4 pt-4">
    <div>
      <h2 class="text-base font-semibold">Activity log</h2>
      <p class="text-xs text-muted-foreground">Recent tournament activity</p>
    </div>

    <span class="text-xs tabular-nums text-muted-foreground">
      {filteredLogs.length} events
    </span>
  </div>

  <div
    class="flex flex-wrap items-center gap-1 px-4"
    role="group"
    aria-label="Filter by activity type"
  >
    <!-- "All" clears the selection -->
    <button
      type="button"
      aria-pressed={activeFilters.length === 0}
      class="rounded-md px-2 py-1 text-xs transition-colors
          {activeFilters.length === 0
        ? 'bg-muted font-medium text-foreground'
        : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'}"
      onclick={clearFilters}
    >
      All
    </button>

    {#each typeOptions as type (type)}
      {@const selected = selectedFilters.has(type)}
      <button
        type="button"
        aria-pressed={selected}
        class="rounded-md px-2 py-1 text-xs transition-colors
            {selected
          ? 'bg-muted font-medium text-foreground'
          : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'}"
        onclick={() => toggleFilter(type)}
      >
        {formatType(type)}
      </button>
    {/each}
  </div>

  <Separator />

  {#if groupedLogs.length === 0}
    <p class="px-4 py-6 text-center text-sm text-muted-foreground">
      No activity found.
    </p>
  {:else}
    <div
      class="max-h-[40vh] space-y-5 overflow-y-auto overscroll-contain px-4 pb-4"
    >
      {#each groupedLogs as group (group.key)}
        <div class="space-y-1">
          <h3 class="pb-1 text-xs font-medium text-muted-foreground">
            {group.label}
          </h3>

          <div class="relative">
            {#each group.logs as log, i (log.id)}
              <div class="relative flex gap-3 py-2">
                <div class="relative flex w-2 shrink-0 justify-center">
                  {#if i < group.logs.length - 1}
                    <span class="absolute bottom-[-24px] top-2 w-px bg-border"
                    ></span>
                  {/if}

                  <span
                    class="relative mt-1.5 size-1.5 shrink-0 rounded-full bg-muted-foreground/60"
                  ></span>
                </div>

                <div class="min-w-0 flex-1">
                  <p class="break-words text-sm leading-5">
                    {log.content}
                  </p>

                  <p class="mt-0.5 text-xs text-muted-foreground">
                    {formatType(log.logType)}
                    <span class="px-1">·</span>
                    <time
                      datetime={new Date(log.createdAt * 1000).toISOString()}
                    >
                      {formatTime(log.createdAt)}
                    </time>
                  </p>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</section>
<Separator />
