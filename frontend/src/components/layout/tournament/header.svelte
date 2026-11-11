<script lang="ts">
  import { LucideArrowLeft } from "@lucide/svelte";
  import { link } from "svelte-spa-router";

  import Progress from "$src/components/ui/progress/progress.svelte";

  type Props = {
    name?: string;
    progress?: number;
  };

  const { name, progress }: Props = $props();
</script>

<div
  class="sticky top-0 z-50 mx-auto w-full bg-background backdrop-blur supports-backdrop-filter:bg-background/80"
>
  <header
    class="flex gap-5 px-4 pb-6 sm:flex-row sm:items-end sm:justify-between sm:px-6 lg:px-8 py-4"
  >
    <div class="flex flex-col gap-5">
      <div class="space-y-2">
        <a
          use:link
          href="/"
          class="inline-flex items-center gap-1 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          <LucideArrowLeft class="size-4" />
          Back to tournament list
        </a>

        <h1 class="text-3xl font-semibold tracking-tight">
          {name ?? "Loading..."}
        </h1>
      </div>
    </div>
  </header>

  <Progress class="rounded-none" value={progress} />
</div>
