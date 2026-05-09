<script lang="ts">
  import { onMount } from "svelte";
  import { ClusterService, KubeconfigService } from "../../bindings/github.com/k3desktop/k3desktop/service";
  import { ClusterDTO, KubeconfigContextDTO } from "../../bindings/github.com/k3desktop/k3desktop/dto";
  import ErrorAlert from "../lib/ErrorAlert.svelte";

  let contexts: KubeconfigContextDTO[] = $state([]);
  let clusters: ClusterDTO[] = $state([]);
  let selectedCluster = $state("");
  let yaml = $state("");
  let loadingContexts = $state(false);
  let loadingYaml = $state(false);
  let exporting = $state(false);
  let exported = $state("");
  let error = $state("");
  let busyContext = $state("");

  async function loadContexts() {
    loadingContexts = true;
    error = "";
    try {
      [contexts, clusters] = await Promise.all([
        KubeconfigService.ListContexts(),
        ClusterService.ListClusters(),
      ]);
      contexts = contexts.sort((a, b) => {
        if (a.current) return -1;
        if (b.current) return 1;
        return a.name.localeCompare(b.name);
      });
      if (clusters.length > 0 && !selectedCluster) selectedCluster = clusters[0].name;
    } catch (e: any) {
      error = String(e);
    } finally {
      loadingContexts = false;
    }
  }

  async function setActive(name: string) {
    busyContext = name;
    error = "";
    try {
      await KubeconfigService.SetCurrentContext(name);
      await loadContexts();
    } catch (e: any) {
      error = String(e);
    } finally {
      busyContext = "";
    }
  }

  async function deleteContext(name: string) {
    if (!confirm(`Delete context "${name}"?`)) return;
    busyContext = name;
    error = "";
    try {
      await KubeconfigService.DeleteContext(name);
      await loadContexts();
    } catch (e: any) {
      error = String(e);
    } finally {
      busyContext = "";
    }
  }

  async function previewYaml() {
    if (!selectedCluster) return;
    loadingYaml = true;
    yaml = "";
    error = "";
    try {
      yaml = await KubeconfigService.GetKubeconfigYAML(selectedCluster);
    } catch (e: any) {
      error = String(e);
    } finally {
      loadingYaml = false;
    }
  }

  async function exportKubeconfig() {
    if (!selectedCluster) return;
    exporting = true;
    exported = "";
    error = "";
    try {
      const path = await KubeconfigService.ExportKubeconfig(selectedCluster);
      exported = path;
      await loadContexts();
    } catch (e: any) {
      error = String(e);
    } finally {
      exporting = false;
    }
  }

  onMount(loadContexts);
</script>

<div class="p-6">
  <div class="flex items-center justify-between mb-6">
    <h4 class="text-2xl font-semibold text-gray-900 dark:text-gray-100">Kubeconfig</h4>
    <button
      onclick={loadContexts}
      disabled={loadingContexts}
      class="px-3 py-1.5 rounded-lg border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 text-xs hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 transition-colors"
    >
      {loadingContexts ? "Refreshing…" : "Refresh"}
    </button>
  </div>

  <ErrorAlert bind:message={error} />

  {#if exported}
    <div class="mb-4 p-3 rounded-lg bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300 text-sm">
      Merged to <span class="font-mono">{exported}</span>
    </div>
  {/if}

  <div class="grid grid-cols-1 xl:grid-cols-2 gap-6">

    <!-- Right: merge from cluster -->
    <div>
      <h5 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 uppercase tracking-wide">Merge from cluster</h5>
      <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4">
        {#if clusters.length === 0}
          <p class="text-sm text-gray-400 dark:text-gray-500">No clusters available.</p>
        {:else}
          <div class="flex items-center gap-2 mb-4">
            <select
              bind:value={selectedCluster}
              class="flex-1 rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
            >
              {#each clusters as c}
                <option value={c.name}>{c.name}</option>
              {/each}
            </select>
          </div>
          <div class="flex gap-2 mb-4">
            <button
              onclick={previewYaml}
              disabled={!selectedCluster || loadingYaml}
              class="px-3 py-1.5 rounded-lg border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 text-sm hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 transition-colors"
            >
              {loadingYaml ? "Loading…" : "Preview YAML"}
            </button>
            <button
              onclick={exportKubeconfig}
              disabled={!selectedCluster || exporting}
              class="px-3 py-1.5 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim disabled:opacity-50 transition-colors"
            >
              {exporting ? "Merging…" : "Merge to ~/.kube/config"}
            </button>
          </div>
          {#if loadingYaml}
            <div class="text-sm text-gray-400 dark:text-gray-500">Loading…</div>
          {:else if yaml}
            <pre class="rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 p-3 text-xs font-mono text-gray-700 dark:text-gray-300 overflow-auto max-h-64">{yaml}</pre>
          {/if}
        {/if}
      </div>
    </div>

    <!-- Left: context manager -->
    <div>
      <h5 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 uppercase tracking-wide">Contexts</h5>
      {#if loadingContexts}
        <div class="text-sm text-gray-400 dark:text-gray-500">Loading…</div>
      {:else if contexts.length === 0}
        <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 text-sm text-gray-400 dark:text-gray-500">
          No contexts in ~/.kube/config. Merge a cluster kubeconfig to get started.
        </div>
      {:else}
        <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 divide-y divide-gray-100 dark:divide-gray-700 overflow-hidden">
          {#each contexts as ctx (ctx.name)}
            {@const isBusy = busyContext === ctx.name}
            <div class="flex items-center gap-3 px-4 py-3 {ctx.current ? 'bg-brand/5 dark:bg-brand/10' : ''}">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span class="font-medium text-sm text-gray-900 dark:text-gray-100 truncate">{ctx.name}</span>
                  {#if ctx.current}
                    <span class="px-1.5 py-0.5 rounded-full text-xs font-semibold bg-brand text-gray-900">active</span>
                  {/if}
                </div>
                <div class="mt-0.5 text-xs text-gray-500 dark:text-gray-400 truncate">
                  cluster: {ctx.cluster} · user: {ctx.user}
                </div>
              </div>
              <div class="flex items-center gap-1.5 shrink-0">
                {#if !ctx.current}
                  <button
                    onclick={() => setActive(ctx.name)}
                    disabled={isBusy || !!busyContext}
                    class="px-2.5 py-1 rounded-md text-xs font-medium border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 transition-colors"
                  >
                    {isBusy ? "…" : "Set active"}
                  </button>
                {/if}
                <button
                  onclick={() => deleteContext(ctx.name)}
                  disabled={isBusy || !!busyContext}
                  class="px-2.5 py-1 rounded-md text-xs font-medium border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-50 transition-colors"
                >
                  {isBusy ? "…" : "Delete"}
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>


  </div>
</div>
