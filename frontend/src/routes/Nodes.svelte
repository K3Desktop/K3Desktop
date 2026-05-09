<script lang="ts">
  import { onMount } from "svelte";
  import { ClusterService } from "../../bindings/github.com/k3desktop/k3desktop/service";
  import { NodeService } from "../../bindings/github.com/k3desktop/k3desktop/service";
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
    <div class="rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
      <table class="w-full text-sm">
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
              <td class="px-4 py-3 font-mono text-xs text-gray-500 dark:text-gray-400 max-w-48 truncate">{n.image}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2 justify-end">
                  <button
                    onclick={() => toggleNode(n)}
                    disabled={isBusy}
                    class="px-2.5 py-1 rounded text-xs border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    {n.state === "running" ? "Stop" : "Start"}
                  </button>
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
