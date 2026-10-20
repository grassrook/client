<script lang="ts">
  import type { Snippet } from "svelte";
  import { link } from "svelte-spa-router";
  import active from "svelte-spa-router/active";

  import { cn } from "$src/utils";
  import Separator from "$src/components/ui/separator/separator.svelte";
  import ThemeToggle from "./theme-toggle.svelte";

  import * as Sidebar from "$src/components/ui/sidebar";

  import Trophy from "@lucide/svelte/icons/trophy";
  import UsersRound from "@lucide/svelte/icons/users-round";
  import MapPinned from "@lucide/svelte/icons/map-pinned";
  import House from "@lucide/svelte/icons/house";
  import SlidersHorizontal from "@lucide/svelte/icons/sliders-horizontal";
  import TournamentsMenu from "./sidebar/tournaments-menu.svelte";

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
    <Sidebar.Root>
      <Sidebar.Header>
        <div class="flex h-16 items-center gap-3 px-4">
          <div
            class="flex size-9 shrink-0 items-center justify-center rounded-md bg-sidebar-primary text-sidebar-primary-foreground"
          >
            <Trophy class="size-4" aria-hidden="true" />
          </div>

          <div class="min-w-0">
            <p class="truncate text-sm font-semibold text-sidebar-foreground">
              Judging Console
            </p>
            <p class="truncate text-xs text-sidebar-foreground/60">
              Tournament operations
            </p>
          </div>
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
                      <span>Home</span>
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
                      <span>Drivers</span>
                      <span class="ml-auto text-[10px] uppercase">Soon</span>
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

        <div class="flex items-center justify-between gap-3 p-3">
          <div class="min-w-0">
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
        class="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border bg-background/95 px-4 backdrop-blur supports-[backdrop-filter]:bg-background/80 lg:px-8"
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
