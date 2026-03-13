import type { Session, SupabaseClient, User } from '@supabase/supabase-js';

// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
	namespace App {
		interface Locals {
			supabase: SupabaseClient;
			session: Session | null;
			user: User | null;
		}
		interface PageData {
			session: Session | null;
		}
		// interface Error {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
