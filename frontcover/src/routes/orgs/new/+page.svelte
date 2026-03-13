<script lang="ts">
	import { goto } from '$app/navigation';
	import { orgsStore } from '$lib/stores/orgs';
	import { orgAvatarColor, slugify } from '$lib/utils/avatar';
	import { mockGitHubOrgs } from '$lib/mock/github-orgs';
	import type { OrgRole } from '$lib/types/account';

	type Tab = 'github' | 'custom';
	let activeTab = $state<Tab>('github');

	// Custom org form
	let customName = $state('');
	let customSlug = $derived(slugify(customName));
	let slugEdited = $state(false);
	let manualSlug = $state('');
	let finalSlug = $derived(slugEdited ? manualSlug : customSlug);

	// GitHub link form
	let selectedGithubOrg = $state<number | null>(null);

	let creating = $state(false);
	let error = $state('');

	function handleSlugInput(e: Event) {
		slugEdited = true;
		manualSlug = (e.target as HTMLInputElement).value.toLowerCase().replace(/[^a-z0-9-]/g, '');
	}

	function createFromGitHub() {
		if (!selectedGithubOrg) return;
		const ghOrg = mockGitHubOrgs.find((o) => o.id === selectedGithubOrg);
		if (!ghOrg) return;

		creating = true;
		const slug = ghOrg.login;
		orgsStore.addOrg({
			slug,
			name: ghOrg.name,
			avatarColor: orgAvatarColor(slug),
			role: ghOrg.role as OrgRole,
			memberCount: 1,
			plan: 'free',
			linkedGithubOrg: ghOrg.login,
			createdAt: new Date().toISOString()
		});
		goto(`/${slug}`);
	}

	function createCustom() {
		if (!customName.trim() || !finalSlug) return;
		error = '';
		creating = true;

		const slug = finalSlug;
		orgsStore.addOrg({
			slug,
			name: customName.trim(),
			avatarColor: orgAvatarColor(slug),
			role: 'owner',
			memberCount: 1,
			plan: 'free',
			createdAt: new Date().toISOString()
		});
		goto(`/${slug}`);
	}

	const roleLabels: Record<string, string> = {
		owner: 'Owner',
		admin: 'Admin',
		member: 'Member'
	};
</script>

