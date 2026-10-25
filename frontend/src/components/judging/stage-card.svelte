<script lang="ts">
  import type { Snippet } from "svelte";

  let {
    children,
  }: {
    children: Snippet<[]>;
  } = $props();
</script>

<div class="gap-0 border rounded-xl border-border/70 shadow-none">
  <div class="p-4 space-y-4">
    {@render children()}
  </div>
</div>
