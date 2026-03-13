<script lang="ts">
	import { page } from '$app/stores';
	import { orgsStore } from '$lib/stores/orgs';

	let orgSlug = $derived($page.params.orgSlug);
	let org = $derived($orgsStore.find((o) => o.slug === orgSlug) ?? null);

	const stats = [
		{ label: 'Symbols', value: '0', icon: '✦' },
		{ label: 'Agents', value: '0', icon: '⚡' },
		{ label: 'Events', value: '0', icon: '📡' },
	];
</script>

{#if org}
	<!-- Stats -->
	<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-8">
		{#each [
			{ label: 'Symbols', value: '0' },
			{ label: 'Members', value: String(org.memberCount) },
			{ label: 'Agents', value: '0' },
			{ label: 'Events processed', value: '0' },
		] as stat}
			<div class="card bg-base-100 border border-base-200">
				<div class="card-body p-4">
					<p class="text-2xl font-bold">{stat.value}</p>
					<p class="text-xs text-base-content/50">{stat.label}</p>
				</div>
			</div>
		{/each}
	</div>

	<!-- Two-column layout -->
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- Recent symbols (2/3 width) -->
		<div class="lg:col-span-2 space-y-4">
			<div class="flex items-center justify-between">
				<h2 class="text-sm font-semibold">Recent symbols</h2>
				<a href="/{org.slug}/symbols" class="text-xs text-primary hover:underline">View all</a>
			</div>
			<div class="card bg-base-100 border border-dashed border-base-300 min-h-[200px]">
				<div class="card-body items-center justify-center text-base-content/30 gap-2">
					<span class="text-4xl">✦</span>
					<p class="text-sm">No symbols created yet</p>
					<a href="/{org.slug}/symbols" class="btn btn-xs btn-ghost text-primary mt-1">Create first symbol</a>
				</div>
			</div>
		</div>

		<!-- Members panel (1/3 width) -->
		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<h2 class="text-sm font-semibold">Members</h2>
				<a href="/{org.slug}/members" class="text-xs text-primary hover:underline">Manage</a>
			</div>
			<div class="card bg-base-100 border border-base-200">
				<div class="card-body p-4 gap-3">
					<!-- Current user as only member -->
					<div class="flex items-center gap-3">
						<div class="w-8 h-8 rounded-full bg-neutral flex items-center justify-center text-xs text-neutral-content font-bold shrink-0">
							You
						</div>
						<div class="min-w-0 flex-1">
							<p class="text-sm font-medium truncate">You</p>
							<p class="text-xs text-base-content/50 capitalize">{org.role}</p>
						</div>
					</div>

					{#if org.role === 'owner' || org.role === 'admin'}
						<div class="border-t border-base-200 pt-3">
							<button class="btn btn-sm btn-ghost w-full text-primary">
								<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
									<path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
								</svg>
								Invite member
							</button>
						</div>
					{/if}
				</div>
			</div>

			<!-- Agents panel -->
			<div class="flex items-center justify-between mt-2">
				<h2 class="text-sm font-semibold">Agents</h2>
				<a href="/{org.slug}/agents" class="text-xs text-primary hover:underline">View all</a>
			</div>
			<div class="card bg-base-100 border border-dashed border-base-300">
				<div class="card-body p-4 items-center text-base-content/30 gap-1.5">
					<span class="text-2xl">⚡</span>
					<p class="text-xs">No agents registered</p>
					<a href="/{org.slug}/agents" class="btn btn-xs btn-ghost text-primary">Register agent</a>
				</div>
			</div>
		</div>
	</div>
{/if}
