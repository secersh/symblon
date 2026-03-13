export type SymbolTheme = 'scouts' | 'wizards' | 'hackers';
export type SymbolKind = 'realtime' | 'temporal';

export interface EarnedSymbol {
	id: string;
	name: string;
	description: string;
	emoji: string;
	kind: SymbolKind;
	/** The Symblon org or user that defined this symbol */
	issuedBy: string;
	/** GitHub org where the qualifying activity happened */
	githubOrg: string;
	/** GitHub repo where the qualifying activity happened */
	githubRepo: string;
	earnedAt: string;
	/** Optional multiplier, e.g. ×3 streak */
	multiplier?: number;
	/** How many times this has been earned */
	earnCount: number;
}
