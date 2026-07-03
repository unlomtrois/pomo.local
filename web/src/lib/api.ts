// Thin client for the pomo daemon HTTP API.
//
// Same-origin in both modes: in production the daemon serves this app; in dev
// the Vite proxy (see vite.config.ts) forwards /api and /healthz to the daemon.
// VITE_POMO_API can override to point at a remote daemon.
const BASE = import.meta.env.VITE_POMO_API ?? '';

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

/** Move a session to a new start time (duration preserved). */
export async function moveSession(id: number, start: Date): Promise<Session> {
	const res = await fetch(`${BASE}/api/sessions/${id}`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ start_time: start.toISOString() })
	});
	if (!res.ok) throw new Error(`PATCH /api/sessions/${id} failed: ${res.status}`);
	return res.json();
}

export async function activeSession(): Promise<Session | null> {
	const res = await fetch(`${BASE}/api/sessions/active`);
	if (res.status === 404) return null;
	if (!res.ok) throw new Error(`GET /api/sessions/active failed: ${res.status}`);
	return res.json();
}
