<script lang="ts">
  import { SvelteMap } from "svelte/reactivity";

  import Card from "$src/components/ui/card/card.svelte";
  import CardContent from "$src/components/ui/card/card-content.svelte";
  import type { tournamentactivitylog } from "$wails/go/models";

  let { logs = [] }: { logs: tournamentactivitylog.TournamentActivityLog[] } =
    $props();

  let activeFilter = $state("ALL");

  const filterOptions = $derived([
    "ALL",
    ...new Set(logs.map((log) => log.logType)),
  ]);

  const filteredLogs = $derived(
    logs
      .filter((log) => activeFilter === "ALL" || log.logType === activeFilter)
      .toSorted((a, b) => b.createdAt - a.createdAt),
  );

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

<Card class="gap-0 mb-4">
  <CardContent>
    <section class="w-full space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-base font-semibold">Activity log</h2>
          <p class="text-xs text-muted-foreground">
            Recent tournament activity
          </p>
        </div>

        <span class="text-xs tabular-nums text-muted-foreground">
          {filteredLogs.length} events
        </span>
      </div>

      <div class="flex flex-wrap items-center gap-1 border-b pb-2">
        {#each filterOptions as filter (filter)}
          <button
            type="button"
            aria-pressed={activeFilter === filter}
            class="rounded-md px-2 py-1 text-xs transition-colors
          {activeFilter === filter
              ? 'bg-muted font-medium text-foreground'
              : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'}"
            onclick={() => (activeFilter = filter)}
          >
            {filter === "ALL" ? "All" : formatType(filter)}
          </button>
        {/each}
      </div>

      {#if groupedLogs.length === 0}
        <p class="py-6 text-center text-sm text-muted-foreground">
          No activity found.
        </p>
      {:else}
        <div class="space-y-5">
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
                        <span
                          class="absolute bottom-[-8px] top-2 w-px bg-border"
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
                          datetime={new Date(
                            log.createdAt * 1000,
                          ).toISOString()}
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
  </CardContent>
</Card>
