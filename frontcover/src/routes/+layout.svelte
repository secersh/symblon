<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import Navbar from '$lib/components/Navbar.svelte';
	import { invalidate } from '$app/navigation';
	import { onMount } from 'svelte';

	let { data, children } = $props();
	let { supabase, session } = $derived(data);

	onMount(() => {
		const { data: authData } = supabase.auth.onAuthStateChange((_, newSession) => {
			if (newSession?.expires_at !== session?.expires_at) {
				invalidate('supabase:auth');
			}
		});
		return () => authData.subscription.unsubscribe();
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{#if data.session}
	<div class="flex flex-col h-dvh overflow-hidden">
		<Navbar user={data.user} />
		<div class="flex-1 overflow-y-auto">
			{@render children()}
		</div>
	</div>
{:else}
	{@render children()}
{/if}
