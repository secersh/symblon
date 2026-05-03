import { redirect, fail } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import {
	listAgents,
	listInstalledAgents,
	installAgent,
	uninstallAgent
} from '$lib/api/registrar';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.session) redirect(303, '/login');

	const token = locals.session.access_token;

	const [catalog, installed] = await Promise.all([
		listAgents(),
		listInstalledAgents(token)
	]);

	// Map publisher/handle → installed agent (version may differ from catalog)
	const installedMap = new Map(installed.map((a) => [`${a.publisher}/${a.handle}`, a]));

	// Merge: every catalog agent gets installed status + installed version if applicable
	const agents = catalog.map((a) => {
		const inst = installedMap.get(`${a.publisher}/${a.handle}`);
		return {
			...a,
			installed: !!inst,
			owned: !!inst,
			installedVersion: inst?.version ?? null
		};
	});

	// Also surface agents that are installed but no longer in the catalog
	for (const inst of installed) {
		const key = `${inst.publisher}/${inst.handle}`;
		if (!catalog.some((a) => `${a.publisher}/${a.handle}` === key)) {
			agents.push({ ...inst, installed: true, owned: true, installedVersion: inst.version });
		}
	}

	return { agents };
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
	},

	update: async ({ request, locals }) => {
		if (!locals.session) return fail(401, { error: 'unauthorized' });
		const data = await request.formData();
		const publisher = data.get('publisher') as string;
		const handle = data.get('handle') as string;
		const oldVersion = data.get('old_version') as string;
		const newVersion = data.get('new_version') as string;
		try {
			await uninstallAgent(locals.session.access_token, publisher, handle, oldVersion);
			await installAgent(locals.session.access_token, publisher, handle, newVersion);
		} catch (e) {
			return fail(500, { error: String(e) });
		}
	}
};
