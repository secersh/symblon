<script lang="ts">
	import type { User } from '@supabase/supabase-js';
	import { page } from '$app/stores';
	import { orgsStore } from '$lib/stores/orgs';
	import { orgInitials } from '$lib/utils/avatar';

	let { user }: { user: User | null } = $props();

	let avatarUrl = $derived(user?.user_metadata?.avatar_url as string | undefined);
	let username = $derived(
		(user?.user_metadata?.user_name as string | undefined) ??
		(user?.user_metadata?.full_name as string | undefined) ??
		user?.email
	);
	let initials = $derived(username?.slice(0, 2).toUpperCase() ?? '?');

	// Detect current context from URL
	let currentOrgSlug = $derived.by(() => {
		const segments = $page.url.pathname.split('/').filter(Boolean);
		if (segments.length > 0) {
			const reserved = new Set(['profile', 'login', 'auth', 'orgs', 'settings']);
			const first = segments[0];
			if (!reserved.has(first) && $orgsStore.some((o) => o.slug === first)) {
				return first;
			}
		}
		return null;
	});

	let currentOrg = $derived($orgsStore.find((o) => o.slug === currentOrgSlug) ?? null);
	let isPersonal = $derived(!currentOrg);
</script>

<div class="navbar bg-base-100 border-b border-base-200 px-4 lg:px-8 shrink-0">
	<!-- Left: brand -->
	<div class="flex-1">
		<a href="/" class="flex items-center gap-2">
			<svg viewBox="0 0 128 128" class="w-7 h-7" aria-hidden="true">
				<rect width="128" height="128" rx="28" fill="#0f0f0f"/>
				<path d="M64 18 C64 18 68 50 82 64 C68 78 64 110 64 110 C64 110 60 78 46 64 C60 50 64 18 64 18 Z" fill="white"/>
				<path d="M18 64 C18 64 50 60 64 46 C78 60 110 64 110 64 C110 64 78 68 64 82 C50 68 18 64 18 64 Z" fill="white"/>
			</svg>
			<span class="text-lg font-bold tracking-tight">symblon</span>
		</a>
	</div>

	<!-- Right: account switcher -->
	{#if user}
		<div class="flex-none">
			<div class="dropdown dropdown-end">
				<!-- Trigger: shows current context -->
				<div
					tabindex="0"
					role="button"
					class="flex items-center gap-2.5 px-3 py-2 rounded-xl hover:bg-base-200 transition-colors cursor-pointer"
				>
					{#if currentOrg}
						<!-- Org context trigger -->
						<div
							class="w-8 h-8 rounded-lg flex items-center justify-center text-white text-xs font-bold shrink-0"
							style="background-color: {currentOrg.avatarColor}"
						>
							{orgInitials(currentOrg.name)}
						</div>
						<span class="text-sm font-medium hidden sm:block max-w-[120px] truncate">{currentOrg.name}</span>
					{:else}
						<!-- Personal context trigger -->
						<div class="avatar shrink-0">
							<div class="w-8 h-8 rounded-full ring ring-base-300 ring-offset-base-100 ring-offset-1">
								{#if avatarUrl}
									<img src={avatarUrl} alt={username} referrerpolicy="no-referrer" />
								{:else}
									<div class="bg-neutral text-neutral-content flex items-center justify-center w-full h-full text-xs font-bold">
										{initials}
									</div>
								{/if}
							</div>
						</div>
						<span class="text-sm font-medium hidden sm:block max-w-[120px] truncate">{username}</span>
					{/if}
					<svg class="w-3.5 h-3.5 text-base-content/40 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7"/>
					</svg>
				</div>

				<!-- Dropdown -->
				<div class="dropdown-content bg-base-100 rounded-xl border border-base-200 shadow-lg w-64 max-w-[calc(100vw-2rem)] p-2 mt-1 z-50">

					<!-- Section: Switch account -->
					<p class="text-[10px] font-semibold uppercase tracking-widest text-base-content/40 px-3 pt-1 pb-1.5">Switch account</p>

					<!-- Personal account -->
					<a href="/" class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-base-200 transition-colors group">
						<div class="avatar shrink-0">
							<div class="w-8 h-8 rounded-full">
								{#if avatarUrl}
									<img src={avatarUrl} alt={username} referrerpolicy="no-referrer" />
								{:else}
									<div class="bg-neutral text-neutral-content flex items-center justify-center w-full h-full text-xs font-bold">
										{initials}
									</div>
								{/if}
							</div>
						</div>
						<div class="min-w-0 flex-1">
							<p class="text-sm font-medium truncate">{username}</p>
							<p class="text-xs text-base-content/50">Personal</p>
						</div>
						{#if isPersonal}
							<svg class="w-4 h-4 text-primary shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
								<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/>
							</svg>
						{/if}
					</a>

					<!-- Org accounts -->
					{#each $orgsStore as org}
						<a href="/{org.slug}" class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-base-200 transition-colors">
							<div
								class="w-8 h-8 rounded-lg flex items-center justify-center text-white text-xs font-bold shrink-0"
								style="background-color: {org.avatarColor}"
							>
								{orgInitials(org.name)}
							</div>
							<div class="min-w-0 flex-1">
								<p class="text-sm font-medium truncate">{org.name}</p>
								<p class="text-xs text-base-content/50 capitalize">{org.role} · {org.plan}</p>
							</div>
							{#if currentOrg?.slug === org.slug}
								<svg class="w-4 h-4 text-primary shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
									<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/>
								</svg>
							{/if}
						</a>
					{/each}

					<!-- New org -->
					<a href="/orgs/new" class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-base-200 transition-colors text-primary">
						<div class="w-8 h-8 rounded-lg border-2 border-dashed border-primary/40 flex items-center justify-center shrink-0">
							<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
							</svg>
						</div>
						<span class="text-sm font-medium">New organization</span>
					</a>

					<div class="border-t border-base-200 my-2 mx-1"></div>

					<!-- Section: Account actions -->
					<a href="/profile" class="flex items-center gap-2.5 text-sm px-3 py-2 rounded-lg hover:bg-base-200 transition-colors">
						<svg class="w-4 h-4 text-base-content/50 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
						</svg>
						Profile
					</a>

					<a href="/agents" class="flex items-center gap-2.5 text-sm px-3 py-2 rounded-lg hover:bg-base-200 transition-colors">
						<svg class="w-4 h-4 text-base-content/50 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"/>
						</svg>
						Agents
					</a>

					<a href="/settings" class="flex items-center gap-2.5 text-sm px-3 py-2 rounded-lg hover:bg-base-200 transition-colors">
						<svg class="w-4 h-4 text-base-content/50 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
						</svg>
						Settings
					</a>

					<div class="border-t border-base-200 my-2 mx-1"></div>

					<form method="POST" action="/auth/logout">
						<button type="submit" class="flex items-center gap-2.5 text-sm text-error w-full px-3 py-2 rounded-lg hover:bg-base-200 transition-colors">
							<svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
							</svg>
							Sign out
						</button>
					</form>
				</div>
			</div>
		</div>
	{/if}
</div>
