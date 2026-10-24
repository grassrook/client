<script lang="ts">
  import type { Snippet } from "svelte";

  import Card from "../ui/card/card.svelte";
  import CardContent from "../ui/card/card-content.svelte";

  let hasRoundedBottom = $state(true);

  $effect(() => {
    if (isQualifying) {
      hasRoundedBottom = false;
    }
  });

  let {
    isQualifying,
    stageContent,
  }: {
    isQualifying: boolean;
    stageContent: Snippet<
      [
        {
          hasRoundedBottom: boolean;
          setHasRoundedBottom: (value: boolean) => void;
        },
      ]
    >;
  } = $props();
</script>

<Card
  class={`gap-0 border-border/70 shadow-none ${
    hasRoundedBottom ? "rounded-b-xl border-b" : "rounded-b-none border-b-0"
  }`}
>
  <CardContent class="space-y-4">
    {@render stageContent({
      hasRoundedBottom,
      setHasRoundedBottom: (value) => (hasRoundedBottom = value),
    })}
  </CardContent>
</Card>
