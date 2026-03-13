export type OrgRole = 'owner' | 'admin' | 'member';
export type OrgPlan = 'free' | 'pro' | 'enterprise';

export interface OrgAccount {
	slug: string;
	name: string;
	avatarColor: string;
	role: OrgRole;
	memberCount: number;
	plan: OrgPlan;
	linkedGithubOrg?: string;
	createdAt: string;
}

export interface GitHubOrg {
	id: number;
	login: string;
	name: string;
	avatarUrl: string;
	role: OrgRole;
}
