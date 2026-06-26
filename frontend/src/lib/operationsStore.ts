import { writable, derived, get } from "svelte/store";
import { Events } from "@wailsio/runtime";
import { OperationsService } from "../../bindings/github.com/k3desktop/k3desktop/service";
import type { OperationEventDTO } from "../../bindings/github.com/k3desktop/k3desktop/dto";

export type OperationPhase = "start" | "done" | "error";

export interface OperationState {
  id: string;
  kind: string;
  target: string;
  phase: OperationPhase;
  message?: string;
  error?: string;
  startedAt: string;
}

// Key shape: `${kind}:${target}` — collision-free; one op per kind per target matches the
// existing `busy[name]` invariant in the routes.
export const operations = writable<Map<string, OperationState>>(new Map());

const DONE_FADE_MS = 4_000;
const fadeTimers = new Map<string, ReturnType<typeof setTimeout>>();

function keyOf(kind: string, target: string) {
  return `${kind}:${target}`;
}

function apply(ev: OperationEventDTO) {
  const k = keyOf(ev.kind, ev.target);
  const state: OperationState = {
    id: ev.id,
    kind: ev.kind,
    target: ev.target,
    phase: ev.phase as OperationPhase,
    message: ev.message || undefined,
    error: ev.error || undefined,
    startedAt: ev.startedAt,
  };

  operations.update((m) => {
    const next = new Map(m);
    next.set(k, state);
    return next;
  });

  const existing = fadeTimers.get(k);
  if (existing) clearTimeout(existing);

  if (state.phase === "done") {
    const timer = setTimeout(() => {
      operations.update((m) => {
        // only delete if the entry hasn't been replaced by a newer op with the same key
        const current = m.get(k);
        if (current?.id !== state.id) return m;
        const next = new Map(m);
        next.delete(k);
        return next;
      });
      fadeTimers.delete(k);
    }, DONE_FADE_MS);
    fadeTimers.set(k, timer);
  }
  // Errors are sticky until dismiss(key) is called.
}

export function dismiss(kind: string, target: string) {
  const k = keyOf(kind, target);
  const timer = fadeTimers.get(k);
  if (timer) {
    clearTimeout(timer);
    fadeTimers.delete(k);
  }
  operations.update((m) => {
    if (!m.has(k)) return m;
    const next = new Map(m);
    next.delete(k);
    return next;
  });
}

// Reactive helpers — each returns a derived store the route can subscribe to.
function opsForTarget(kindPrefix: string, target: string) {
  return derived(operations, ($ops) => {
    for (const [k, v] of $ops) {
      if (v.target === target && (k.startsWith(`${kindPrefix}.`) || v.kind === kindPrefix)) {
        return v;
      }
    }
    return null;
  });
}

export const clusterBusy = (name: string) => opsForTarget("cluster", name);
export const nodeBusy = (name: string) => opsForTarget("node", name);
export const registryBusy = (name: string) => opsForTarget("registry", name);

// Imperative read for one-off checks (e.g. button-click guards).
export function isBusy(kindPrefix: string, target: string): boolean {
  const ops = get(operations);
  for (const v of ops.values()) {
    if (v.target === target && v.phase === "start" && (v.kind.startsWith(`${kindPrefix}.`) || v.kind === kindPrefix)) {
      return true;
    }
  }
  return false;
}

let bootstrapped = false;

export async function bootstrapOperations() {
  if (bootstrapped) return;
  bootstrapped = true;
  try {
    const active = await OperationsService.ListActive();
    if (active?.length) {
      operations.update((m) => {
        const next = new Map(m);
        for (const op of active) {
          next.set(keyOf(op.kind, op.target), {
            id: op.id,
            kind: op.kind,
            target: op.target,
            phase: "start",
            startedAt: op.startedAt,
          });
        }
        return next;
      });
    }
  } catch { /* non-fatal — events will populate */ }

  const handler = (ev: any) => {
    const payload: OperationEventDTO = ev?.data ?? ev;
    // Wails wraps event payloads in { data: ... } for some transports; some in array form.
    const entry = Array.isArray(payload) ? payload[0] : payload;
    if (entry?.id && entry?.kind) apply(entry);
  };
  Events.On("op:start", handler);
  Events.On("op:done", handler);
  Events.On("op:error", handler);
}
