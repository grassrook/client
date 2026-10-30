<script lang="ts">
  import { Save } from "@lucide/svelte";

  import { Button } from "$src/components/ui/button";
  import { Input } from "$src/components/ui/input";
  import { Label } from "$src/components/ui/label";
  import { Textarea } from "$src/components/ui/textarea";

  type Props = {
    name: string;
    description: string;
    errorMessage: string;
    successMessage: string;
    saving: boolean;
    onSubmit: (event: SubmitEvent) => void;
  };

  let {
    name = $bindable(),
    description = $bindable(),
    errorMessage,
    successMessage,
    saving,
    onSubmit,
  }: Props = $props();
</script>

<section class="rounded-xl border">
  <div class="border-b px-6 py-4">
    <h2 class="font-semibold">Personal information</h2>
    <p class="mt-1 text-sm text-muted-foreground">
      Update this driver's profile information.
    </p>
  </div>

  <form class="space-y-5 p-6" onsubmit={onSubmit}>
    <div class="space-y-2">
      <Label for="driver-name">Driver name</Label>
      <Input
        id="driver-name"
        bind:value={name}
        placeholder="Enter driver name"
        maxlength={255}
        required
      />
    </div>

    <div class="space-y-2">
      <Label for="driver-description">Description</Label>
      <Textarea
        id="driver-description"
        bind:value={description}
        placeholder="Add information about this driver..."
        rows={5}
      />
      <p class="text-xs text-muted-foreground">
        Optional. Use this field for notes or other relevant details.
      </p>
    </div>

    {#if errorMessage}
      <p class="text-sm text-destructive">{errorMessage}</p>
    {/if}

    {#if successMessage}
      <p class="text-sm text-green-600 dark:text-green-400">
        {successMessage}
      </p>
    {/if}

    <div class="flex justify-end border-t pt-5">
      <Button
        class="bg-blue-600 text-white hover:bg-blue-700"
        type="submit"
        disabled={saving}
      >
        <Save class="mr-2 size-4" />
        {saving ? "Saving changes..." : "Save changes"}
      </Button>
    </div>
  </form>
</section>
