<script lang="ts">
import SymbolCard from '$lib/components/SymbolCard.svelte';
import { mockSymbols, groupByOrg, groupByRepo } from '$lib/mock/symbols';
import { orgsStore } from '$lib/stores/orgs';
import { themeStore } from '$lib/stores/theme';

let { data } = $props();
let { user } = $derived(data);

let username = $derived(
(user?.user_metadata?.user_name as string | undefined) ??
(user?.user_metadata?.full_name as string | undefined) ??
'there'
);

let isConnected = $derived(!!user?.user_metadata?.github_installation_id);

type Grouping = 'flat' | 'org' | 'repo';
let grouping = $state<Grouping>('flat');

let orgCount = $derived($orgsStore.length);
let symbols = mockSymbols;
let theme = $derived($themeStore);

let grouped = $derived.by(() => {
if (grouping === 'org') return Object.entries(groupByOrg(symbols));
if (grouping === 'repo') return Object.entries(groupByRepo(symbols));
return [['', symbols]] as [string, typeof symbols][];
});
</script>

<div class="min-h-full bg-base-200">
<main class="container mx-auto px-4 py-8 max-w-5xl">

{#if !isConnected}
<div class="mb-8">
<h1 class="text-2xl font-bold">Welcome to symblon, {username} 👋</h1>
<p class="text-base-content/60 mt-1 text-sm">Let's get your GitHub activity connected.</p>
</div>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
<div class="lg:col-span-2 card bg-base-100 border border-base-200">
<div class="card-body p-6">
<div class="flex items-center gap-3 mb-5">
<div class="w-12 h-12 rounded-xl bg-base-200 flex items-center justify-center shrink-0">
<svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
</svg>
</div>
<div>
<h2 class="font-semibold">Connect GitHub</h2>
<p class="text-xs text-base-content/50">Install the GitHub App to start earning symbols</p>
</div>
</div>
<ol class="space-y-4 mb-6">
{#each [
{ n: 1, title: 'Install the GitHub App', desc: 'Grant Symblon access to the repositories you want to track.' },
{ n: 2, title: 'Activity gets recorded', desc: 'Symblon listens to events: PRs merged, issues closed, reviews submitted, and more.' },
{ n: 3, title: 'Symbols are issued', desc: 'Agents evaluate your activity in real-time and over time windows.' },
] as step}
<li class="flex items-start gap-4">
<span class="w-6 h-6 rounded-full bg-primary/10 text-primary text-xs font-bold flex items-center justify-center shrink-0 mt-0.5">{step.n}</span>
<div>
<p class="text-sm font-medium">{step.title}</p>
<p class="text-sm text-base-content/50 mt-0.5">{step.desc}</p>
</div>
</li>
{/each}
</ol>
<a href="/settings" class="btn btn-primary gap-2">
<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
<path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"/>
</svg>
Set up GitHub integration
</a>
</div>
</div>
<div class="card bg-base-100 border border-base-200">
<div class="card-body p-5">
<h3 class="text-sm font-semibold mb-3">What you can earn</h3>
<div class="space-y-2">
{#each [
{ icon: '🔀', name: 'First Merge', desc: 'Merge your first PR' },
{ icon: '🐛', name: 'Bug Squasher', desc: 'Close 5 bugs in 48h' },
{ icon: '👁️', name: 'Code Reviewer', desc: 'Review 10 PRs in a week' },
{ icon: '🔥', name: 'Streak Master', desc: 'Commit 7 days in a row' },
] as s}
<div class="flex items-center gap-2.5 p-2 rounded-lg bg-base-200">
<span class="text-lg">{s.icon}</span>
<div class="min-w-0">
<p class="text-xs font-medium truncate">{s.name}</p>
<p class="text-[11px] text-base-content/50 truncate">{s.desc}</p>
</div>
</div>
{/each}
</div>
<p class="text-[11px] text-base-content/40 mt-3">These are examples. Agents define the actual criteria.</p>
</div>
</div>
</div>

{:else}
<div class="mb-6">
<h1 class="text-2xl font-bold">Hey, {username} 👋</h1>
<p class="text-base-content/60 mt-1 text-sm">Here are your symbols.</p>
</div>

<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-8">
	<div class="card bg-base-100 border border-base-200">
		<div class="card-body p-4">
			<p class="text-2xl font-bold">{symbols.length}</p>
			<p class="text-xs text-base-content/50">Symbols earned</p>
		</div>
	</div>
	<div class="card bg-base-100 border border-base-200">
		<div class="card-body p-4">
			<p class="text-2xl font-bold">2</p>
			<p class="text-xs text-base-content/50">Agents active</p>
		</div>
	</div>
	<div class="card bg-base-100 border border-base-200">
		<div class="card-body p-4">
			<p class="text-2xl font-bold">3</p>
			<p class="text-xs text-base-content/50">Repos tracked</p>
		</div>
	</div>
	<div class="card bg-base-100 border border-base-200">
		<div class="card-body p-4">
			<p class="text-2xl font-bold">{orgCount}</p>
			<p class="text-xs text-base-content/50">Organizations</p>
		</div>
	</div>
</div>

<div class="flex items-center justify-between gap-4 mb-4 flex-wrap">
<h2 class="text-base font-semibold">Your Symbols</h2>
<div class="flex items-center gap-1 bg-base-100 border border-base-200 rounded-lg p-1">
{#each [
{ id: 'flat', label: 'All' },
{ id: 'org',  label: 'By org' },
{ id: 'repo', label: 'By repo' },
] as opt}
<button
class="px-3 py-1 rounded-md text-xs font-medium transition-colors
{grouping === opt.id ? 'bg-base-200 text-base-content' : 'text-base-content/50 hover:text-base-content'}"
onclick={() => grouping = opt.id as Grouping}
>
{opt.label}
</button>
{/each}
</div>
</div>

{#each grouped as [label, group]}
{#if label}
<h3 class="text-xs font-semibold uppercase tracking-wider text-base-content/40 mb-3 mt-6 first:mt-0">
{label}
<span class="font-normal text-base-content/30">({group.length})</span>
</h3>
{/if}
<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
{#each group as symbol}
<SymbolCard {symbol} {theme} />
{/each}
</div>
{/each}
{/if}

</main>
</div>
