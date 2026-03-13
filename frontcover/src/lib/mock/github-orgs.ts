import type { GitHubOrg } from '$lib/types/account';

export const mockGitHubOrgs: GitHubOrg[] = [
	{
		id: 1001,
		login: 'acme-corp',
		name: 'Acme Corp',
		avatarUrl: 'https://avatars.githubusercontent.com/u/1342004',
		role: 'owner'
	},
	{
		id: 1002,
		login: 'devteam-xyz',
		name: 'DevTeam XYZ',
		avatarUrl: 'https://avatars.githubusercontent.com/u/6128107',
		role: 'admin'
	},
	{
		id: 1003,
		login: 'open-source-collective',
		name: 'Open Source Collective',
		avatarUrl: 'https://avatars.githubusercontent.com/u/9919',
		role: 'member'
	}
];
