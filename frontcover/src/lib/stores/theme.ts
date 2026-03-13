import { browser } from '$app/environment';
import { writable } from 'svelte/store';
import type { SymbolTheme } from '$lib/types/symbol';

const stored = browser ? (localStorage.getItem('symbolTheme') as SymbolTheme | null) : null;

export const themeStore = writable<SymbolTheme>(stored ?? 'scouts');

themeStore.subscribe((v) => {
	if (browser) localStorage.setItem('symbolTheme', v);
});
