<script lang="ts">
	import { page } from '$app/stores';
	import { invalidate } from '$app/navigation';
	import { env } from '$env/dynamic/public';

	let { data } = $props();
	let { user } = $derived(data);

	let installationId = $derived(user?.user_metadata?.github_installation_id as string | undefined);
	let isGithubConnected = $derived(!!installationId);

	let justConnected = $derived($page.url.searchParams.get('github') === 'connected');

	const appSlug = env.PUBLIC_GITHUB_APP_SLUG ?? 'symblon-cc';
	const installUrl = `https://github.com/apps/${appSlug}/installations/new`;
	const manageUrl = `https://github.com/settings/installations`;

	const settingsNav = [
		{ id: 'integrations', label: 'Integrations' },
		{ id: 'profile', label: 'Profile' },
		{ id: 'billing', label: 'Billing' },
	];

	let activeSection = $state('integrations');
	let disconnecting = $state(false);
	let showDisconnectConfirm = $state(false);

	async function disconnect() {
		disconnecting = true;
		await data.supabase.auth.updateUser({
			data: {
				github_installation_id: null,
				github_setup_action: null
			}
		});
		showDisconnectConfirm = false;
		disconnecting = false;
		// Re-run the layout load so user metadata updates in the UI
		invalidate('supabase:auth');
	}

	// Providers registry — GitHub live, rest planned
	const providers = [
		{
			id: 'github',
			name: 'GitHub',
			description: 'Record activity from GitHub repos: PRs, issues, reviews, commits and more.',
			available: true,
		},
		{
			id: 'gitlab',
			name: 'GitLab',
			description: 'Record activity from GitLab projects.',
			available: false,
		},
		{
			id: 'bitbucket',
			name: 'Bitbucket',
			description: 'Record activity from Bitbucket repositories.',
			available: false,
		},
	];
</script>

