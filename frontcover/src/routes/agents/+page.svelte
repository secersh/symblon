<script lang="ts">
type AgentKind = 'realtime' | 'temporal';
type AgentPricing = 'free' | 'paid';

interface AgentSymbol {
emoji: string;
name: string;
description: string;
}

interface Agent {
id: string;
name: string;
description: string;
author: string;
kind: AgentKind;
pricing: AgentPricing;
symbols: AgentSymbol[];
installed: boolean;
enabled: boolean;
}

let agents = $state<Agent[]>([
{
id: 'core-realtime',
name: 'Core Real-time',
description: 'Issues symbols for common GitHub events: first PR, first merge, first issue, first review.',
author: 'symblon',
kind: 'realtime',
pricing: 'free',
installed: true,
enabled: true,
symbols: [
{ emoji: '🔀', name: 'First Merge',    description: 'Merged your first pull request' },
{ emoji: '🐣', name: 'First Issue',    description: 'Opened your first issue' },
{ emoji: '👁️', name: 'First Review',   description: 'Submitted your first PR review' },
{ emoji: '⭐', name: 'First Star',     description: 'Starred your first repository' },
{ emoji: '🍴', name: 'First Fork',     description: 'Forked a repository for the first time' },
{ emoji: '💬', name: 'First Comment',  description: 'Left your first issue comment' },
{ emoji: '📦', name: 'First Release',  description: 'Published your first GitHub release' },
{ emoji: '🏷️', name: 'First Tag',      description: 'Created your first git tag' },
]
},
{
id: 'streak-tracker',
name: 'Streak Tracker',
description: 'Tracks commit streaks and awards escalating symbols for consecutive active days.',
author: 'symblon',
kind: 'temporal',
pricing: 'free',
installed: true,
enabled: true,
symbols: [
{ emoji: '🔥', name: 'Streak 3',   description: '3 consecutive days with commits' },
{ emoji: '🔥🔥', name: 'Streak 7',  description: '7 consecutive days with commits' },
{ emoji: '🌋', name: 'Streak 14',  description: '14 consecutive days with commits' },
{ emoji: '💫', name: 'Streak 30',  description: '30 consecutive days with commits' },
{ emoji: '🏔️', name: 'Streak 100', description: '100 consecutive days with commits' },
]
},
{
id: 'bug-hunter',
name: 'Bug Hunter',
description: 'Evaluates bug-fix velocity. Rewards closing multiple bug-labelled issues within time windows.',
author: 'symblon',
kind: 'temporal',
pricing: 'free',
installed: false,
enabled: false,
symbols: [
{ emoji: '🐛', name: 'Bug Squasher',  description: 'Closed 5 bug-labelled issues within 48 hours' },
{ emoji: '🦟', name: 'Exterminator',  description: 'Closed 20 bug-labelled issues in a week' },
{ emoji: '🔬', name: 'Root Causer',   description: 'Diagnosed and fixed a regression' },
{ emoji: '🛡️', name: 'Zero Defects',  description: 'No bug-labelled issues opened for 30 days' },
]
},
{
id: 'review-master',
name: 'Review Master',
description: 'Tracks PR review activity. Awards symbols for volume, quality signals, and cross-repo reviews.',
author: 'symblon',
kind: 'temporal',
pricing: 'free',
installed: false,
enabled: false,
symbols: [
{ emoji: '🔍', name: 'Reviewer',       description: 'Reviewed 10 pull requests in a week' },
{ emoji: '🧐', name: 'Nitpicker',      description: 'Left 50 review comments in a month' },
{ emoji: '✅', name: 'Approver',       description: 'Approved 25 PRs without any being reverted' },
{ emoji: '🌉', name: 'Cross-repo',     description: 'Reviewed PRs in 3 different repositories in one week' },
{ emoji: '🎓', name: 'Mentor',         description: 'Had 10 of your review suggestions accepted' },
{ emoji: '⚡', name: 'Fast Reviewer',  description: 'Reviewed a PR within 30 minutes of it opening' },
]
},
{
id: 'night-coder',
name: 'Night Coder',
description: 'Awards symbols for activity outside typical working hours. For the dedicated night owls.',
author: 'community',
kind: 'realtime',
pricing: 'free',
installed: false,
enabled: false,
symbols: [
{ emoji: '🦉', name: 'Night Owl',     description: 'Merged a PR between midnight and 5am' },
{ emoji: '🌙', name: 'Midnight Push',  description: 'Pushed commits at midnight on 5 separate nights' },
{ emoji: '🌅', name: 'Early Bird',     description: 'Committed before 6am on 3 consecutive days' },
]
},
{
id: 'polyglot-pro',
name: 'Polyglot Pro',
description: 'Deep analysis of language diversity across repositories. Tracks progression from generalist to specialist.',
author: 'community',
kind: 'temporal',
pricing: 'paid',
installed: false,
enabled: false,
symbols: [
{ emoji: '🌐', name: 'Polyglot',     description: 'Contributed to repos in 3 different languages' },
{ emoji: '🗣️', name: 'Multilingual',  description: 'Contributed to repos in 5 different languages' },
{ emoji: '🔤', name: 'Linguist',      description: 'Primary language expert — top 10 contributors in a language' },
{ emoji: '🧩', name: 'Generalist',    description: 'Active in 10+ languages across the year' },
{ emoji: '🎯', name: 'Specialist',    description: '80%+ of contributions in a single language' },
{ emoji: '🏛️', name: 'Architect',     description: 'Defined language standards for an organization' },
{ emoji: '🔭', name: 'Explorer',      description: 'First contribution in a new-to-you language' },
{ emoji: '🧬', name: 'Polymath',      description: 'Contributions spanning frontend, backend, and infrastructure' },
{ emoji: '📚', name: 'Scholar',       description: 'Documentation contributions in 3+ languages' },
{ emoji: '🌈', name: 'Full Spectrum',  description: 'Every major language category covered' },
]
},
{
id: 'release-engineer',
name: 'Release Engineer',
description: 'Tracks deployment and release activity. Awards symbols for release cadence and reliability.',
author: 'community',
kind: 'temporal',
pricing: 'paid',
installed: false,
enabled: false,
symbols: [
{ emoji: '🚀', name: 'Ship It',       description: 'Published 5 releases in a week' },
{ emoji: '📦', name: 'Packager',      description: 'Maintained a consistent release cadence for 30 days' },
{ emoji: '🏁', name: 'Deployer',      description: 'Successfully deployed to production 10 times' },
{ emoji: '🔄', name: 'Hotfixer',      description: 'Released a hotfix within 1 hour of a bug report' },
{ emoji: '🌊', name: 'Continuous',    description: 'Shipped at least once a day for 7 days' },
{ emoji: '🛸', name: 'Zero Downtime', description: '10 consecutive deployments with no rollbacks' },
{ emoji: '📋', name: 'Changelog Pro', description: 'Every release has a detailed changelog' },
]
},
]);

