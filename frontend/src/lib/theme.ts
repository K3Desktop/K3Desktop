import { writable } from "svelte/store";

const stored = localStorage.getItem("theme");
const prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;

export const dark = writable(stored ? stored === "dark" : prefersDark);

dark.subscribe((v) => {
  document.documentElement.classList.toggle("dark", v);
  localStorage.setItem("theme", v ? "dark" : "light");
});

const storedWidth = localStorage.getItem("sidebar:width");
const storedCollapsed = localStorage.getItem("sidebar:collapsed");

export const sidebarWidth = writable(storedWidth ? Number(storedWidth) : 224);
export const sidebarCollapsed = writable(storedCollapsed === "true");

sidebarWidth.subscribe((v) => localStorage.setItem("sidebar:width", String(v)));
sidebarCollapsed.subscribe((v) => localStorage.setItem("sidebar:collapsed", String(v)));