<div class="min-h-full bg-base-200">
	<main class="container mx-auto px-4 py-10 max-w-2xl">
		<!-- Header -->
		<div class="mb-8">
			<a href="/" class="text-sm text-base-content/50 hover:text-base-content flex items-center gap-1 mb-4 w-fit">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7"/>
				</svg>
				Back
			</a>
			<h1 class="text-2xl font-bold">New organization</h1>
			<p class="text-base-content/60 mt-1 text-sm">
				Organizations let teams share symbols, manage agents, and track achievements together.
			</p>
		</div>

		<!-- Tab switcher -->
		<div class="tabs tabs-boxed bg-base-100 border border-base-200 p-1 mb-6 w-fit rounded-xl">
			<button
				class="tab rounded-lg text-sm font-medium transition-colors {activeTab === 'github' ? 'tab-active' : ''}"
				onclick={() => activeTab = 'github'}
			>
				<svg class="w-4 h-4 mr-1.5" fill="currentColor" viewBox="0 0 24 24">
					<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
				</svg>
				Link GitHub org
			</button>
			<button
				class="tab rounded-lg text-sm font-medium transition-colors {activeTab === 'custom' ? 'tab-active' : ''}"
				onclick={() => activeTab = 'custom'}
			>
				<svg class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-2 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
				</svg>
				Create new
			</button>
		</div>

		{#if activeTab === 'github'}
			<!-- GitHub org picker -->
			<div class="card bg-base-100 border border-base-200">
				<div class="card-body p-5">
					<h2 class="font-semibold text-base mb-1">Your GitHub organizations</h2>
					<p class="text-sm text-base-content/50 mb-4">
						Select an org where you have owner or admin access. Members will be able to opt in separately.
					</p>

					<div class="space-y-2">
						{#each mockGitHubOrgs as org}
							{@const isEligible = org.role === 'owner' || org.role === 'admin'}
							<button
								class="w-full flex items-center gap-3 p-3 rounded-xl border transition-all text-left
									{selectedGithubOrg === org.id
										? 'border-primary bg-primary/5'
										: 'border-base-200 hover:border-base-300 hover:bg-base-50'}
									{!isEligible ? 'opacity-50 cursor-not-allowed' : ''}"
								onclick={() => isEligible && (selectedGithubOrg = org.id)}
								disabled={!isEligible}
							>
								<img src={org.avatarUrl} alt={org.name} class="w-10 h-10 rounded-xl" />
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium truncate">{org.name}</p>
									<p class="text-xs text-base-content/50">@{org.login}</p>
								</div>
								<span class="badge badge-sm {org.role === 'owner' ? 'badge-primary' : org.role === 'admin' ? 'badge-secondary' : 'badge-ghost'}">
									{roleLabels[org.role]}
								</span>
								{#if selectedGithubOrg === org.id}
									<svg class="w-5 h-5 text-primary shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
										<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/>
									</svg>
								{:else if !isEligible}
									<svg class="w-4 h-4 text-base-content/30 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
										<path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
									</svg>
								{/if}
							</button>
						{/each}
					</div>

					{#if !mockGitHubOrgs.some(o => o.role === 'owner' || o.role === 'admin')}
						<div class="alert alert-warning mt-3 text-sm">
							You need owner or admin access to link a GitHub organization.
						</div>
					{/if}

					<div class="mt-5">
						<button
							class="btn btn-primary w-full"
							disabled={!selectedGithubOrg || creating}
							onclick={createFromGitHub}
						>
							{creating ? 'Creating…' : 'Link organization'}
						</button>
					</div>
				</div>
			</div>

		{:else}
			<!-- Custom org creation -->
			<div class="card bg-base-100 border border-base-200">
				<div class="card-body p-5">
					<h2 class="font-semibold text-base mb-1">Organization details</h2>
					<p class="text-sm text-base-content/50 mb-4">
						Create a Symblon organization not tied to GitHub. You can link a GitHub org later.
					</p>

					<div class="space-y-4">
						<div class="form-control">
							<label class="label pb-1" for="org-name">
								<span class="label-text text-sm font-medium">Name</span>
							</label>
							<input
								id="org-name"
								type="text"
								class="input input-bordered w-full"
								placeholder="My Awesome Team"
								bind:value={customName}
								maxlength={50}
							/>
						</div>

						<div class="form-control">
							<label class="label pb-1" for="org-slug">
								<span class="label-text text-sm font-medium">Slug</span>
								<span class="label-text-alt text-base-content/40">symblon.cc/<strong>{finalSlug || '…'}</strong></span>
							</label>
							<input
								id="org-slug"
								type="text"
								class="input input-bordered w-full font-mono text-sm"
								value={finalSlug}
								oninput={handleSlugInput}
								placeholder="my-awesome-team"
								maxlength={39}
							/>
							{#if finalSlug && finalSlug.length < 3}
								<p class="text-xs text-error mt-1">Slug must be at least 3 characters.</p>
							{/if}
						</div>
					</div>

					{#if error}
						<div class="alert alert-error mt-4 text-sm">{error}</div>
					{/if}

					<div class="mt-5">
						<button
							class="btn btn-primary w-full"
							disabled={!customName.trim() || finalSlug.length < 3 || creating}
							onclick={createCustom}
						>
							{creating ? 'Creating…' : 'Create organization'}
						</button>
					</div>
				</div>
			</div>
		{/if}

		<!-- Free tier callout -->
		<div class="mt-6 flex items-start gap-3 text-sm text-base-content/50">
			<svg class="w-4 h-4 mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
			</svg>
			<p>Organizations start on the <strong class="text-base-content/80">Free plan</strong> — no credit card required. You can upgrade anytime to unlock more agents, symbols, and members.</p>
		</div>
	</main>
</div>
