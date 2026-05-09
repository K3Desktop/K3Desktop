<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { EditorState } from "@codemirror/state";
  import { EditorView, lineNumbers, highlightActiveLine, keymap } from "@codemirror/view";
  import { linter, lintGutter, type Diagnostic } from "@codemirror/lint";
  import { yaml } from "@codemirror/lang-yaml";
  import { parse } from "yaml";
  import { oneDark } from "@codemirror/theme-one-dark";
  import { history, defaultKeymap, historyKeymap, indentWithTab } from "@codemirror/commands";
  import { indentOnInput, bracketMatching, foldGutter } from "@codemirror/language";
  import { closeBrackets, closeBracketsKeymap } from "@codemirror/autocomplete";
  import { dark } from "./theme";

  let {
    value = $bindable(""),
    placeholder = "key: value\n",
    minHeight = "120px",
  }: { value: string; placeholder?: string; minHeight?: string } = $props();

  let container: HTMLDivElement | undefined = $state();
  let view: EditorView | undefined;
  let internalUpdate = false;

  function buildExtensions(isDark: boolean) {
    return [
      history(),
      indentOnInput(),
      bracketMatching(),
      closeBrackets(),
      foldGutter(),
      keymap.of([indentWithTab, ...closeBracketsKeymap, ...defaultKeymap, ...historyKeymap]),
      lineNumbers(),
      highlightActiveLine(),
      lintGutter(),
      linter((view) => {
        const diagnostics: Diagnostic[] = [];
        try {
          parse(view.state.doc.toString());
        } catch (e: any) {
          const from = e.pos?.[0] || 0;
          const to = e.pos?.[1] || view.state.doc.length;
          diagnostics.push({
            from,
            to,
            severity: "error",
            message: e.message,
          });
        }
        return diagnostics;
      }),
      yaml(),
      EditorView.lineWrapping,
      ...(isDark ? [oneDark] : []),
      EditorView.theme({
        "&": { minHeight, fontSize: "12px" },
        ".cm-scroller": { fontFamily: "ui-monospace, monospace" },
      }),
      EditorView.updateListener.of((update) => {
        if (update.docChanged && !internalUpdate) {
          value = update.state.doc.toString();
        }
      }),
    ];
  }

  onMount(() => {
    view = new EditorView({
      state: EditorState.create({
        doc: value || "",
        extensions: buildExtensions($dark),
      }),
      parent: container!,
    });
  });

  onDestroy(() => view?.destroy());

  // Sync value changes from parent → editor (only if different)
  $effect(() => {
    if (!view) return;
    const current = view.state.doc.toString();
    if (current !== value) {
      internalUpdate = true;
      view.dispatch({
        changes: { from: 0, to: current.length, insert: value || "" },
      });
      internalUpdate = false;
    }
  });

  // Recreate extensions on dark mode toggle
  $effect(() => {
    const isDark = $dark;
    if (!view) return;
    view.dispatch({
      effects: [],
    });
    // Rebuild state preserving doc
    const doc = view.state.doc.toString();
    view.setState(
      EditorState.create({
        doc,
        extensions: buildExtensions(isDark),
      })
    );
  });
</script>

<div
  bind:this={container}
  class="rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden text-xs text-left"
  style="min-height: {minHeight}"
></div>