type Tab = 'mine' | 'discover';
let tab = $state<Tab>('mine');
let expandedSymbols = $state<Set<string>>(new Set());

let myAgents = $derived(agents.filter((a) => a.installed));
let catalogAgents = $derived(agents.filter((a) => !a.installed));
let enabledCount = $derived(agents.filter((a) => a.enabled).length);

function toggle(id: string) {
agents = agents.map((a) => (a.id === id ? { ...a, enabled: !a.enabled } : a));
}

function install(id: string) {
agents = agents.map((a) => (a.id === id ? { ...a, installed: true, enabled: true } : a));
tab = 'mine';
}

function uninstall(id: string) {
agents = agents.map((a) => (a.id === id ? { ...a, installed: false, enabled: false } : a));
}

function toggleSymbols(id: string) {
expandedSymbols = new Set(
expandedSymbols.has(id)
? [...expandedSymbols].filter((x) => x !== id)
: [...expandedSymbols, id]
);
}
</script>

<div class="min-h-full bg-base-200">
<main class="container mx-auto px-4 py-8 max-w-4xl">

<div class="mb-6">
<h1 class="text-xl font-bold">Agents</h1>
<p class="text-sm text-base-content/50 mt-0.5">
Agents evaluate your GitHub activity and issue symbols. {enabledCount} active.
</p>
</div>

