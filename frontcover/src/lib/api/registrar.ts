import { REGISTRAR_URL } from '$env/static/private';

export interface AgentSymbol {
	id: string;
	symbol_id: string;
	name: string;
	description: string;
	type: 'realtime' | 'temporal';
	window_hours: number;
}

export interface Agent {
	id: string;
	publisher: string;
	handle: string;
	version: string;
	name: string;
	description: string;
	visibility: string;
	pricing_model: 'free' | 'paid';
	price_usd?: number;
	published_at: string;
	symbols: AgentSymbol[];
}

function headers(token?: string): HeadersInit {
	const h: HeadersInit = { 'Content-Type': 'application/json' };
	if (token) h['Authorization'] = `Bearer ${token}`;
	return h;
}

export async function listAgents(): Promise<Agent[]> {
	const res = await fetch(`${REGISTRAR_URL}/registrar/v1/agents`);
	if (!res.ok) throw new Error(`listAgents: ${res.status}`);
	return res.json();
}

export async function listInstalledAgents(token: string): Promise<Agent[]> {
	const res = await fetch(`${REGISTRAR_URL}/registrar/v1/me/agents`, {
		headers: headers(token)
	});
	if (!res.ok) throw new Error(`listInstalledAgents: ${res.status}`);
	const data = await res.json();
	return (data ?? []).map((a: Agent) => ({ ...a, symbols: a.symbols ?? [] }));
}

export async function installAgent(
	token: string,
	publisher: string,
	handle: string,
	version: string
): Promise<void> {
	const res = await fetch(`${REGISTRAR_URL}/registrar/v1/agents/${publisher}/${handle}/${version}/install`, {
		method: 'POST',
		headers: headers(token)
	});
	if (!res.ok) throw new Error(`installAgent: ${res.status}`);
}

export async function uninstallAgent(
	token: string,
	publisher: string,
	handle: string,
	version: string
): Promise<void> {
	const res = await fetch(`${REGISTRAR_URL}/registrar/v1/agents/${publisher}/${handle}/${version}/install`, {
		method: 'DELETE',
		headers: headers(token)
	});
	if (!res.ok && res.status !== 204) throw new Error(`uninstallAgent: ${res.status}`);
}
