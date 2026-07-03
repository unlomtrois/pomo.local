// Thin client for the pomo daemon HTTP API.
//
// In production the Svelte app is served by the daemon itself, so the API is
// same-origin (base = ''). During `vite dev` the app runs on :5173 while the
// daemon listens on :7420, so target it directly. VITE_POMO_API overrides both.
const BASE =
	import.meta.env.VITE_POMO_API ?? (import.meta.env.DEV ? 'http://127.0.0.1:7420' : '');

export interface Session {
	id: number;
	topic: string;
	start_time: string;
	stop_time: string;
	duration: number; // nanoseconds
	status: 'active' | 'done' | 'cancelled';
	source: 'cli' | 'web' | 'mcp';
}

export async function listSessions(limit = 20): Promise<Session[]> {
	const res = await fetch(`${BASE}/api/sessions?limit=${limit}`);
	if (!res.ok) throw new Error(`GET /api/sessions failed: ${res.status}`);
	return res.json();
}

export async function activeSession(): Promise<Session | null> {
	const res = await fetch(`${BASE}/api/sessions/active`);
	if (res.status === 404) return null;
	if (!res.ok) throw new Error(`GET /api/sessions/active failed: ${res.status}`);
	return res.json();
}
