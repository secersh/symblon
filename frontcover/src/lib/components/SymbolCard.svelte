<script lang="ts">
	import type { EarnedSymbol, SymbolTheme } from '$lib/types/symbol';

	let { symbol, theme = 'scouts' }: { symbol: EarnedSymbol; theme?: SymbolTheme } = $props();

	const themeStyles: Record<SymbolTheme, { bg: string; border: string; badge: string; label: string }> = {
		scouts:  { bg: 'bg-amber-50  dark:bg-amber-950/30',   border: 'border-amber-200  dark:border-amber-800',   badge: 'bg-amber-100  text-amber-700  dark:bg-amber-900/50  dark:text-amber-300',   label: 'Badge'  },
		wizards: { bg: 'bg-violet-50  dark:bg-violet-950/30', border: 'border-violet-200 dark:border-violet-800',  badge: 'bg-violet-100 text-violet-700 dark:bg-violet-900/50 dark:text-violet-300', label: 'Spell'  },
		hackers: { bg: 'bg-emerald-50 dark:bg-emerald-950/30', border: 'border-emerald-200 dark:border-emerald-800', badge: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/50 dark:text-emerald-300', label: 'Hack' },
	};

	let style = $derived(themeStyles[theme]);

	let earnedDate = $derived(
		new Intl.DateTimeFormat('en', { month: 'short', day: 'numeric', year: 'numeric' })
			.format(new Date(symbol.earnedAt))
	);
</script>

<div class="relative rounded-2xl border p-4 flex flex-col gap-3 transition-shadow hover:shadow-md {style.bg} {style.border}">
	<!-- Theme + kind badges (top row) -->
	<div class="flex items-center justify-between gap-2">
		<span class="text-[10px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded-full {style.badge}">
			{style.label}
		</span>
		<span class="text-[10px] text-base-content/40 font-medium uppercase tracking-wider">
			{symbol.kind === 'temporal' ? '⏱ temporal' : '⚡ real-time'}
		</span>
	</div>

	<!-- Emoji + name -->
	<div class="flex items-start gap-3">
		<span class="text-3xl leading-none mt-0.5 select-none">{symbol.emoji}</span>
		<div class="min-w-0">
			<div class="flex items-center gap-1.5 flex-wrap">
				<h3 class="text-sm font-bold leading-tight">{symbol.name}</h3>
				{#if (symbol.multiplier ?? 1) > 1}
					<span class="text-[10px] font-bold text-primary bg-primary/10 px-1.5 py-0.5 rounded-full">
						×{symbol.multiplier}
					</span>
				{/if}
			</div>
			<p class="text-xs text-base-content/60 mt-0.5 leading-snug">{symbol.description}</p>
		</div>
	</div>

	<!-- Footer: org/repo + date -->
	<div class="flex items-end justify-between gap-2 pt-1 border-t border-black/5 dark:border-white/5">
		<div class="flex flex-col gap-0.5 min-w-0">
			<span class="text-[11px] text-base-content/50 truncate font-mono">
				{symbol.githubOrg}/{symbol.githubRepo}
			</span>
		</div>
		<div class="flex items-center gap-1.5 shrink-0">
			{#if symbol.earnCount > 1}
				<span class="text-[10px] text-base-content/40">×{symbol.earnCount} earned</span>
			{/if}
			<span class="text-[11px] text-base-content/40">{earnedDate}</span>
		</div>
	</div>
</div>