<div class="min-h-full bg-base-200">
	<main class="container mx-auto px-4 py-8 max-w-5xl">
		<div class="mb-6">
			<h1 class="text-xl font-bold">Settings</h1>
		</div>

		<div class="flex flex-col md:flex-row gap-6">
			<!-- Sidebar nav -->
			<nav class="md:w-48 shrink-0">
				<ul class="space-y-0.5">
					{#each settingsNav as item}
						<li>
							<button
								class="w-full text-left px-3 py-2 rounded-lg text-sm transition-colors
									{activeSection === item.id
										? 'bg-base-100 font-semibold border border-base-200'
										: 'text-base-content/60 hover:bg-base-100 hover:text-base-content'}"
								onclick={() => activeSection = item.id}
							>
								{item.label}
							</button>
						</li>
					{/each}
				</ul>
			</nav>

			<!-- Main panel -->
			<div class="flex-1 min-w-0 space-y-4">

				{#if activeSection === 'integrations'}

					{#if justConnected}
						<div class="alert alert-success text-sm py-3">
							<svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
								<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/>
							</svg>
							GitHub App connected successfully.
						</div>
					{/if}

					<div>
						<h2 class="text-sm font-semibold mb-1">Integrations</h2>
						<p class="text-xs text-base-content/50 mb-4">
							Connect activity sources. Symbols are issued based on events from these providers.
						</p>
					</div>

					{#each providers as provider}
						<div class="card bg-base-100 border border-base-200
							{!provider.available ? 'opacity-60' : ''}">
							<div class="card-body p-5">
								<div class="flex items-start gap-4">
									<!-- Provider logo -->
									<div class="w-10 h-10 rounded-xl bg-base-200 flex items-center justify-center shrink-0 mt-0.5">
										{#if provider.id === 'github'}
											<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
												<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
											</svg>
										{:else if provider.id === 'gitlab'}
											<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
												<path d="M23.955 13.587l-1.342-4.135-2.664-8.189a.455.455 0 00-.867 0L16.418 9.45H7.582L4.918 1.263a.455.455 0 00-.867 0L1.387 9.45.045 13.587a.924.924 0 00.331 1.023L12 23.054l11.624-8.443a.92.92 0 00.331-1.024"/>
											</svg>
										{:else}
											<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
												<path d="M.227 8.268L0 9.429l.764 4.761C1.35 17.68 4.37 20.5 8.009 21.498l3.99 1.06 3.99-1.06c3.64-.998 6.66-3.818 7.246-7.308L24 9.43l-.227-1.161H.227z"/>
												<path d="M12 22.558l-3.99-1.06C4.37 20.5 1.35 17.68.764 14.19L0 9.43l12 3.459 12-3.459-.764 4.761c-.586 3.49-3.606 6.31-7.246 7.308L12 22.558z"/>
											</svg>
										{/if}
									</div>

									<!-- Provider info -->
									<div class="flex-1 min-w-0">
										<div class="flex items-center gap-2 flex-wrap mb-0.5">
											<h3 class="text-sm font-semibold">{provider.name}</h3>
											{#if !provider.available}
												<span class="badge badge-ghost badge-sm">Coming soon</span>
											{:else if provider.id === 'github' && isGithubConnected}
												<span class="badge badge-success badge-sm">Connected</span>
											{/if}
										</div>
										<p class="text-xs text-base-content/50">{provider.description}</p>

										{#if provider.id === 'github' && provider.available}
											{#if isGithubConnected}
												<div class="mt-3 flex flex-col gap-3">
													<div class="flex items-center gap-2 bg-base-200 rounded-lg px-3 py-2 w-fit">
														<svg class="w-3.5 h-3.5 text-success shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
															<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/>
														</svg>
														<span class="text-xs font-mono text-base-content/60">installation #{installationId}</span>
													</div>
													<div class="flex flex-wrap gap-2">
														<a href={manageUrl} target="_blank" rel="noopener" class="btn btn-xs btn-outline gap-1">
															<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
																<path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
															</svg>
															Manage on GitHub
														</a>
														<a href={installUrl} target="_blank" rel="noopener" class="btn btn-xs btn-ghost gap-1">
															Add another account
														</a>
														<button
															class="btn btn-xs btn-ghost text-error gap-1 ml-auto"
															onclick={() => showDisconnectConfirm = true}
														>
															Disconnect
														</button>
													</div>

													<!-- Disconnect confirmation -->
													{#if showDisconnectConfirm}
														<div class="mt-3 p-3 rounded-xl border border-error/30 bg-error/5 space-y-2">
															<p class="text-xs font-medium text-error">Disconnect GitHub?</p>
															<p class="text-xs text-base-content/60">
																This removes the connection in Symblon. To fully stop Symblon from accessing your GitHub,
																also <a href={manageUrl} target="_blank" rel="noopener" class="underline">uninstall the App on GitHub</a>.
															</p>
															<div class="flex gap-2 pt-1">
																<button
																	class="btn btn-xs btn-error"
																	disabled={disconnecting}
																	onclick={disconnect}
																>
																	{disconnecting ? 'Disconnecting…' : 'Yes, disconnect'}
																</button>
																<button
																	class="btn btn-xs btn-ghost"
																	onclick={() => showDisconnectConfirm = false}
																>
																	Cancel
																</button>
															</div>
														</div>
													{/if}
												</div>
											{:else}
												<div class="mt-3 space-y-3">
													<ol class="space-y-1.5">
														{#each [
															'Install the App on your GitHub account or org',
															'Select which repos Symblon can access',
															"You'll be redirected back here automatically",
														] as step, i}
															<li class="flex items-start gap-2 text-xs text-base-content/60">
																<span class="w-4 h-4 rounded-full bg-primary/10 text-primary font-bold flex items-center justify-center shrink-0 mt-0.5 text-[10px]">{i + 1}</span>
																{step}
															</li>
														{/each}
													</ol>
													<a href={installUrl} class="btn btn-sm btn-primary gap-2">
														<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
															<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
														</svg>
														Install GitHub App
													</a>
												</div>
											{/if}
										{/if}
									</div>
								</div>
							</div>
						</div>
					{/each}

				{:else if activeSection === 'profile'}
					<div class="card bg-base-100 border border-base-200">
						<div class="card-body p-6">
							<h2 class="font-semibold mb-1">Profile</h2>
							<p class="text-sm text-base-content/50">Coming soon.</p>
						</div>
					</div>

				{:else if activeSection === 'billing'}
					<div class="card bg-base-100 border border-base-200">
						<div class="card-body p-6">
							<h2 class="font-semibold mb-1">Billing</h2>
							<p class="text-sm text-base-content/50">Coming soon.</p>
						</div>
					</div>
				{/if}

			</div>
		</div>
	</main>
</div>

