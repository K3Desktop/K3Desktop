<script lang="ts">
  import { onMount } from "svelte";
  import { ClusterService, NodeService, VersionService } from "../../bindings/github.com/k3desktop/k3desktop/service";
  import { ClusterDTO, NodeDTO } from "../../bindings/github.com/k3desktop/k3desktop/dto";
  import OpLog from "../lib/OpLog.svelte";
  import { clearOpLog } from "../lib/logStore";
  import ErrorAlert from "../lib/ErrorAlert.svelte";

  let clusters: ClusterDTO[] = $state([]);
  let selected = $state("");
  let nodes: NodeDTO[] = $state([]);
  let loading = $state(false);
  let addingAgent = $state(false);
  let busy: Record<string, boolean> = $state({});
  let visibleLogs: Record<string, boolean> = $state({}); // node names with log panel open
  let error = $state("");

  // Upgrade modal state
  let upgradeModalNode: NodeDTO | null = $state(null);
  let upgradeImage = $state("");
  let k3sVersions: string[] = $state([]);
  let loadingVersions = $state(false);
  let upgrading = $state(false);
  let versionDropdownOpen = $state(false);

  async function loadClusters() {
    try {
      clusters = await ClusterService.ListClusters();
      if (clusters.length > 0 && !selected) {
        selected = clusters[0].name;
        await loadNodes();
      }
    } catch (e: any) {
      error = String(e);
    }
  }

  async function loadNodes() {
    if (!selected) return;
    loading = true;
    error = "";
    try {
      nodes = await NodeService.ListNodes(selected);
    } catch (e: any) {
      error = String(e);
    } finally {
      loading = false;
    }
  }

  async function addAgent() {
    if (!selected) return;
    addingAgent = true;
    visibleLogs["__adding_agent__"] = true;
    clearOpLog(selected);
    try {
      await NodeService.AddAgent(selected);
      await loadNodes();
    } catch (e: any) {
      error = String(e);
    } finally {
      addingAgent = false;
    }
  }

  async function del(name: string) {
    if (!confirm(`Delete node "${name}"?`)) return;
    busy[name] = true;
    visibleLogs[name] = true;
    clearOpLog(name);
    try {
      await NodeService.DeleteNode(name);
      await loadNodes();
    } catch (e: any) {
      error = String(e);
    } finally {
      delete busy[name];
    }
  }

  async function toggleNode(n: NodeDTO) {
    busy[n.name] = true;
    visibleLogs[n.name] = true;
    clearOpLog(n.name);
    try {
      if (n.state === "running") {
        await NodeService.StopNode(n.name);
      } else {
        await NodeService.StartNode(n.name);
      }
      await loadNodes();
    } catch (e: any) {
      error = String(e);
    } finally {
      delete busy[n.name];
    }
  }

  async function restartNode(n: NodeDTO) {
    busy[n.name] = true;
    visibleLogs[n.name] = true;
    clearOpLog(n.name);
    try {
      await NodeService.RestartNode(n.name);
      await loadNodes();
    } catch (e: any) {
      error = String(e);
    } finally {
      delete busy[n.name];
    }
  }

  async function openUpgradeModal(n: NodeDTO) {
    upgradeModalNode = n;
    upgradeImage = "";
    k3sVersions = [];
    loadingVersions = true;
    try {
      k3sVersions = await VersionService.ListK3sVersions(20);
      if (k3sVersions.length > 0) upgradeImage = k3sVersions[0];
    } catch (e: any) {
      error = String(e);
      upgradeModalNode = null;
    } finally {
      loadingVersions = false;
    }
  }

  function closeUpgradeModal() {
    upgradeModalNode = null;
    upgradeImage = "";
    versionDropdownOpen = false;
  }

  async function confirmUpgrade() {
    if (!upgradeModalNode || !upgradeImage) return;
    const node = upgradeModalNode;
    const image = `rancher/k3s:${upgradeImage}`;
    closeUpgradeModal();
    busy[node.name] = true;
    visibleLogs[node.name] = true;
    upgrading = true;
    clearOpLog(node.name);
    try {
      await NodeService.UpgradeNode(node.name, image);
      // Reload in a fresh async tick so Docker has settled after the upgrade.
      setTimeout(() => loadNodes(), 500);
    } catch (e: any) {
      error = String(e);
    } finally {
      delete busy[node.name];
      upgrading = false;
    }
  }

  // Returns human-readable image label: tag if available, else short SHA.
  function displayImage(image: string): string {
    // "rancher/k3s:v1.32.0-k3s1" → "v1.32.0-k3s1"
    if (image.includes(':') && !image.startsWith('sha256:')) {
      const tag = image.split(':').pop()!;
      if (!/^[0-9a-f]{12,}$/.test(tag)) return tag;
    }
    // "sha256:abcdef..." or "repo@sha256:abcdef..."
    const m = image.match(/sha256:([0-9a-f]+)/);
    if (m) return `sha256:${m[1].slice(0, 12)}…`;
    return image;
  }

  onMount(loadClusters);
</script>

