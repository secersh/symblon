<script lang="ts">
	import { page } from '$app/stores';
	import { orgsStore } from '$lib/stores/orgs';
	import { orgInitials } from '$lib/utils/avatar';
	import { goto } from '$app/navigation';

	let { children } = $props();

	let orgSlug = $derived($page.params.orgSlug);
	let org = $derived($orgsStore.find((o) => o.slug === orgSlug) ?? null);

	// Active tab from URL
	let subpath = $derived($page.url.pathname.replace(`/${orgSlug}`, '').replace(/^\//, '') || 'overview');

	const tabs = [
		{ id: 'overview', label: 'Overview', href: '' },
		{ id: 'symbols', label: 'Symbols', href: '/symbols' },
		{ id: 'members', label: 'Members', href: '/members' },
		{ id: 'agents', label: 'Agents', href: '/agents' },
		{ id: 'settings', label: 'Settings', href: '/settings' }
	];

	const planColors: Record<string, string> = {
		free: 'badge-ghost',
		pro: 'badge-primary',
		enterprise: 'badge-secondary'
	};
</script>

{#if !org}
	<div class="min-h-full bg-base-200 flex items-center justify-center">
		<div class="text-center">
			<p class="text-4xl mb-4">✦</p>
			<h1 class="text-xl font-bold mb-2">Organization not found</h1>
			<p class="text-base-content/60 text-sm mb-6">
				<code class="font-mono bg-base-300 px-2 py-0.5 rounded">/{orgSlug}</code> doesn't exist in your account.
			</p>
			<a href="/" class="btn btn-primary btn-sm">Back to dashboard</a>
		</div>
	</div>
{:else}
	<div class="min-h-full bg-base-200">
		<!-- Org header bar -->
		<div class="bg-base-100 border-b border-base-200 px-4 lg:px-8">
			<div class="container mx-auto max-w-6xl">
				<!-- Org identity row -->
				<div class="flex items-center gap-3 py-4">
					<div
						class="w-10 h-10 rounded-xl flex items-center justify-center text-white text-sm font-bold shrink-0"
						style="background-color: {org.avatarColor}"
					>
						{orgInitials(org.name)}
					</div>
					<div class="min-w-0 flex-1">
						<div class="flex items-center gap-2 flex-wrap">
							<h1 class="text-base font-bold truncate">{org.name}</h1>
							<span class="badge badge-sm {planColors[org.plan]} capitalize">{org.plan}</span>
							{#if org.linkedGithubOrg}
								<span class="badge badge-sm badge-ghost gap-1">
									<svg class="w-3 h-3" fill="currentColor" viewBox="0 0 24 24">
										<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
									</svg>
									{org.linkedGithubOrg}
								</span>
							{/if}
						</div>
						<p class="text-xs text-base-content/50">{org.memberCount} member{org.memberCount !== 1 ? 's' : ''} · <span class="capitalize">{org.role}</span></p>
					</div>
				</div>

				<!-- Tab navigation -->
				<div class="flex gap-0 -mb-px overflow-x-auto">
					{#each tabs as tab}
						<a
							href="/{orgSlug}{tab.href}"
							class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors whitespace-nowrap
								{subpath === tab.id
									? 'border-primary text-primary'
									: 'border-transparent text-base-content/60 hover:text-base-content hover:border-base-300'}"
						>
							{tab.label}
						</a>
					{/each}
				</div>
			</div>
		</div>

		<!-- Page content -->
		<div class="container mx-auto max-w-6xl px-4 py-8">
			{@render children()}
		</div>
	</div>
{/if}
