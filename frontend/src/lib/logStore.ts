import { writable } from "svelte/store";
import { Events } from "@wailsio/runtime";
import { LogService } from "../../bindings/github.com/k3desktop/k3desktop/service";
import type { LogEntryDTO } from "../../bindings/github.com/k3desktop/k3desktop/dto";

const MAX = 500;

export const logEntries = writable<LogEntryDTO[]>([]);

// Per-target log lines — keyed by operation target (cluster/node name).
// OpLog reads from this directly; no per-instance event subscriptions needed.
export const opLogs = writable<Map<string, LogEntryDTO[]>>(new Map());

function push(entry: LogEntryDTO) {
  logEntries.update((entries) => {
    const next = [...entries, entry];
    return next.length > MAX ? next.slice(next.length - MAX) : next;
  });
  if (entry.target) {
    opLogs.update((map) => {
      const clone = new Map(map);
      const existing = clone.get(entry.target) ?? [];
      clone.set(entry.target, [...existing, entry]);
      return clone;
    });
  }
}

export function clearLogs() {
  logEntries.set([]);
}

export function clearOpLog(target: string) {
  opLogs.update((map) => {
    const clone = new Map(map);
    clone.set(target, []);
    return clone;
  });
}

let bootstrapped = false;

export async function bootstrapLogs() {
  if (bootstrapped) return;
  bootstrapped = true;
  try {
    const recent = await LogService.GetRecentLogs();
    if (recent?.length) logEntries.set(recent);
  } catch { /* non-fatal */ }
  Events.On("log:entry", (ev: any) => {
    const entry: LogEntryDTO = ev?.data ?? ev;
    if (entry?.message) push(entry);
  });
}
