<script lang="ts">
  import { dark } from "./theme";

  type Route = "clusters" | "nodes" | "registries" | "kubeconfig" | "profiles" | "blueprints";

  let {
    route = $bindable<Route>("clusters"),
    logDrawerOpen = $bindable(false),
    currentLevel = $bindable("INFO"),
    logLevels,
    onSetLogLevel,
    width = 224,
    collapsed = false,
    onWidthChange,
    onCollapsedChange,
  }: {
    route: Route;
    logDrawerOpen: boolean;
    currentLevel: string;
    logLevels: readonly string[];
    onSetLogLevel: (level: string) => void;
    width: number;
    collapsed: boolean;
    onWidthChange: (v: number) => void;
    onCollapsedChange: (v: boolean) => void;
  } = $props();

  const MIN_WIDTH = 48;  // icon-only collapsed
  const MAX_WIDTH = 400;
  const COLLAPSE_THRESHOLD = 80; // snap to icon-only below this

  let dragging = $state(false);
  let dragStartX = 0;
  let dragStartWidth = 0;

  const nav: { id: Route; label: string; icon: string }[] = [
    { id: "clusters",   label: "Clusters",   icon: "⬡" },
    { id: "nodes",      label: "Nodes",      icon: "◈" },
    { id: "registries", label: "Registries", icon: "⊡" },
    { id: "kubeconfig", label: "Kubeconfig", icon: "⚿" },
    { id: "profiles",   label: "Profiles",   icon: "⊞" },
    { id: "blueprints", label: "Blueprints", icon: "⧉" },
  ];

  function toggleCollapse() {
    const next = !collapsed;
    onCollapsedChange(next);
    onWidthChange(next ? MIN_WIDTH : 224);
  }

  function onDragStart(e: MouseEvent) {
    dragging = true;
    dragStartX = e.clientX;
    dragStartWidth = width;
    window.addEventListener("mousemove", onDragMove);
    window.addEventListener("mouseup", onDragEnd);
    e.preventDefault();
  }

  function onDragMove(e: MouseEvent) {
    const delta = e.clientX - dragStartX;
    const next = Math.max(MIN_WIDTH, Math.min(MAX_WIDTH, dragStartWidth + delta));
    onWidthChange(next);
    onCollapsedChange(next <= COLLAPSE_THRESHOLD);
  }

  function onDragEnd() {
    dragging = false;
    // snap: if near min, fully collapse
    if (width <= COLLAPSE_THRESHOLD) {
      onWidthChange(MIN_WIDTH);
      onCollapsedChange(true);
    }
    window.removeEventListener("mousemove", onDragMove);
    window.removeEventListener("mouseup", onDragEnd);
  }
</script>

<aside
  class="relative flex flex-col border-r border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 shrink-0"
  style="width: {width}px; transition: {dragging ? 'none' : 'width 150ms ease'};"
>
  <!-- Header -->
  {#if !collapsed}
    <div class="flex items-center justify-between px-3 py-4 border-b border-gray-100 dark:border-gray-800 overflow-hidden">
      <span class="text-lg font-black tracking-tight leading-none shrink-0">
        <span style="color:#0DCEFF">k3</span><span style="color:#CDF700">d</span><span class="text-gray-400 dark:text-gray-500 font-semibold">esktop</span>
      </span>
      <button
        onclick={toggleCollapse}
        title="Collapse sidebar"
        class="shrink-0 ml-2 p-1 rounded text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-xs leading-none"
      >◀</button>
    </div>
  {:else}
    <div class="flex flex-col items-center gap-2 py-3 border-b border-gray-100 dark:border-gray-800">
      <!-- lime "k" badge -->
      <div class="w-7 h-7 rounded-md flex items-center justify-center" style="background:#CDF700">
        <span class="text-sm font-black leading-none" style="color:#1a1a1a">k</span>
      </div>
      <button
        onclick={toggleCollapse}
        title="Expand sidebar"
        class="p-1 rounded text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-xs leading-none"
      >▶</button>
    </div>
  {/if}

  <!-- Nav -->
  <nav class="flex-1 flex flex-col gap-1 p-2 overflow-hidden">
    {#each nav as item}
      <button
        onclick={() => (route = item.id)}
        title={collapsed ? item.label : ""}
        class="relative flex items-center rounded-lg text-sm transition-colors overflow-hidden
          {collapsed ? 'justify-center px-0 py-2' : 'gap-3 px-3 py-2'}
          {route === item.id
            ? 'bg-accent/10 text-accent font-semibold'
            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200'}"
        style="width: 100%"
      >
        <!-- active left-bar indicator -->
        {#if route === item.id && !collapsed}
          <span class="absolute left-0 top-1 bottom-1 w-0.5 rounded-full" style="background:#CDF700"></span>
        {/if}
        <span class="text-base leading-none shrink-0">{item.icon}</span>
        {#if !collapsed}
          <span class="truncate">{item.label}</span>
        {/if}
      </button>
    {/each}
  </nav>

  <!-- Footer controls -->
  <div class="p-2 border-t border-gray-100 dark:border-gray-800 space-y-1 overflow-hidden">
    <!-- Dark mode -->
    <button
      onclick={() => dark.update((v) => !v)}
      title={collapsed ? ($dark ? "Light mode" : "Dark mode") : ""}
      class="w-full flex items-center rounded-lg text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors overflow-hidden
        {collapsed ? 'justify-center px-0 py-2' : 'gap-2 px-3 py-2'}"
        style="width: 100%"
    >
      {#if $dark}
        <span class="shrink-0">☀</span>{#if !collapsed}<span class="truncate">Light mode</span>{/if}
      {:else}
        <span class="shrink-0">☾</span>{#if !collapsed}<span class="truncate">Dark mode</span>{/if}
      {/if}
    </button>

    <!-- Logs toggle -->
    <button
      onclick={() => (logDrawerOpen = !logDrawerOpen)}
      title={collapsed ? "Logs" : ""}
      class="w-full flex items-center rounded-lg text-sm transition-colors overflow-hidden
        {collapsed ? 'justify-center px-0 py-2' : 'gap-2 px-3 py-2'}
        {logDrawerOpen ? 'bg-accent/10 text-accent' : 'text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'}"
        style="width: 100%"
    >
      <span class="shrink-0">▤</span>{#if !collapsed}<span class="truncate">Logs</span>{/if}
    </button>

    <!-- Log level selector — hidden when collapsed -->
    {#if !collapsed}
      <div class="flex items-center gap-2 px-3 py-1">
        <span class="text-xs text-gray-400 dark:text-gray-500 shrink-0">Level</span>
        <select
          value={currentLevel}
          onchange={(e) => onSetLogLevel((e.target as HTMLSelectElement).value)}
          class="flex-1 text-xs rounded bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 px-2 py-0.5 focus:outline-none focus:ring-1 focus:ring-accent"
        >
          {#each logLevels as l}
            <option value={l} selected={l === currentLevel}>{l}</option>
          {/each}
        </select>
      </div>
    {/if}
  </div>

  <!-- Resize handle -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    onmousedown={onDragStart}
    class="absolute top-0 right-0 h-full w-1 cursor-col-resize hover:bg-accent/40 transition-colors {dragging ? 'bg-accent/60' : ''}"
    style="z-index: 10"
  ></div>
</aside>
