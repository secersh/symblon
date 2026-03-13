<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let { data } = $props();

	let status = $state<'connecting' | 'success' | 'error'>('connecting');
	let errorMsg = $state('');

	onMount(async () => {
		const installationId = $page.url.searchParams.get('installation_id');
		const setupAction = $page.url.searchParams.get('setup_action');

		if (!installationId) {
			status = 'error';
			errorMsg = 'No installation ID received from GitHub.';
			return;
		}

		// Persist installation_id to Supabase user metadata
		const { error } = await data.supabase.auth.updateUser({
			data: {
				github_installation_id: installationId,
				github_setup_action: setupAction ?? 'install'
			}
		});

		if (error) {
			status = 'error';
			errorMsg = error.message;
			return;
		}

		status = 'success';
		// Small delay so the user sees the success state before redirect
		setTimeout(() => goto('/settings?github=connected'), 1200);
	});
</script>

<div class="min-h-full bg-base-200 flex items-center justify-center">
	<div class="card bg-base-100 border border-base-200 w-full max-w-sm p-8 text-center">
		{#if status === 'connecting'}
			<span class="loading loading-spinner loading-lg text-primary mx-auto mb-4"></span>
			<h2 class="text-base font-semibold">Connecting GitHub App…</h2>
			<p class="text-sm text-base-content/50 mt-1">Saving your installation</p>
		{:else if status === 'success'}
			<div class="w-14 h-14 rounded-full bg-success/10 flex items-center justify-center mx-auto mb-4">
				<svg class="w-7 h-7 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/>
				</svg>
			</div>
			<h2 class="text-base font-semibold">GitHub App connected!</h2>
			<p class="text-sm text-base-content/50 mt-1">Redirecting to settings…</p>
		{:else}
			<div class="w-14 h-14 rounded-full bg-error/10 flex items-center justify-center mx-auto mb-4">
				<svg class="w-7 h-7 text-error" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
				</svg>
			</div>
			<h2 class="text-base font-semibold">Connection failed</h2>
			<p class="text-sm text-error mt-1">{errorMsg}</p>
			<a href="/settings" class="btn btn-sm btn-ghost mt-4">Back to settings</a>
		{/if}
	</div>
</div>
