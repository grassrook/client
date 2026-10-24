<script lang="ts">
  import House from "@lucide/svelte/icons/house";
  import MapPinned from "@lucide/svelte/icons/map-pinned";
  import SlidersHorizontal from "@lucide/svelte/icons/sliders-horizontal";
  import Trophy from "@lucide/svelte/icons/trophy";
  import UsersRound from "@lucide/svelte/icons/users-round";
  import type { Snippet } from "svelte";
  import { link } from "svelte-spa-router";
  import active from "svelte-spa-router/active";

  import Separator from "$src/components/ui/separator/separator.svelte";
  import * as Sidebar from "$src/components/ui/sidebar";
  import { cn } from "$src/utils";

  import TournamentsMenu from "./sidebar/tournaments-menu.svelte";
  import ThemeToggle from "./theme-toggle.svelte";

  let {
    fixed = false,
    fluid = false,
    class: className = "",
    children,
  }: {
    fixed?: boolean;
    fluid?: boolean;
    class?: string;
    children?: Snippet;
  } = $props();
</script>

<div class="min-h-svh bg-muted/20 text-foreground">
  <Sidebar.Provider>
    <Sidebar.Root collapsible="icon" class="bg-sidebar">
      <Sidebar.Header>
        <div
          class="flex h-16 items-center gap-3 px-4 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0"
        >
          <div
            class="flex size-9 shrink-0 items-center justify-center rounded-md bg-sidebar-primary text-sidebar-primary-foreground group-data-[collapsible=icon]:hidden"
          >
            <Trophy class="size-4" aria-hidden="true" />
          </div>

          <div class="min-w-0 flex-1 group-data-[collapsible=icon]:hidden">
            <p class="truncate text-sm font-semibold text-sidebar-foreground">
              Judging Console
            </p>
            <p class="truncate text-xs text-sidebar-foreground/60">
              Tournament operations
            </p>
          </div>

          <Sidebar.Trigger />
        </div>
      </Sidebar.Header>

      <Sidebar.Content>
        <Sidebar.Group>
          <Sidebar.GroupLabel>Workspace</Sidebar.GroupLabel>

          <Sidebar.GroupContent>
            <Sidebar.Menu>
              <Sidebar.MenuItem>
                <Sidebar.MenuButton>
                  {#snippet child({ props })}
                    <a
                      {...props}
                      href="/"
                      use:link
                      use:active={{
                        path: "/",
                        className:
                          "bg-sidebar-accent text-sidebar-accent-foreground",
                        inactiveClassName:
                          "hover:bg-sidebar-accent/70 hover:text-sidebar-foreground",
                      }}
                    >
                      <House aria-hidden="true" />
                      <span class="group-data-[collapsible=icon]:hidden"
                        >Home</span
                      >
                    </a>
                  {/snippet}
                </Sidebar.MenuButton>
              </Sidebar.MenuItem>
              <TournamentsMenu />

              <Sidebar.MenuItem>
                <Sidebar.MenuButton class="cursor-not-allowed opacity-45">
                  {#snippet child({ props })}
                    <div {...props} aria-disabled="true">
                      <UsersRound aria-hidden="true" />
                      <span class="group-data-[collapsible=icon]:hidden"
                        >Drivers</span
                      >
                      <span
                        class="ml-auto text-[10px] uppercase group-data-[collapsible=icon]:hidden"
                      >
                        Soon
                      </span>
                    </div>
                  {/snippet}
                </Sidebar.MenuButton>
              </Sidebar.MenuItem>

              <Sidebar.MenuItem>
                <Sidebar.MenuButton class="cursor-not-allowed opacity-45">
                  {#snippet child({ props })}
                    <div {...props} aria-disabled="true">
                      <MapPinned aria-hidden="true" />
                      <span>Tracks</span>
                      <span class="ml-auto text-[10px] uppercase">Soon</span>
                    </div>
                  {/snippet}
                </Sidebar.MenuButton>
              </Sidebar.MenuItem>
              <Sidebar.MenuItem>
                <Sidebar.MenuButton>
                  {#snippet child({ props })}
                    <a
                      {...props}
                      href="/criteria-configs"
                      use:link
                      use:active={{
                        path: "/criteria-configs",
                        className: "bg-sidebar-accent",
                        inactiveClassName:
                          "hover:bg-sidebar-accent/70 hover:text-sidebar-foreground",
                      }}
                    >
                      <SlidersHorizontal aria-hidden="true" />
                      <span>Criteria Configs</span>
                    </a>
                  {/snippet}
                </Sidebar.MenuButton>
              </Sidebar.MenuItem>
            </Sidebar.Menu>
          </Sidebar.GroupContent>
        </Sidebar.Group>
      </Sidebar.Content>

      <Sidebar.Footer class="px-0">
        <Separator class="bg-sidebar-border" />

        <div
          class="flex items-center justify-between gap-3 p-3 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0"
        >
          <div class="min-w-0 group-data-[collapsible=icon]:hidden">
            <p class="text-xs font-medium text-sidebar-foreground">
              Appearance
            </p>
            <p class="truncate text-xs text-sidebar-foreground/60">
              Switch between light and dark
            </p>
          </div>

          <ThemeToggle />
        </div>
      </Sidebar.Footer>
    </Sidebar.Root>

    <Sidebar.Inset>
      <header
        class="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border bg-background/50 px-4 backdrop-blur supports-[backdrop-filter]:bg-background/80 lg:px-8 lg:hidden"
      >
        <div class="flex items-center gap-3">
          <Sidebar.Trigger />

          <div>
            <p class="text-sm font-semibold lg:hidden">Judging Console</p>
          </div>
        </div>
        <div class="lg:hidden">
          <ThemeToggle />
        </div>
      </header>

      <main
        data-layout={fixed ? "fixed" : ""}
        class={cn(
          "flex-1",
          fixed && "flex grow flex-col overflow-hidden",
          !fluid && "w-full max-w-none",
          className,
        )}
      >
        {@render children?.()}
      </main>
    </Sidebar.Inset>
  </Sidebar.Provider>
</div>
