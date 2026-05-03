import { redirect, fail } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import {
	listAgents,
	listInstalledAgentIDs,
	installAgent,
	uninstallAgent
} from '$lib/api/registrar';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.session) redirect(303, '/login');

	const token = locals.session.access_token;

	const [agents, installedIDs] = await Promise.all([
		listAgents(),
		listInstalledAgentIDs(token)
	]);

	const installedSet = new Set(installedIDs);

	return {
		agents: agents.map((a) => ({ ...a, installed: installedSet.has(a.id) }))
	};
};

export const actions: Actions = {
	install: async ({ request, locals }) => {
		if (!locals.session) return fail(401, { error: 'unauthorized' });
		const data = await request.formData();
		const publisher = data.get('publisher') as string;
		const handle = data.get('handle') as string;
		const version = data.get('version') as string;
		try {
			await installAgent(locals.session.access_token, publisher, handle, version);
		} catch (e) {
			return fail(500, { error: String(e) });
		}
	},

	uninstall: async ({ request, locals }) => {
		if (!locals.session) return fail(401, { error: 'unauthorized' });
		const data = await request.formData();
		const publisher = data.get('publisher') as string;
		const handle = data.get('handle') as string;
		const version = data.get('version') as string;
		try {
			await uninstallAgent(locals.session.access_token, publisher, handle, version);
		} catch (e) {
			return fail(500, { error: String(e) });
		}
	}
};
