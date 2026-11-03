<script lang="ts">
  import ChevronDown from "@lucide/svelte/icons/chevron-down";
  import { createMutation, useQueryClient } from "@tanstack/svelte-query";
  import { untrack } from "svelte";

  import { Button } from "$src/components/ui/button";
  import * as Collapsible from "$src/components/ui/collapsible";
  import * as Dialog from "$src/components/ui/dialog";
  import { Dropzone } from "$src/components/ui/dropzone";
  import { Input } from "$src/components/ui/input";
  import Label from "$src/components/ui/label/label.svelte";
  import * as Select from "$src/components/ui/select";
  import * as Switch from "$src/components/ui/switch";
  import {
    CreateCriteriaConfig,
    UpdateCriteriaConfig,
  } from "$wails/go/main/App";
  import { repository } from "$wails/go/models";

  import Separator from "../ui/separator/separator.svelte";
  import CategoryHeader from "./category-header.svelte";
  import Partition from "./partition.svelte";

  type CategoryKey = "common" | "lead" | "chase";
  type TagType = "neutral" | "deduction";
  type ValueType = "dynamic" | "switch";
  type Mode = "create" | "update";

  type TagForm = {
    id?: string;
    label: string;
    type: TagType;
    valueType: ValueType;
    max: number;
    multiplier: number;
    defaultValue: number;
  };

  type CategoryForm = {
    label: string;
    tags: TagForm[];
  };

  type Props = {
    mode: Mode;
    config?: repository.CriteriaConfig;
    open?: boolean;
  };

  let { mode, config, open = $bindable(false) }: Props = $props();

  const queryClient = useQueryClient();

  function defaultCategories(): Record<CategoryKey, CategoryForm> {
    return {
      common: { label: "Style Criteria", tags: [] },
      lead: { label: "Lead Criteria", tags: [] },
      chase: { label: "Chase Criteria", tags: [] },
    };
  }

  let categories = $state(defaultCategories());
  let newConfigName = $state("");
  let formError = $state("");

  let weighMaxScore = $state(false);

  let leadAllocation = $state<repository.WeighAllocation>({
    maxScore: 100,
    commonScore: 50,
  });

  let chaseAllocation = $state<repository.WeighAllocation>({
    maxScore: 100,
    commonScore: 50,
  });

  let leadMaxScoreInput = $state(100);
  let chaseMaxScoreInput = $state(100);

  const leadParts = $derived([
    {
      id: "common",
      label: "Common",
      size: leadAllocation.commonScore,
      minSize: 0,
      color: "#3b82f6",
    },
    {
      id: "lead",
      label: "Lead",
      size: leadAllocation.maxScore - leadAllocation.commonScore,
      minSize: 0,
      color: "#22c55e",
    },
  ]);

  const chaseParts = $derived([
    {
      id: "common",
      label: "Common",
      size: chaseAllocation.commonScore,
      minSize: 0,
      color: "#3b82f6",
    },
    {
      id: "chase",
      label: "Chase",
      size: chaseAllocation.maxScore - chaseAllocation.commonScore,
      minSize: 0,
      color: "#eab308",
    },
  ]);

  function updateLeadScore(parts: { size: number }[]) {
    const common = parts[0]?.size;

    if (common === undefined) return;

    leadAllocation.commonScore = Math.max(
      0,
      Math.min(leadAllocation.maxScore, Math.round(common)),
    );
  }

  function updateChaseScore(parts: { size: number }[]) {
    const common = parts[0]?.size;

    if (common === undefined) return;

    chaseAllocation.commonScore = Math.max(
      0,
      Math.min(chaseAllocation.maxScore, Math.round(common)),
    );
  }

  function resetForm() {
    newConfigName = "";
    categories = defaultCategories();
    formError = "";

    weighMaxScore = false;

    leadAllocation = {
      maxScore: 100,
      commonScore: 50,
    };

    chaseAllocation = {
      maxScore: 100,
      commonScore: 50,
    };

    leadMaxScoreInput = 100;
    chaseMaxScoreInput = 100;
  }

  type CategoryInput = {
    label: string;
    tags: {
      id?: string;
      label: string;
      type: string;
      valueType: string;
      max?: number | null;
      multiplier?: number | null;
      defaultValue?: number | null;
    }[];
  };

  function toFormCategory(category: CategoryInput): CategoryForm {
    return {
      label: category.label,
      tags: category.tags.map((tag) => ({
        id: tag.id,
        label: tag.label,
        type: tag.type as TagType,
        valueType: tag.valueType as ValueType,
        max: Number(tag.max ?? 0),
        multiplier: Number(tag.multiplier ?? 1),
        defaultValue: Number(tag.defaultValue ?? 0),
      })),
    };
  }

  function loadConfig(value: repository.CriteriaConfig) {
    newConfigName = value.name;
    categories = {
      common: toFormCategory(value.categories.common),
      lead: toFormCategory(value.categories.lead),
      chase: toFormCategory(value.categories.chase),
    };

    weighMaxScore = value.weighMaxScore ?? false;

    leadAllocation = {
      maxScore: value.scoreWeighs?.lead?.maxScore || 100,
      commonScore: value.scoreWeighs?.lead?.commonScore || 50,
    };

    chaseAllocation = {
      maxScore: value.scoreWeighs?.chase?.maxScore || 100,
      commonScore: value.scoreWeighs?.chase?.commonScore || 50,
    };

    leadMaxScoreInput = leadAllocation.maxScore;
    chaseMaxScoreInput = chaseAllocation.maxScore;

    formError = "";
  }

  let openTagKey = $state<string | undefined>(undefined);

  $effect(() => {
    if (!open) return;

    if (mode === "update" && config) {
      untrack(() => {
        openTagKey = undefined;
        loadConfig(config);
      });
    } else {
      resetForm();
    }
  });

  async function addTag(category: CategoryKey) {
    const tags = categories[category].tags;
    const newIndex = tags.length;

    openTagKey = `${category}-${newIndex}`;

    tags.push({
      label: "",
      type: "neutral",
      valueType: "dynamic",
      max: 0,
      multiplier: 1,
      defaultValue: 0,
    });
  }

  function removeTag(category: CategoryKey, index: number) {
    categories[category].tags.splice(index, 1);
  }

  function buildConfig(): repository.CriteriaConfig {
    const toCategory = (key: CategoryKey): repository.CriteriaCategory => {
      const category = categories[key];

      return new repository.CriteriaCategory({
        label: category.label,
        tags: category.tags.map(
          (tag) =>
            new repository.CriteriaTag({
              id: tag.id ?? `${key}_${crypto.randomUUID()}`,
              label: tag.label.trim(),
              type: tag.type,
              valueType: tag.valueType,
              max: tag.max,
              multiplier: tag.multiplier,
              defaultValue: tag.defaultValue,
            }),
        ),
        weighMaxScore,
        scoreWeighs: new repository.ScoreWeighs({
          lead: new repository.WeighAllocation({
            maxScore: leadAllocation.maxScore,
            commonScore: leadAllocation.commonScore,
          }),
          chase: new repository.WeighAllocation({
            maxScore: chaseAllocation.maxScore,
            commonScore: chaseAllocation.commonScore,
          }),
        }),
      });
    };

    return new repository.CriteriaConfig({
      id: mode === "update" && config ? config.id : 0,
      name: newConfigName.trim(),
      categories: new repository.CriteriaCategories({
        common: toCategory("common"),
        lead: toCategory("lead"),
        chase: toCategory("chase"),
      }),
      weighMaxScore,
      scoreWeighs: new repository.ScoreWeighs({
        lead: new repository.WeighAllocation({
          maxScore: leadAllocation.maxScore,
          commonScore: leadAllocation.commonScore,
        }),
        chase: new repository.WeighAllocation({
          maxScore: chaseAllocation.maxScore,
          commonScore: chaseAllocation.commonScore,
        }),
      }),
    });
  }

  const saveMutation = createMutation(() => ({
    mutationFn: async () => {
      const value = buildConfig();

      if (mode === "update") {
        if (!config) {
          throw new Error("No criteria config selected for updating.");
        }

        return UpdateCriteriaConfig(config.id, value);
      }

      return CreateCriteriaConfig(value);
    },

    onSuccess: async () => {
      open = false;
      resetForm();

      await queryClient.invalidateQueries({
        queryKey: ["criteria-configs"],
      });
    },
  }));

  function submitConfig() {
    if (!newConfigName.trim()) {
      formError = "Enter a config name.";
      return;
    }

    for (const key of ["common", "lead", "chase"] as const) {
      const category = categories[key];

      if (category.tags.some((tag) => !tag.label.trim())) {
        formError = `Every tag in ${category.label} needs a label.`;
        return;
      }
    }

    formError = "";
    saveMutation.mutate();
  }

  function copyCategory(source: "lead" | "chase", target: "lead" | "chase") {
    const sourceCategory = categories[source];

    categories[target] = {
      label: sourceCategory.label,
      tags: sourceCategory.tags.map((tag) => ({ ...tag })),
    };
  }

  function isValidConfig(
    value: unknown,
  ): value is Record<CategoryKey, CategoryForm> {
    if (typeof value !== "object" || value === null) return false;

    const categories = value as Record<string, unknown>;
    const requiredKeys: CategoryKey[] = ["common", "lead", "chase"];

    return requiredKeys.every((key) => {
      const category = categories[key];

      if (typeof category !== "object" || category === null) return false;

      const { label, tags } = category as Record<string, unknown>;

      return typeof label === "string" && Array.isArray(tags);
    });
  }

  function handleFilesChange(files: File[]) {
    formError = "";

    const current = files[0];
    if (!current) return;

    const reader = new FileReader();

    reader.onload = () => {
      try {
        const parsed: unknown = JSON.parse(reader.result as string);

        if (!isValidConfig(parsed)) {
          formError = "Invalid JSON config format";
          return;
        }

        categories = {
          common: toFormCategory(parsed.common),
          lead: toFormCategory(parsed.lead),
          chase: toFormCategory(parsed.chase),
        };
      } catch {
        formError = "Invalid JSON syntax";
      }
    };

    reader.readAsText(current);
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-h-[85vh] overflow-y-auto sm:max-w-3xl">
    <Dialog.Header>
      <Dialog.Title>
        {mode === "update"
          ? "Update Criteria Config"
          : "Create Criteria Config"}
      </Dialog.Title>
      <Dialog.Description>
        Define your judging categories and their scoring tags.
      </Dialog.Description>
    </Dialog.Header>

    <form
      class="space-y-4"
      onsubmit={(event) => {
        event.preventDefault();
        submitConfig();
      }}
    >
      <div class="space-y-2">
        <Label for="config-name">Config Name</Label>
        <Input
          id="config-name"
          bind:value={newConfigName}
          placeholder="e.g. Default Criteria"
          required
        />
      </div>

      {#if mode === "create"}
        <Dropzone
          accept=".json"
          showFileList={false}
          maxSizeMB={1}
          multiple={false}
          autoUpload={true}
          onfileschange={handleFilesChange}
        />

        <div class="flex items-center my-4 w-full">
          <Separator class="flex-1" />
          <span class="mx-4 shrink-0 text-xs uppercase text-muted-foreground"
            >Or create manually</span
          >
          <Separator class="flex-1" />
        </div>
      {/if}

      {#each [{ key: "common", title: "Common", category: categories.common }, { key: "lead", title: "Lead", category: categories.lead }, { key: "chase", title: "Chase", category: categories.chase }] as section (section.key)}
        <section class="space-y-4 rounded-lg border p-4">
          <CategoryHeader
            title={section.title}
            categoryKey={section.key as CategoryKey}
            tagCount={section.category.tags.length}
            onAddTag={addTag}
            onCopyCategory={copyCategory}
          />

          <div class="space-y-2">
            <!-- <Label>Category Label</Label> -->
            <Input type="hidden" bind:value={section.category.label} required />
          </div>
          {#each section.category.tags as tag, index (`${section.key}-${index}`)}
            <Collapsible.Root
              open={openTagKey === `${section.key}-${index}`}
              onOpenChange={(open) => {
                openTagKey = open ? `${section.key}-${index}` : undefined;
              }}
              class="rounded-md border bg-muted/40 p-3"
            >
              <div class="flex items-center justify-between gap-2">
                <Collapsible.Trigger
                  class="group flex min-w-0 flex-1 items-center justify-between gap-2 text-left"
                >
                  <div class="min-w-0">
                    <p class="truncate text-sm font-medium">
                      {tag.label.trim() || `Tag ${index + 1}`}
                    </p>
                    <p class="text-muted-foreground text-xs">
                      {tag.type} · {tag.valueType}
                    </p>
                  </div>

                  <ChevronDown
                    class="size-4 shrink-0 transition-transform group-data-[state=open]:rotate-180"
                  />
                </Collapsible.Trigger>

                <Button
                  type="button"
                  variant="destructive"
                  size="sm"
                  onclick={() => removeTag(section.key as CategoryKey, index)}
                >
                  Remove
                </Button>
              </div>

              <Collapsible.Content>
                <div class="grid gap-3 pt-4 sm:grid-cols-2">
                  <div class="space-y-2 sm:col-span-2">
                    <Label>Tag Label</Label>
                    <Input
                      bind:value={tag.label}
                      placeholder="e.g. Angle"
                      required
                    />
                  </div>

                  <div class="space-y-2">
                    <Label>Type</Label>
                    <Select.Root type="single" bind:value={tag.type}>
                      <Select.Trigger class="w-full">
                        {tag.type === "neutral" ? "Neutral" : "Deduction"}
                      </Select.Trigger>
                      <Select.Content>
                        <Select.Item value="neutral">Neutral</Select.Item>
                        <Select.Item value="deduction">Deduction</Select.Item>
                      </Select.Content>
                    </Select.Root>
                  </div>

                  <div class="space-y-2">
                    <Label>Value Type</Label>
                    <Select.Root type="single" bind:value={tag.valueType}>
                      <Select.Trigger class="w-full">
                        {tag.valueType === "dynamic" ? "Dynamic" : "Switch"}
                      </Select.Trigger>
                      <Select.Content>
                        <Select.Item value="dynamic">Dynamic</Select.Item>
                        <Select.Item value="switch">Switch</Select.Item>
                      </Select.Content>
                    </Select.Root>
                  </div>

                  {#if !(tag.type === "deduction" && tag.valueType === "switch")}
                    <div class="space-y-2">
                      <Label>Max Value</Label>
                      <Input
                        type="number"
                        min="0"
                        step="any"
                        bind:value={tag.max}
                      />
                    </div>

                    <div class="space-y-2">
                      <Label>Multiplier</Label>
                      <Input
                        type="number"
                        min="0"
                        step="any"
                        bind:value={tag.multiplier}
                      />
                    </div>
                  {/if}

                  <div class="space-y-2 sm:col-span-2">
                    <Label>Default Value</Label>
                    <Input
                      type="number"
                      step="any"
                      bind:value={tag.defaultValue}
                    />
                  </div>
                </div>
              </Collapsible.Content>
            </Collapsible.Root>
          {:else}
            <p class="text-muted-foreground py-4 text-center text-sm">
              No tags in this category yet.
            </p>
          {/each}
        </section>
      {/each}

      <section class="space-y-4 rounded-lg border p-4">
        <div class="flex items-center justify-between gap-4">
          <div class="space-y-1">
            <Label>Force score weighting</Label>
            <p class="text-sm text-muted-foreground">
              Keep the final judging score within the new maximum score. Scores
              on every category are converted proportionally to fit the new
              maximum score.
            </p>
          </div>

          <Switch.Root bind:checked={weighMaxScore} />
        </div>

        {#if weighMaxScore}
          <section class="space-y-6 rounded-lg border p-4">
            <!-- Lead -->
            <div class="space-y-3">
              <div class="space-y-1">
                <Label>Lead Score</Label>
                <p class="text-sm text-muted-foreground">
                  Configure the maximum score allocation for Lead.
                </p>
              </div>

              <Input
                type="number"
                min="0"
                step="1"
                bind:value={leadMaxScoreInput}
                onblur={() => {
                  const value = Math.max(
                    0,
                    Math.floor(Number(leadMaxScoreInput) || 0),
                  );

                  leadAllocation.maxScore = value;
                  leadAllocation.commonScore = Math.min(
                    leadAllocation.commonScore,
                    value,
                  );

                  leadMaxScoreInput = value;
                }}
              />

              <Partition parts={leadParts} onResize={updateLeadScore} />
            </div>

            <Separator />

            <!-- Chase -->
            <div class="space-y-3">
              <div class="space-y-1">
                <Label>Chase Score</Label>
                <p class="text-sm text-muted-foreground">
                  Configure the maximum score allocation for Chase.
                </p>
              </div>

              <Input
                type="number"
                min="0"
                step="1"
                bind:value={chaseMaxScoreInput}
                onblur={() => {
                  const value = Math.max(
                    0,
                    Math.floor(Number(chaseMaxScoreInput) || 0),
                  );

                  chaseAllocation.maxScore = value;
                  chaseAllocation.commonScore = Math.min(
                    chaseAllocation.commonScore,
                    value,
                  );

                  chaseMaxScoreInput = value;
                }}
              />

              <Partition parts={chaseParts} onResize={updateChaseScore} />
            </div>
          </section>
        {/if}
      </section>

      {#if formError}
        <p class="text-destructive text-sm">{formError}</p>
      {/if}

      {#if saveMutation.isError}
        <p class="text-destructive text-sm">
          Failed to {mode === "update" ? "update" : "create"} config:
          {String(saveMutation.error)}
        </p>
      {/if}

      <Dialog.Footer>
        <Button
          type="button"
          variant="outline"
          onclick={() => (open = false)}
          disabled={saveMutation.isPending}
        >
          Cancel
        </Button>

        <Button type="submit" disabled={saveMutation.isPending}>
          {#if saveMutation.isPending}
            {mode === "update" ? "Updating..." : "Creating..."}
          {:else}
            {mode === "update" ? "Save Changes" : "Create Config"}
          {/if}
        </Button>
      </Dialog.Footer>
    </form>
  </Dialog.Content>
</Dialog.Root>
