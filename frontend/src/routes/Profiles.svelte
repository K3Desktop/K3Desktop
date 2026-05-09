<script lang="ts">
  import { onMount } from "svelte";
  import { ProfileService, VersionService } from "../../bindings/github.com/k3desktop/k3desktop/service";
  import type { ProfileDTO } from "../../bindings/github.com/k3desktop/k3desktop/dto";
  import ClusterForm from "../lib/ClusterForm.svelte";
  import { clusterFormPrefill, defaultAdv, dtoToAdv, advToDto } from "../lib/prefill";
  import ErrorAlert from "../lib/ErrorAlert.svelte";
  import type { AdvState } from "../lib/prefill";

  let { onNavigateToClusters }: { onNavigateToClusters: () => void } = $props();

  let profiles: ProfileDTO[] = $state([]);
  let loading = $state(false);
  let error = $state("");
  let saving = $state(false);

  // Edit/New form state
  let showForm = $state(false);
  let editingProfileName = $state(""); // stem name of profile being edited (empty = new)
  let profileName = $state(""); // the name/label for this profile file
  let adv: AdvState = $state(defaultAdv());

  let k3sVersions: string[] = $state([]);
  let versionsLoading = $state(false);

  async function loadVersions() {
    if (k3sVersions.length > 0) return;
    versionsLoading = true;
    try {
      k3sVersions = await VersionService.ListK3sVersions(20);
    } catch { /* non-fatal */ } finally {
      versionsLoading = false;
    }
  }

  async function load() {
    loading = true;
    error = "";
    try {
      profiles = await ProfileService.ListProfiles() ?? [];
    } catch (e: any) {
      error = String(e);
    } finally {
      loading = false;
    }
  }

  function openNew() {
    editingProfileName = "";
    profileName = "";
    adv = defaultAdv();
    loadVersions();
    showForm = true;
  }

  async function openEdit(p: ProfileDTO) {
    error = "";
    try {
      const yaml = await ProfileService.GetProfile(p.name);
      const req = await ProfileService.YAMLToAdvancedRequest(yaml);
      editingProfileName = p.name;
      profileName = p.name;
      adv = dtoToAdv(req);
      loadVersions();
      showForm = true;
    } catch (e: any) {
      error = String(e);
    }
  }

  async function saveProfile() {
    if (!profileName.trim()) { error = "Profile name required."; return; }
    saving = true;
    error = "";
    try {
      const req = advToDto(adv);
      const yaml = await ProfileService.AdvancedRequestToYAML(req);
      await ProfileService.SaveProfile(profileName.trim(), yaml);
      showForm = false;
      await load();
    } catch (e: any) {
      error = String(e);
    } finally {
      saving = false;
    }
  }

  async function deleteProfile(p: ProfileDTO) {
    if (!confirm(`Delete profile "${p.name}"?`)) return;
    error = "";
    try {
      await ProfileService.DeleteProfile(p.name);
      await load();
    } catch (e: any) {
      error = String(e);
    }
  }

  async function createClusterFromProfile(p: ProfileDTO) {
    error = "";
    try {
      const yaml = await ProfileService.GetProfile(p.name);
      const req = await ProfileService.YAMLToAdvancedRequest(yaml);
      const state = dtoToAdv(req);
      clusterFormPrefill.set(state);
      onNavigateToClusters();
    } catch (e: any) {
      error = String(e);
    }
  }

  async function importProfile() {
    error = "";
    try {
      const name = await ProfileService.ImportProfile();
      if (name) await load();
    } catch (e: any) {
      error = String(e);
    }
  }

  onMount(load);
</script>

<div class="p-6">
  <div class="flex items-center justify-between mb-6">
    <h4 class="text-2xl font-semibold text-gray-900 dark:text-gray-100">Profiles</h4>
    <div class="flex gap-2">
      <button
        onclick={importProfile}
        class="px-4 py-2 rounded-lg border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
      >
        Import
      </button>
      <button
        onclick={openNew}
        class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim transition-colors"
      >
        + New profile
      </button>
    </div>
  </div>

  <ErrorAlert bind:message={error} />

  {#if loading}
    <div class="text-sm text-gray-400 dark:text-gray-500">Loading…</div>
  {:else if profiles.length === 0}
    <div class="text-sm text-gray-400 dark:text-gray-500">
      No profiles yet. Create a profile to save a reusable cluster configuration.
    </div>
  {:else}
    <div class="grid gap-4">
      {#each profiles as p}
        <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-5 flex items-center gap-4">
          <div class="flex-1 min-w-0">
            <span class="font-medium text-gray-900 dark:text-gray-100">{p.name}</span>
            <span class="ml-2 text-xs text-gray-400 dark:text-gray-500">{p.fileName}</span>
          </div>
          <div class="flex items-center gap-2">
            <button
              onclick={() => createClusterFromProfile(p)}
              class="px-3 py-1.5 rounded-lg text-xs font-medium bg-brand text-gray-900 font-semibold hover:bg-brand-dim transition-colors"
            >
              Create cluster
            </button>
            <button
              onclick={() => openEdit(p)}
              class="px-3 py-1.5 rounded-lg text-xs font-medium border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
            >
              Edit
            </button>
            <button
              onclick={() => deleteProfile(p)}
              class="px-3 py-1.5 rounded-lg text-xs font-medium border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
            >
              Delete
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Profile form overlay -->
{#if showForm}
  <div class="fixed inset-0 bg-gray-50 dark:bg-gray-950 z-50 overflow-y-auto">
    <div class="max-w-2xl mx-auto px-6 py-8">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100">
          {editingProfileName ? `Edit profile — ${editingProfileName}` : "New profile"}
        </h2>
        <button onclick={() => (showForm = false)} class="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">✕ Cancel</button>
      </div>

      <!-- Profile name (file stem) -->
      <section class="mb-6 p-4 rounded-xl border border-accent/30 bg-accent/5">
        <label class="block">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Profile name *</span>
          <input
            bind:value={profileName}
            class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
            placeholder="my-dev-cluster"
          />
          <p class="mt-1 text-xs text-gray-400">Saved as <code>~/.config/k3desktop/profiles/{profileName || "…"}.yaml</code></p>
        </label>
      </section>

      <ClusterForm bind:adv {k3sVersions} {versionsLoading} datalistId="k3s-versions-profile" />

      <div class="flex justify-end gap-3 border-t border-gray-200 dark:border-gray-700 pt-6">
        <button onclick={() => (showForm = false)} class="px-4 py-2 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">Cancel</button>
        <button onclick={saveProfile} disabled={saving || !profileName.trim()} class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim disabled:opacity-50 transition-colors">
          {saving ? "Saving…" : "Save profile"}
        </button>
      </div>
    </div>
  </div>
{/if}
