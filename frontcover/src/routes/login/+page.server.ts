import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	// Already logged in — go to dashboard
	if (locals.session) redirect(303, '/');
	return {};
};
