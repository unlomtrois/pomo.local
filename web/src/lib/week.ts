// Date helpers for the weekly calendar. All local-time based.

/** Monday 00:00 local of the week containing `d`. */
export function startOfWeek(d: Date): Date {
	const r = new Date(d);
	r.setHours(0, 0, 0, 0);
	const day = (r.getDay() + 6) % 7; // Mon=0 … Sun=6
	r.setDate(r.getDate() - day);
	return r;
}

export function addDays(d: Date, n: number): Date {
	const r = new Date(d);
	r.setDate(r.getDate() + n);
	return r;
}

/** Stable per-day key (local), e.g. "2026-07-03". */
export function dayKey(d: Date): string {
	return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

export function isSameDay(a: Date, b: Date): boolean {
	return dayKey(a) === dayKey(b);
}

/** Combine a target day with a source time-of-day into a new Date. */
export function withTimeOfDay(day: Date, timeSource: Date): Date {
	const r = new Date(day);
	r.setHours(timeSource.getHours(), timeSource.getMinutes(), timeSource.getSeconds(), 0);
	return r;
}

export const WEEKDAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
