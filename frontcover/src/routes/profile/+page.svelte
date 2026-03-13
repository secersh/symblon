<script lang="ts">
	import SymbolCard from '$lib/components/SymbolCard.svelte';
	import { mockSymbols, groupByOrg, groupByRepo } from '$lib/mock/symbols';
	import { orgsStore } from '$lib/stores/orgs';
	import { themeStore } from '$lib/stores/theme';
	import { orgInitials } from '$lib/utils/avatar';

	let { data } = $props();
	let { user } = $derived(data);

	let username = $derived(
		(user?.user_metadata?.user_name as string | undefined) ??
		(user?.user_metadata?.full_name as string | undefined) ??
		'unknown'
	);
	let avatarUrl = $derived(user?.user_metadata?.avatar_url as string | undefined);
	let initials = $derived(username.slice(0, 2).toUpperCase());

	type Grouping = 'flat' | 'org' | 'repo';
	let grouping = $state<Grouping>('flat');

	let symbols = mockSymbols;
	let orgs = $derived($orgsStore);
	let theme = $derived($themeStore);

	let grouped = $derived.by(() => {
		if (grouping === 'org') return Object.entries(groupByOrg(symbols));
		if (grouping === 'repo') return Object.entries(groupByRepo(symbols));
		return [['', symbols]] as [string, typeof symbols][];
	});
</script>

<div class="min-h-full bg-base-200">
	<main class="container mx-auto px-4 py-8 max-w-4xl">

		<!-- Profile header card -->
		<div class="card bg-base-100 border border-base-200 mb-6">
			<div class="card-body p-6">
				<div class="flex flex-col sm:flex-row items-start sm:items-center gap-5">
					<!-- Avatar -->
					<div class="avatar shrink-0">
						<div class="w-20 h-20 rounded-2xl ring ring-base-300 ring-offset-base-100 ring-offset-2">
							{#if avatarUrl}
								<img src={avatarUrl} alt={username} referrerpolicy="no-referrer" />
							{:else}
								<div class="bg-neutral text-neutral-content flex items-center justify-center w-full h-full text-2xl font-bold">
									{initials}
								</div>
							{/if}
						</div>
					</div>

					<!-- Name + meta -->
					<div class="flex-1 min-w-0">
						<h1 class="text-xl font-bold">{username}</h1>
						<p class="text-sm text-base-content/50 mt-0.5">{user?.email}</p>

						<!-- Org membership badges -->
						{#if orgs.length > 0}
							<div class="flex flex-wrap gap-1.5 mt-3">
								{#each orgs as org}
									<a
										href="/{org.slug}"
										class="flex items-center gap-1.5 px-2.5 py-1 rounded-full border border-base-200 bg-base-100 hover:bg-base-200 transition-colors text-xs font-medium"
									>
										<span
											class="w-3.5 h-3.5 rounded-sm flex items-center justify-center text-white text-[8px] font-bold shrink-0"
											style="background-color: {org.avatarColor}"
										>{orgInitials(org.name)}</span>
										{org.name}
									</a>
								{/each}
							</div>
						{/if}
					</div>

					<!-- Quick stats -->
					<div class="flex sm:flex-col gap-4 sm:gap-2 sm:text-right shrink-0">
						<div>
							<p class="text-xl font-bold">{symbols.length}</p>
							<p class="text-xs text-base-content/50">symbols</p>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Symbols section -->
		<div class="flex items-center justify-between gap-4 mb-4 flex-wrap">
			<h2 class="text-base font-semibold">Symbols</h2>
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
				<h3 class="text-xs font-semibold uppercase tracking-wider text-base-content/40 mb-3 mt-6 first:mt-0 flex items-center gap-2">
					<svg class="w-3.5 h-3.5 shrink-0" fill="currentColor" viewBox="0 0 24 24">
						<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
					</svg>
					{label}
					<span class="font-normal text-base-content/30">({group.length})</span>
				</h3>
			{/if}
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each group as symbol}
					<SymbolCard {symbol} {theme} />
				{/each}
			</div>
		{/each}

	</main>
</div>
