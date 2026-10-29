<script lang="ts">
  import { CalendarDays, Clock3, UserRound } from "@lucide/svelte";

  import Separator from "$src/components/ui/separator/separator.svelte";
  import type { driver } from "$wails/go/models";

  type Props = {
    driver: driver.Driver;
  };

  let { driver }: Props = $props();

  function formatDate(timestamp: number) {
    if (!timestamp) return "Not available";

    return new Date(timestamp * 1000).toLocaleDateString(undefined, {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  }
</script>

<section class="overflow-hidden rounded-xl border">
  <div
    class="h-24 bg-gradient-to-r from-muted via-muted/60 to-background"
  ></div>

  <div class="px-6 pb-6">
    <div class="-mt-10 flex flex-wrap items-end justify-between gap-4">
      <div class="flex items-end gap-4">
        <div
          class="flex size-20 items-center justify-center rounded-2xl border-4 border-card bg-muted shadow-sm"
        >
          <UserRound class="size-9 text-muted-foreground" />
        </div>

        <div class="min-w-0 pb-1">
          <h2 class="text-xl font-semibold tracking-tight">
            {driver.name}
          </h2>
          <p class="mt-1 text-sm text-muted-foreground">Driver profile</p>
        </div>
      </div>
    </div>

    <Separator class="my-5" />

    <div class="grid gap-4 sm:grid-cols-2">
      <div class="flex items-start gap-3">
        <div class="rounded-lg border bg-muted/40 p-2">
          <CalendarDays class="size-4 text-muted-foreground" />
        </div>
        <div>
          <p class="text-xs text-muted-foreground">Date added</p>
          <p class="mt-1 text-sm font-medium">
            {formatDate(driver.createdAt)}
          </p>
        </div>
      </div>

      <div class="flex items-start gap-3">
        <div class="rounded-lg border bg-muted/40 p-2">
          <Clock3 class="size-4 text-muted-foreground" />
        </div>
        <div>
          <p class="text-xs text-muted-foreground">Last updated</p>
          <p class="mt-1 text-sm font-medium">
            {formatDate(driver.updatedAt)}
          </p>
        </div>
      </div>
    </div>
  </div>
</section>
