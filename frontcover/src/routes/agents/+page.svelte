<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData } from './$types';
	import type { Agent } from '$lib/api/registrar';

	let { data }: { data: PageData } = $props();

	type AgentWithInstalled = Agent & { installed: boolean; owned: boolean; installedVersion: string | null };

	let agents = $state<AgentWithInstalled[]>([...data.agents]);

	type Tab = 'mine' | 'discover';
	let tab = $state<Tab>('mine');
	let expandedSymbols = $state<Set<string>>(new Set());

	let myAgents = $derived(agents.filter((a) => a.installed));
	let catalogAgents = $derived(agents.filter((a) => !a.installed));

	// ── Dialog state ──────────────────────────────────────────────────────────
	let pendingInstall = $state<AgentWithInstalled | null>(null);
	let pendingRemove  = $state<AgentWithInstalled | null>(null);
	let pendingUpdate  = $state<AgentWithInstalled | null>(null);

	let installDialog: HTMLDialogElement;
	let removeDialog:  HTMLDialogElement;
	let updateDialog:  HTMLDialogElement;

	let installForm: HTMLFormElement;
	let removeForm:  HTMLFormElement;
	let updateForm:  HTMLFormElement;

	function openInstall(agent: AgentWithInstalled) {
		pendingInstall = agent;
		installDialog.showModal();
	}

	function openRemove(agent: AgentWithInstalled) {
		pendingRemove = agent;
		removeDialog.showModal();
	}

	function openUpdate(agent: AgentWithInstalled) {
		pendingUpdate = agent;
		updateDialog.showModal();
	}

	function confirmInstall() {
		installDialog.close();
		agents = agents.map(a =>
			a.id === pendingInstall!.id
				? { ...a, installed: true, installedVersion: a.version }
				: a
		);
		tab = 'mine';
		installForm.requestSubmit();
	}

	function confirmRemove() {
		removeDialog.close();
		agents = agents.map(a => a.id === pendingRemove!.id ? { ...a, installed: false } : a);
		// owned stays true — paid agents remain in user's library after removal
		removeForm.requestSubmit();
	}

	function confirmUpdate() {
		updateDialog.close();
		agents = agents.map(a =>
			a.id === pendingUpdate!.id
				? { ...a, installedVersion: a.version }
				: a
		);
		updateForm.requestSubmit();
	}

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

