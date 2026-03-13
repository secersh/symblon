import type { EarnedSymbol } from '$lib/types/symbol';

export const mockSymbols: EarnedSymbol[] = [
	{
		id: '1',
		name: 'First Merge',
		description: 'Merged your first pull request.',
		emoji: '🔀',
		kind: 'realtime',
		issuedBy: 'symblon',
		githubOrg: 'secersh',
		githubRepo: 'symblon',
		earnedAt: '2026-01-04T10:00:00Z',
		earnCount: 1
	},
	{
		id: '2',
		name: 'Bug Squasher',
		description: 'Closed 5 bug-labelled issues within 48 hours.',
		emoji: '🐛',
		kind: 'temporal',
		issuedBy: 'symblon',
		githubOrg: 'secersh',
		githubRepo: 'symblon',
		earnedAt: '2026-01-12T14:30:00Z',
		multiplier: 2,
		earnCount: 2
	},
	{
		id: '3',
		name: 'Code Reviewer',
		description: 'Reviewed 10 pull requests in a single week.',
		emoji: '👁️',
		kind: 'temporal',
		issuedBy: 'acme-corp',
		githubOrg: 'acme-corp',
		githubRepo: 'acme-platform',
		earnedAt: '2026-01-18T09:15:00Z',
		earnCount: 1
	},
	{
		id: '4',
		name: 'Streak Master',
		description: 'Committed code 7 days in a row.',
		emoji: '🔥',
		kind: 'temporal',
		issuedBy: 'symblon',
		githubOrg: 'secersh',
		githubRepo: 'symblon',
		earnedAt: '2026-01-25T20:00:00Z',
		multiplier: 3,
		earnCount: 3
	},
	{
		id: '5',
		name: 'Documenter',
		description: 'Opened a PR that only modifies documentation.',
		emoji: '📝',
		kind: 'realtime',
		issuedBy: 'acme-corp',
		githubOrg: 'acme-corp',
		githubRepo: 'acme-docs',
		earnedAt: '2026-02-02T11:00:00Z',
		earnCount: 1
	},
	{
		id: '6',
		name: 'Night Owl',
		description: 'Merged a PR between midnight and 5am.',
		emoji: '🦉',
		kind: 'realtime',
		issuedBy: 'symblon',
		githubOrg: 'secersh',
		githubRepo: 'frontcover',
		earnedAt: '2026-02-08T02:47:00Z',
		earnCount: 1
	},
	{
		id: '7',
		name: 'Issue Magnet',
		description: 'Opened 20 issues that were subsequently closed as fixed.',
		emoji: '🧲',
		kind: 'temporal',
		issuedBy: 'acme-corp',
		githubOrg: 'acme-corp',
		githubRepo: 'acme-platform',
		earnedAt: '2026-02-14T16:20:00Z',
		earnCount: 1
	},
	{
		id: '8',
		name: 'Ship It',
		description: 'Deployed to production 5 times in a week.',
		emoji: '🚀',
		kind: 'temporal',
		issuedBy: 'acme-corp',
		githubOrg: 'acme-corp',
		githubRepo: 'acme-platform',
		earnedAt: '2026-02-21T13:00:00Z',
		multiplier: 2,
		earnCount: 2
	},
	{
		id: '9',
		name: 'Polyglot',
		description: 'Contributed to repos in 3 different programming languages.',
		emoji: '🌐',
		kind: 'temporal',
		issuedBy: 'symblon',
		githubOrg: 'secersh',
		githubRepo: 'symblon',
		earnedAt: '2026-03-01T08:00:00Z',
		earnCount: 1
	},
	{
		id: '10',
		name: 'Rubber Duck',
		description: 'Left a comment that helped close a stale issue.',
		emoji: '🦆',
		kind: 'realtime',
		issuedBy: 'acme-corp',
		githubOrg: 'acme-corp',
		githubRepo: 'acme-docs',
		earnedAt: '2026-03-05T10:30:00Z',
		earnCount: 1
	}
];

/** Unique orgs present in the symbol list */
export function getOrgs(symbols: EarnedSymbol[]): string[] {
	return [...new Set(symbols.map((s) => s.githubOrg))];
}

/** Group symbols by githubOrg */
export function groupByOrg(symbols: EarnedSymbol[]): Record<string, EarnedSymbol[]> {
	return symbols.reduce(
		(acc, s) => {
			(acc[s.githubOrg] ??= []).push(s);
			return acc;
		},
		{} as Record<string, EarnedSymbol[]>
	);
}

/** Group symbols by githubOrg/githubRepo */
export function groupByRepo(symbols: EarnedSymbol[]): Record<string, EarnedSymbol[]> {
	return symbols.reduce(
		(acc, s) => {
			const key = `${s.githubOrg}/${s.githubRepo}`;
			(acc[key] ??= []).push(s);
			return acc;
		},
		{} as Record<string, EarnedSymbol[]>
	);
}
