<script lang="ts">
  import * as Sidebar from "$src/components/ui/sidebar";
  import * as Collapsible from "$src/components/ui/collapsible";
  import { link } from "svelte-spa-router";
  import active from "svelte-spa-router/active";

  import { createQuery } from "@tanstack/svelte-query";
  import { GetTournaments } from "$wails/go/main/App";

  import Trophy from "@lucide/svelte/icons/trophy";
  import ChevronRight from "@lucide/svelte/icons/chevron-right";

  const tournamentsQuery = createQuery(() => ({
    queryKey: ["tournaments"],
    queryFn: GetTournaments,
  }));

  let eliminationTournaments = $derived(
    (tournamentsQuery.data ?? []).filter(
      (tournament) => tournament.status === "elimination",
    ),
  );
</script>

<Collapsible.Root>
  <Sidebar.MenuItem>
    <Collapsible.Trigger>
      {#snippet child({ props })}
        <Sidebar.MenuButton {...props} isActive={false}>
          <Trophy aria-hidden="true" />
          <span>Tournaments</span>

          {#if eliminationTournaments.length > 0}
            <ChevronRight
              class="ml-auto transition-transform duration-200 [[data-state=open]>&]:rotate-90"
              aria-hidden="true"
            />
          {/if}
        </Sidebar.MenuButton>
      {/snippet}
    </Collapsible.Trigger>

    <Collapsible.Content>
      <Sidebar.MenuSub>
        {#if eliminationTournaments.length > 0}
          {#each eliminationTournaments as tournament (tournament.id)}
            <Sidebar.MenuSubItem>
              <Sidebar.MenuSubButton>
                {#snippet child({ props })}
                  <a
                    {...props}
                    href="/tournaments/{tournament.id}"
                    use:link
                    use:active={{
                      path: `/tournaments/${tournament.id}`,
                      className:
                        "bg-sidebar-accent text-sidebar-accent-foreground",
                      inactiveClassName:
                        "text-sidebar-foreground/70 hover:bg-sidebar-accent/70 hover:text-sidebar-foreground",
                    }}
                  >
                    <span>{tournament.name}</span>
                  </a>
                {/snippet}
              </Sidebar.MenuSubButton>
            </Sidebar.MenuSubItem>
          {/each}
        {:else}
          <Sidebar.MenuSubItem>
            <span class="px-2 py-1.5 text-xs text-muted-foreground">
              No tournaments in elimination.
            </span>
          </Sidebar.MenuSubItem>
        {/if}
      </Sidebar.MenuSub>
    </Collapsible.Content>
  </Sidebar.MenuItem>
</Collapsible.Root>