<!-- Tabs -->
<div class="flex gap-1 mb-5 border-b border-base-200">
{#each [
{ id: 'mine',     label: 'My Agents',  count: myAgents.length },
{ id: 'discover', label: 'Discover',   count: catalogAgents.length },
] as t}
<button
class="px-4 py-2 text-sm font-medium border-b-2 transition-colors -mb-px
{tab === t.id
? 'border-primary text-primary'
: 'border-transparent text-base-content/50 hover:text-base-content'}"
onclick={() => tab = t.id as Tab}
>
{t.label}
{#if t.count > 0}
<span class="ml-1.5 text-[11px] font-normal opacity-60">{t.count}</span>
{/if}
</button>
{/each}
</div>

{#if tab === 'mine'}
{#if myAgents.length === 0}
<div class="card bg-base-100 border border-dashed border-base-300">
<div class="card-body items-center py-12 text-base-content/30 gap-2">
<span class="text-4xl">⚡</span>
<p class="text-sm font-medium">No agents installed</p>
<p class="text-xs text-center">Browse the catalog to find agents that match your workflow.</p>
<button class="btn btn-sm btn-primary mt-3" onclick={() => tab = 'discover'}>Browse catalog</button>
</div>
</div>
{:else}
<div class="space-y-3">
{#each myAgents as agent}
{@const open = expandedSymbols.has(agent.id)}
<div class="card bg-base-100 border border-base-200">
<div class="card-body p-4">
<div class="flex items-start gap-4">
<div class="pt-0.5 shrink-0">
<input
type="checkbox"
class="toggle toggle-primary toggle-sm"
checked={agent.enabled}
onchange={() => toggle(agent.id)}
/>
</div>
<div class="flex-1 min-w-0">
<div class="flex items-center gap-2 flex-wrap mb-0.5">
<h3 class="text-sm font-semibold">{agent.name}</h3>
<span class="badge badge-sm badge-ghost text-[10px]">
{agent.kind === 'temporal' ? '⏱ temporal' : '⚡ real-time'}
</span>
{#if agent.pricing === 'paid'}
<span class="badge badge-sm badge-warning text-[10px]">Paid</span>
{/if}
</div>
<p class="text-xs text-base-content/60">{agent.description}</p>
<div class="flex items-center gap-3 mt-2">
<button
class="text-[11px] text-primary hover:underline flex items-center gap-1"
onclick={() => toggleSymbols(agent.id)}
>
{agent.symbols.length} symbol{agent.symbols.length !== 1 ? 's' : ''}
<svg class="w-3 h-3 transition-transform {open ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
<path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7"/>
</svg>
</button>
<span class="text-[11px] text-base-content/40">
by <span class="font-medium">{agent.author}</span>
</span>
</div>
</div>
<button
class="btn btn-xs btn-ghost text-error opacity-40 hover:opacity-100 shrink-0 self-start"
onclick={() => uninstall(agent.id)}
title="Remove agent"
>
Remove
</button>
</div>

{#if open}
<div class="mt-3 pt-3 border-t border-base-200 grid grid-cols-1 sm:grid-cols-2 gap-2">
{#each agent.symbols as s}
<div class="flex items-center gap-2.5 p-2 rounded-lg bg-base-200">
<span class="text-base leading-none">{s.emoji}</span>
<div class="min-w-0">
<p class="text-xs font-semibold truncate">{s.name}</p>
<p class="text-[11px] text-base-content/50 truncate">{s.description}</p>
</div>
</div>
{/each}
</div>
{/if}
</div>
</div>
{/each}
</div>
{/if}

{:else}
{#if catalogAgents.length === 0}
<div class="card bg-base-100 border border-base-200">
<div class="card-body items-center py-12 text-base-content/30 gap-2">
<span class="text-4xl">✓</span>
<p class="text-sm font-medium">All available agents are installed</p>
</div>
</div>
{:else}
<div class="space-y-3">
{#each catalogAgents as agent}
{@const open = expandedSymbols.has(agent.id)}
<div class="card bg-base-100 border border-base-200">
<div class="card-body p-4">
<div class="flex items-start gap-4">
<div class="flex-1 min-w-0">
<div class="flex items-center gap-2 flex-wrap mb-0.5">
<h3 class="text-sm font-semibold">{agent.name}</h3>
<span class="badge badge-sm badge-ghost text-[10px]">
{agent.kind === 'temporal' ? '⏱ temporal' : '⚡ real-time'}
</span>
{#if agent.pricing === 'paid'}
<span class="badge badge-sm badge-warning text-[10px]">Paid</span>
{/if}
</div>
<p class="text-xs text-base-content/60">{agent.description}</p>
<div class="flex items-center gap-3 mt-2">
<button
class="text-[11px] text-primary hover:underline flex items-center gap-1"
onclick={() => toggleSymbols(agent.id)}
>
{agent.symbols.length} symbol{agent.symbols.length !== 1 ? 's' : ''}
<svg class="w-3 h-3 transition-transform {open ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
<path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7"/>
</svg>
</button>
<span class="text-[11px] text-base-content/40">
by <span class="font-medium">{agent.author}</span>
</span>
</div>
</div>
<button
class="btn btn-sm btn-outline shrink-0 self-start"
onclick={() => install(agent.id)}
>
Install
</button>
</div>

{#if open}
<div class="mt-3 pt-3 border-t border-base-200 grid grid-cols-1 sm:grid-cols-2 gap-2">
{#each agent.symbols as s}
<div class="flex items-center gap-2.5 p-2 rounded-lg bg-base-200">
<span class="text-base leading-none">{s.emoji}</span>
<div class="min-w-0">
<p class="text-xs font-semibold truncate">{s.name}</p>
<p class="text-[11px] text-base-content/50 truncate">{s.description}</p>
</div>
</div>
{/each}
</div>
{/if}
</div>
</div>
{/each}
</div>
{/if}
{/if}

</main>
</div>
