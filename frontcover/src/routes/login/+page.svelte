<script lang="ts">
	import { page } from '$app/stores';

	let { data } = $props();
	let loading = $state(false);
	let error = $derived($page.url.searchParams.get('error'));

	async function loginWithGitHub() {
		loading = true;
		const { error: authError } = await data.supabase.auth.signInWithOAuth({
			provider: 'github',
			options: {
				redirectTo: `${$page.url.origin}/auth/callback`,
				scopes: 'read:user user:email'
			}
		});
		if (authError) loading = false;
	}
</script>

<div class="min-h-screen flex items-center justify-center bg-base-200">
	<div class="card w-full max-w-sm bg-base-100 shadow-xl">
		<div class="card-body items-center text-center gap-6">
			<h1 class="text-3xl font-bold tracking-tight">symblon</h1>
			<p class="text-base-content/60">Collect achievements. Build your profile.</p>

			{#if error}
				<div role="alert" class="alert alert-error w-full">
					<span>Authentication failed. Please try again.</span>
				</div>
			{/if}

			<button class="btn btn-neutral w-full gap-2" onclick={loginWithGitHub} disabled={loading}>
				{#if loading}
					<span class="loading loading-spinner loading-sm"></span>
				{:else}
					<svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
						<path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
					</svg>
				{/if}
				Continue with GitHub
			</button>
		</div>
	</div>
</div>