<!-- ── Install dialog ──────────────────────────────────────────────────────── -->
<dialog bind:this={installDialog} class="modal">
	<div class="modal-box max-w-sm">
		{#if pendingInstall}
			<h3 class="font-bold text-base mb-1">Install {pendingInstall.name}</h3>
			<p class="text-xs text-base-content/60 mb-4">{pendingInstall.description}</p>

			{#if pendingInstall.symbols.length > 0}
				<div class="mb-4 space-y-1.5">
					{#each pendingInstall.symbols as s}
						<div class="flex items-center gap-2 p-2 rounded-lg bg-base-200">
							{#if s.image_url}
								<img src={s.image_url} alt={s.name} class="w-6 h-6 shrink-0" />
							{/if}
							<div class="min-w-0">
								<p class="text-xs font-semibold truncate">{s.name}</p>
								<p class="text-[11px] text-base-content/50 truncate">{s.description}</p>
							</div>
						</div>
					{/each}
				</div>
			{/if}

			{#if pendingInstall.pricing_model === 'paid' && !pendingInstall.owned}
				<div class="rounded-xl border border-base-300 p-3 mb-4 space-y-3">
					<div class="flex items-center justify-between">
						<span class="text-xs font-semibold">Payment</span>
						<span class="badge badge-sm badge-warning">
							${pendingInstall.price_usd}
						</span>
					</div>

					<!-- Mock card UI -->
					<div class="space-y-2">
						<div>
							<label class="text-[11px] text-base-content/50 mb-0.5 block">Card number</label>
							<input
								type="text"
								class="input input-bordered input-sm w-full font-mono text-xs"
								placeholder="4242 4242 4242 4242"
								maxlength="19"
								disabled
							/>
						</div>
						<div class="grid grid-cols-2 gap-2">
							<div>
								<label class="text-[11px] text-base-content/50 mb-0.5 block">Expiry</label>
								<input
									type="text"
									class="input input-bordered input-sm w-full font-mono text-xs"
									placeholder="MM / YY"
									maxlength="7"
									disabled
								/>
							</div>
							<div>
								<label class="text-[11px] text-base-content/50 mb-0.5 block">CVC</label>
								<input
									type="text"
									class="input input-bordered input-sm w-full font-mono text-xs"
									placeholder="•••"
									maxlength="3"
									disabled
								/>
							</div>
						</div>
					</div>

					<p class="text-[10px] text-base-content/40 flex items-center gap-1">
						<svg class="w-3 h-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
						</svg>
						Stripe payments — coming soon. One-time purchase, no subscription.
					</p>
				</div>
			{/if}

			<div class="modal-action mt-2">
				<button class="btn btn-sm btn-ghost" onclick={() => installDialog.close()}>Cancel</button>
				<button
					class="btn btn-sm btn-primary"
					onclick={confirmInstall}
				>
					{pendingInstall.pricing_model === 'paid' && !pendingInstall.owned ? `Install · $${pendingInstall.price_usd}` : 'Install'}
				</button>
			</div>
		{/if}
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>close</button>
	</form>
</dialog>

<!-- ── Remove dialog ───────────────────────────────────────────────────────── -->
<dialog bind:this={removeDialog} class="modal">
	<div class="modal-box max-w-sm">
		{#if pendingRemove}
			<h3 class="font-bold text-base mb-1">Remove {pendingRemove.name}?</h3>
			<p class="text-xs text-base-content/60 mb-3">
				Your GitHub activity will no longer be evaluated by this agent and no new symbols will be issued.
				Symbols you've already earned are yours to keep.
			</p>
			{#if pendingRemove.pricing_model === 'paid'}
				<div class="alert alert-info py-2 px-3 text-xs mb-3">
					<svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
					</svg>
					You own this agent. You can reinstall it any time from Discover — no need to pay again.
				</div>
			{/if}
			<div class="modal-action mt-2">
				<button class="btn btn-sm btn-ghost" onclick={() => removeDialog.close()}>Keep it</button>
				<button class="btn btn-sm btn-error" onclick={confirmRemove}>Remove</button>
			</div>
		{/if}
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>close</button>
	</form>
</dialog>

<!-- ── Update dialog ───────────────────────────────────────────────────────── -->
<dialog bind:this={updateDialog} class="modal">
	<div class="modal-box max-w-sm">
		{#if pendingUpdate}
			<h3 class="font-bold text-base mb-1">Update {pendingUpdate.name}</h3>
			<p class="text-xs text-base-content/60 mb-4">
				Upgrade from
				<span class="font-mono badge badge-sm badge-ghost">v{pendingUpdate.installedVersion}</span>
				to
				<span class="font-mono badge badge-sm badge-primary">v{pendingUpdate.version}</span>.
			</p>
			<div class="modal-action mt-2">
				<button class="btn btn-sm btn-ghost" onclick={() => updateDialog.close()}>Cancel</button>
				<button class="btn btn-sm btn-warning" onclick={confirmUpdate}>Update</button>
			</div>
		{/if}
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>close</button>
	</form>
</dialog>

<!-- ── Hidden forms (submitted programmatically on dialog confirm) ─────────── -->
<form
	bind:this={installForm}
	method="POST"
	action="?/install"
	use:enhance={() => async ({ update }) => update({ reset: false })}
	class="hidden"
>
	<input type="hidden" name="publisher" value={pendingInstall?.publisher ?? ''} />
	<input type="hidden" name="handle"    value={pendingInstall?.handle ?? ''} />
	<input type="hidden" name="version"   value={pendingInstall?.version ?? ''} />
</form>

<form
	bind:this={removeForm}
	method="POST"
	action="?/uninstall"
	use:enhance={() => async ({ update }) => update({ reset: false })}
	class="hidden"
>
	<input type="hidden" name="publisher" value={pendingRemove?.publisher ?? ''} />
	<input type="hidden" name="handle"    value={pendingRemove?.handle ?? ''} />
	<input type="hidden" name="version"   value={pendingRemove?.installedVersion ?? pendingRemove?.version ?? ''} />
</form>

<form
	bind:this={updateForm}
	method="POST"
	action="?/update"
	use:enhance={() => async ({ update }) => update({ reset: false })}
	class="hidden"
>
	<input type="hidden" name="publisher"    value={pendingUpdate?.publisher ?? ''} />
	<input type="hidden" name="handle"       value={pendingUpdate?.handle ?? ''} />
	<input type="hidden" name="old_version"  value={pendingUpdate?.installedVersion ?? ''} />
	<input type="hidden" name="new_version"  value={pendingUpdate?.version ?? ''} />
</form>

<!-- ── Page ───────────────────────────────────────────────────────────────── -->
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
												by <span class="font-medium">{agent.publisher_name || agent.publisher}</span>
											</span>
										</div>
									</div>
									<div class="flex flex-col gap-1 shrink-0">
										{#if needsUpdate}
											<button
												class="btn btn-xs btn-warning"
												onclick={() => openUpdate(agent)}
											>
												Update
											</button>
										{/if}
										<button
											class="btn btn-xs btn-ghost text-error opacity-40 hover:opacity-100"
											onclick={() => openRemove(agent)}
										>
											Remove
										</button>
									</div>
								</div>

								{#if open}
									<div class="mt-3 pt-3 border-t border-base-200 grid grid-cols-1 sm:grid-cols-2 gap-2">
										{#each agent.symbols as s}
											<div class="flex items-center gap-2.5 p-2 rounded-lg bg-base-200">
												{#if s.image_url}
													<img src={s.image_url} alt={s.name} class="w-8 h-8 shrink-0" />
												{/if}
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
												{#if agent.owned}
													<span class="badge badge-sm badge-success text-[10px]">Purchased</span>
												{/if}
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
												by <span class="font-medium">{agent.publisher_name || agent.publisher}</span>
											</span>
										</div>
									</div>
									<button
										class="btn btn-sm btn-outline shrink-0"
										onclick={() => openInstall(agent)}
									>
										Install
									</button>
								</div>

								{#if open}
									<div class="mt-3 pt-3 border-t border-base-200 grid grid-cols-1 sm:grid-cols-2 gap-2">
										{#each agent.symbols as s}
											<div class="flex items-center gap-2.5 p-2 rounded-lg bg-base-200">
												{#if s.image_url}
													<img src={s.image_url} alt={s.name} class="w-8 h-8 shrink-0" />
												{/if}
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
