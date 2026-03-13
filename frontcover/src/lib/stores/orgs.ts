import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import type { OrgAccount } from '$lib/types/account';

const STORAGE_KEY = 'symblon_orgs';

function createOrgsStore() {
	const initial: OrgAccount[] = browser
		? JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]')
		: [];

	const { subscribe, update } = writable<OrgAccount[]>(initial);

	function persist(orgs: OrgAccount[]) {
		if (browser) localStorage.setItem(STORAGE_KEY, JSON.stringify(orgs));
		return orgs;
	}

	return {
		subscribe,
		addOrg(org: OrgAccount) {
			update((orgs) => persist([...orgs.filter((o) => o.slug !== org.slug), org]));
		},
		removeOrg(slug: string) {
			update((orgs) => persist(orgs.filter((o) => o.slug !== slug)));
		}
	};
}

export const orgsStore = createOrgsStore();
