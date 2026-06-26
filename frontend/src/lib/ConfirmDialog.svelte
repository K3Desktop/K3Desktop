<script lang="ts">
  import { tick } from "svelte";

  interface Props {
    open: boolean;
    title: string;
    message: string;
    confirmLabel?: string;
    cancelLabel?: string;
    danger?: boolean;
    onconfirm?: () => void;
    oncancel?: () => void;
  }

  let {
    open = $bindable(),
    title,
    message,
    confirmLabel = "Confirm",
    cancelLabel = "Cancel",
    danger = false,
    onconfirm,
    oncancel,
  }: Props = $props();

  let confirmBtn: HTMLButtonElement | undefined = $state();
  let cancelBtn: HTMLButtonElement | undefined = $state();
  let lastActive: HTMLElement | null = null;

  $effect(() => {
    if (open) {
      lastActive = (document.activeElement as HTMLElement | null) ?? null;
      tick().then(() => confirmBtn?.focus());
    } else if (lastActive) {
      const el = lastActive;
      lastActive = null;
      tick().then(() => el.focus?.());
    }
  });

  $effect(() => {
    if (!open) return;
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") {
        e.preventDefault();
        handleCancel();
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  });

  function handleCancel() {
    open = false;
    oncancel?.();
  }

  function handleConfirm() {
    open = false;
    onconfirm?.();
  }

  function trapTab(e: KeyboardEvent) {
    if (e.key !== "Tab") return;
    if (!e.shiftKey && document.activeElement === confirmBtn) {
      e.preventDefault();
      cancelBtn?.focus();
    } else if (e.shiftKey && document.activeElement === cancelBtn) {
      e.preventDefault();
      confirmBtn?.focus();
    }
  }
</script>

{#if open}
  <div
    class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
    onclick={handleCancel}
    role="presentation"
  >
    <div
      role="dialog"
      tabindex="-1"
      aria-modal="true"
      aria-labelledby="confirm-title"
      onclick={(e) => e.stopPropagation()}
      onkeydown={trapTab}
      class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md p-6 mx-4"
    >
      <h5
        id="confirm-title"
        class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2"
      >
        {title}
      </h5>
      <p class="text-sm text-gray-600 dark:text-gray-400 whitespace-pre-line">
        {message}
      </p>
      <div class="flex justify-end gap-3 mt-6">
        <button
          bind:this={cancelBtn}
          type="button"
          onclick={handleCancel}
          class="px-4 py-2 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors focus:outline-none focus:ring-2 focus:ring-accent"
        >
          {cancelLabel}
        </button>
        <button
          bind:this={confirmBtn}
          type="button"
          onclick={handleConfirm}
          class={danger
            ? "px-4 py-2 rounded-lg text-sm font-semibold bg-red-600 text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-400 transition-colors"
            : "px-4 py-2 rounded-lg text-sm font-semibold bg-brand text-gray-900 hover:bg-brand-dim focus:outline-none focus:ring-2 focus:ring-accent transition-colors"}
        >
          {confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}
