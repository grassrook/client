<script lang="ts">
  import { ArrowLeft } from "@lucide/svelte";
  import { push } from "svelte-spa-router";

  import { Button } from "$src/components/ui/button";
  import Separator from "$src/components/ui/separator/separator.svelte";
</script>

<div
  class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-6 sm:px-6 lg:px-8 lg:py-8"
>
  <div class="flex items-center gap-3">
    <Button
      variant="ghost"
      size="icon"
      onclick={() => push("/drivers")}
      aria-label="Back to drivers"
    >
      <ArrowLeft class="size-5" />
    </Button>

    <div>
      <p class="text-sm text-muted-foreground">Driver directory</p>
      <h1 class="text-2xl font-semibold tracking-tight">Driver profile</h1>
    </div>
  </div>
</div>

<Separator />