<div class="p-6">
  <div class="flex items-center justify-between mb-6">
    <div class="flex items-center gap-4">
      <h4 class="text-2xl font-semibold text-gray-900 dark:text-gray-100">Nodes</h4>
      {#if clusters.length > 0}
        <select
          bind:value={selected}
          onchange={loadNodes}
          class="rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
        >
          {#each clusters as c}
            <option value={c.name}>{c.name}</option>
          {/each}
        </select>
      {/if}
    </div>
    {#if selected}
      <button
        onclick={addAgent}
        disabled={addingAgent}
        class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        {addingAgent ? "Adding…" : "+ Add agent"}
      </button>
    {/if}
  </div>

  <ErrorAlert bind:message={error} />

  {#if visibleLogs["__adding_agent__"]}
    <OpLog active={addingAgent} target={selected} onclose={() => { delete visibleLogs["__adding_agent__"]; }} />
  {/if}

  {#if loading}
    <div class="text-sm text-gray-400 dark:text-gray-500">Loading…</div>
  {:else if nodes.length === 0}
    <div class="text-sm text-gray-400 dark:text-gray-500">{selected ? "No nodes." : "Select a cluster."}</div>
  {:else}
    <div class="rounded-xl border border-gray-200 dark:border-gray-700 overflow-x-auto">
      <table class="w-full min-w-max text-sm">
        <thead class="bg-gray-50 dark:bg-gray-800/50">
          <tr>
            <th class="text-left px-4 py-3 font-medium text-gray-600 dark:text-gray-400">Name</th>
            <th class="text-left px-4 py-3 font-medium text-gray-600 dark:text-gray-400">Role</th>
            <th class="text-left px-4 py-3 font-medium text-gray-600 dark:text-gray-400">State</th>
            <th class="text-left px-4 py-3 font-medium text-gray-600 dark:text-gray-400">Image</th>
            <th class="px-4 py-3"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
          {#each nodes as n (n.name)}
            {@const isBusy = !!busy[n.name]}
            <tr class="bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-750">
              <td class="px-4 py-3 font-mono text-gray-900 dark:text-gray-100">{n.name}</td>
              <td class="px-4 py-3 text-gray-600 dark:text-gray-400 capitalize">{n.role}</td>
              <td class="px-4 py-3">
                {#if isBusy}
                  <span class="px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400 animate-pulse">
                    {n.state === "running" ? "stopping…" : "starting…"}
                  </span>
                {:else}
                  <span class="px-2 py-0.5 rounded-full text-xs font-medium {n.state === 'running' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'}">
                    {n.state}
                  </span>
                {/if}
              </td>
              <td class="px-4 py-3 font-mono text-xs text-gray-500 dark:text-gray-400 max-w-48 truncate" title={n.image}>{displayImage(n.image)}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2 justify-end">
                  <button
                    onclick={() => toggleNode(n)}
                    disabled={isBusy}
                    class="px-2.5 py-1 rounded text-xs border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    {n.state === "running" ? "Stop" : "Start"}
                  </button>
                  {#if n.state === "running"}
                    <button
                      onclick={() => restartNode(n)}
                      disabled={isBusy}
                      class="px-2.5 py-1 rounded text-xs border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    >
                      Restart
                    </button>
                  {/if}
                  {#if n.role === "agent"}
                  <button
                    onclick={() => openUpgradeModal(n)}
                    disabled={isBusy}
                    class="px-2.5 py-1 rounded text-xs border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    Upgrade
                  </button>
                  {/if}
                  {#if n.role === "agent"}
                    <button
                      onclick={() => del(n.name)}
                      disabled={isBusy}
                      class="px-2.5 py-1 rounded text-xs border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    >Delete</button>
                  {/if}
                </div>
              </td>
            </tr>
            {#if visibleLogs[n.name]}
              <tr class="bg-white dark:bg-gray-800">
                <td colspan="5" class="px-4 pb-3">
                  <OpLog active={isBusy} target={n.name} onclose={() => { delete visibleLogs[n.name]; }} />
                </td>
              </tr>
            {/if}
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<!-- Upgrade modal -->
{#if upgradeModalNode}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm mx-4">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">Upgrade node</h3>
      <p class="text-sm text-gray-500 dark:text-gray-400 font-mono mb-4">{upgradeModalNode.name}</p>

      <div class="mb-2 text-xs text-gray-500 dark:text-gray-400">
        Current: <span class="font-mono" title={upgradeModalNode.image}>{displayImage(upgradeModalNode.image)}</span>
      </div>

      <label for="upgrade-version-btn" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">New version</label>
      {#if loadingVersions}
        <div class="text-sm text-gray-400 dark:text-gray-500 py-2">Loading versions…</div>
      {:else}
        <div class="relative mb-4">
          <button
            id="upgrade-version-btn"
            type="button"
            onclick={() => versionDropdownOpen = !versionDropdownOpen}
            class="w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm text-left focus:outline-none focus:ring-2 focus:ring-accent flex items-center justify-between"
          >
            <span>{upgradeImage || "Select version"}</span>
            <svg class="w-4 h-4 text-gray-500 dark:text-gray-400 flex-shrink-0" viewBox="0 0 20 20" fill="none">
              <path d="M5 8l5 5 5-5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>
          {#if versionDropdownOpen}
            <div class="absolute z-10 w-full mt-1 rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 shadow-lg overflow-y-auto max-h-48">
              {#each k3sVersions as v}
                <button
                  type="button"
                  onclick={() => { upgradeImage = v; versionDropdownOpen = false; }}
                  class="w-full text-left px-3 py-2 text-sm text-gray-900 dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-600 {upgradeImage === v ? 'bg-gray-100 dark:bg-gray-600 font-medium' : ''}"
                >
                  {v}
                </button>
              {/each}
            </div>
          {/if}
        </div>
      {/if}

      <div class="flex gap-3 justify-end">
        <button
          onclick={closeUpgradeModal}
          class="px-4 py-2 rounded-lg border border-gray-200 dark:border-gray-600 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
        >
          Cancel
        </button>
        <button
          onclick={confirmUpgrade}
          disabled={!upgradeImage || loadingVersions}
          class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Upgrade
        </button>
      </div>
    </div>
  </div>
{/if}
