<script lang="ts">
  import "./app.css";
  import { onMount } from "svelte";
  import Clusters from "./routes/Clusters.svelte";
  import Nodes from "./routes/Nodes.svelte";
  import Registries from "./routes/Registries.svelte";
  import Kubeconfig from "./routes/Kubeconfig.svelte";
  import Profiles from "./routes/Profiles.svelte";
  import Blueprints from "./routes/Blueprints.svelte";
  import Sidebar from "./lib/Sidebar.svelte";
  import LogDrawer from "./lib/LogDrawer.svelte";
  import { bootstrapLogs } from "./lib/logStore";
  import { sidebarWidth, sidebarCollapsed } from "./lib/theme";
  import { LogService } from "../bindings/github.com/k3desktop/k3desktop/service";

  type Route = "clusters" | "nodes" | "registries" | "kubeconfig" | "profiles" | "blueprints";
  let route: Route = $state("clusters");

  const logLevels = ["DEBUG", "INFO", "WARN", "ERROR"] as const;
  let currentLevel = $state("INFO");
  let logDrawerOpen = $state(false);

  async function setLogLevel(level: string) {
    try {
      await LogService.SetLevel(level);
      currentLevel = level;
    } catch { /* non-fatal */ }
  }

  onMount(async () => {
    try { currentLevel = await LogService.GetLevel(); } catch { /* non-fatal */ }
    bootstrapLogs();
  });
</script>

<div class="flex h-full bg-gray-50 dark:bg-gray-950 text-gray-900 dark:text-gray-100 overflow-hidden">
  <Sidebar
    bind:route
    bind:logDrawerOpen
    bind:currentLevel
    width={$sidebarWidth}
    collapsed={$sidebarCollapsed}
    onWidthChange={(v) => sidebarWidth.set(v)}
    onCollapsedChange={(v) => sidebarCollapsed.set(v)}
    {logLevels}
    onSetLogLevel={setLogLevel}
  />

  <main class="flex-1 overflow-y-auto min-w-0" style={logDrawerOpen ? "padding-bottom: 280px" : ""}>
    {#if route === "clusters"}
      <Clusters />
    {:else if route === "nodes"}
      <Nodes />
    {:else if route === "registries"}
      <Registries />
    {:else if route === "kubeconfig"}
      <Kubeconfig />
    {:else if route === "profiles"}
      <Profiles onNavigateToClusters={() => (route = "clusters")} />
    {:else if route === "blueprints"}
      <Blueprints />
    {/if}
  </main>
</div>

<LogDrawer bind:open={logDrawerOpen} sidebarWidth={$sidebarWidth} />
