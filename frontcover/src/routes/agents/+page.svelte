<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData } from './$types';
	import type { Agent } from '$lib/api/registrar';

	let { data }: { data: PageData } = $props();

	type AgentWithInstalled = Agent & { installed: boolean; installedVersion: string | null };

	let agents = $state<AgentWithInstalled[]>([...data.agents]);

	type Tab = 'mine' | 'discover';
	let tab = $state<Tab>('mine');
	let expandedSymbols = $state<Set<string>>(new Set());

	let myAgents = $derived(agents.filter((a) => a.installed));
	let catalogAgents = $derived(agents.filter((a) => !a.installed));

	function toggleSymbols(id: string) {
		expandedSymbols = new Set(
			expandedSymbols.has(id)
				? [...expandedSymbols].filter((x) => x !== id)
				: [...expandedSymbols, id]
		);
	}

	function hasUpdate(agent: AgentWithInstalled) {
		return agent.installed && agent.installedVersion !== null && agent.installedVersion !== agent.version;
	}
</script>

<div class="min-h-full bg-base-200">
	<main class="container mx-auto px-4 py-8 max-w-4xl">

		<div class="mb-6">
			<h1 class="text-xl font-bold">Agents</h1>
			<p class="text-sm text-base-content/50 mt-0.5">
				Agents evaluate your GitHub activity and issue symbols. {myAgents.length} installed.
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
						{@const needsUpdate = hasUpdate(agent)}
						<div class="card bg-base-100 border {needsUpdate ? 'border-warning/50' : 'border-base-200'}">
							<div class="card-body p-4">
								<div class="flex items-start gap-4">
									<div class="flex-1 min-w-0">
										<div class="flex items-center gap-2 flex-wrap mb-0.5">
											<h3 class="text-sm font-semibold">{agent.name}</h3>
											<span class="badge badge-sm badge-ghost text-[10px] font-mono">
												v{agent.installedVersion ?? agent.version}
											</span>
											{#if needsUpdate}
												<span class="badge badge-sm badge-warning text-[10px]">
													v{agent.version} available
												</span>
											{/if}
											<span class="badge badge-sm badge-ghost text-[10px]">
												{agent.symbols.some(s => s.type === 'temporal') ? '⏱ temporal' : '⚡ real-time'}
											</span>
											{#if agent.pricing_model === 'paid'}
												<span class="badge badge-sm badge-warning text-[10px]">
													{agent.price_usd ? `$${agent.price_usd}` : 'Paid'}
												</span>
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
												by <span class="font-medium">{agent.publisher}</span>
											</span>
										</div>
									</div>
									<div class="flex flex-col gap-1 shrink-0">
										{#if needsUpdate}
											<form
												method="POST"
												action="?/update"
												use:enhance={() => {
													agents = agents.map(a =>
														a.id === agent.id
															? { ...a, installedVersion: a.version }
															: a
													);
													return async ({ update }) => update({ reset: false });
												}}
											>
												<input type="hidden" name="publisher" value={agent.publisher} />
												<input type="hidden" name="handle" value={agent.handle} />
												<input type="hidden" name="old_version" value={agent.installedVersion} />
												<input type="hidden" name="new_version" value={agent.version} />
												<button class="btn btn-xs btn-warning" type="submit">
													Update
												</button>
											</form>
										{/if}
										<form
											method="POST"
											action="?/uninstall"
											use:enhance={() => {
												agents = agents.map(a => a.id === agent.id ? { ...a, installed: false } : a);
												return async ({ update }) => update({ reset: false });
											}}
										>
											<input type="hidden" name="publisher" value={agent.publisher} />
											<input type="hidden" name="handle" value={agent.handle} />
											<input type="hidden" name="version" value={agent.installedVersion ?? agent.version} />
											<button class="btn btn-xs btn-ghost text-error opacity-40 hover:opacity-100" type="submit">
												Remove
											</button>
										</form>
									</div>
								</div>

								{#if open}
									<div class="mt-3 pt-3 border-t border-base-200 grid grid-cols-1 sm:grid-cols-2 gap-2">
										{#each agent.symbols as s}
											<div class="flex items-center gap-2.5 p-2 rounded-lg bg-base-200">
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
											<span class="badge badge-sm badge-ghost text-[10px] font-mono">
												v{agent.version}
											</span>
											<span class="badge badge-sm badge-ghost text-[10px]">
												{agent.symbols.some(s => s.type === 'temporal') ? '⏱ temporal' : '⚡ real-time'}
											</span>
											{#if agent.pricing_model === 'paid'}
												<span class="badge badge-sm badge-warning text-[10px]">
													{agent.price_usd ? `$${agent.price_usd}` : 'Paid'}
												</span>
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
												by <span class="font-medium">{agent.publisher}</span>
											</span>
										</div>
									</div>
									<form
										method="POST"
										action="?/install"
										use:enhance={() => {
											agents = agents.map(a => a.id === agent.id ? { ...a, installed: true, installedVersion: a.version } : a);
											tab = 'mine';
											return async ({ update }) => update({ reset: false });
										}}
									>
										<input type="hidden" name="publisher" value={agent.publisher} />
										<input type="hidden" name="handle" value={agent.handle} />
										<input type="hidden" name="version" value={agent.version} />
										<button class="btn btn-sm btn-outline shrink-0" type="submit">
											Install
										</button>
									</form>
								</div>

								{#if open}
									<div class="mt-3 pt-3 border-t border-base-200 grid grid-cols-1 sm:grid-cols-2 gap-2">
										{#each agent.symbols as s}
											<div class="flex items-center gap-2.5 p-2 rounded-lg bg-base-200">
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
