<script lang="ts">
	import { onMount } from 'svelte';
	import { listSessions, moveSession, type Session } from '$lib/api';
	import { startOfWeek, addDays, dayKey, isSameDay, withTimeOfDay, WEEKDAYS } from '$lib/week';

	let sessions = $state<Session[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let weekOffset = $state(0);
	let draggingId = $state<number | null>(null);
	let dragOverKey = $state<string | null>(null);

	const today = new Date();
	let weekStart = $derived(startOfWeek(addDays(today, weekOffset * 7)));
	let days = $derived(Array.from({ length: 7 }, (_, i) => addDays(weekStart, i)));

	function sessionsForDay(day: Date): Session[] {
		return sessions
			.filter((s) => isSameDay(new Date(s.start_time), day))
			.sort((a, b) => a.start_time.localeCompare(b.start_time));
	}

	function fmtRange(s: Session): string {
		const start = new Date(s.start_time);
		const mins = Math.max(1, Math.round(s.duration / 1e9 / 60));
		return `${start.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })} · ${mins}m`;
	}

	function weekLabel(): string {
		const end = addDays(weekStart, 6);
		const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric' };
		return `${weekStart.toLocaleDateString([], opts)} – ${end.toLocaleDateString([], opts)}`;
	}

	function onDragStart(e: DragEvent, s: Session) {
		draggingId = s.id;
		e.dataTransfer?.setData('text/plain', String(s.id));
		if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move';
	}

	function onDragEnd() {
		draggingId = null;
		dragOverKey = null;
	}

	async function onDrop(e: DragEvent, day: Date) {
		e.preventDefault();
		dragOverKey = null;
		const id = draggingId;
		draggingId = null;
		if (id == null) return;

		const s = sessions.find((x) => x.id === id);
		if (!s || isSameDay(new Date(s.start_time), day)) return;

		const newStart = withTimeOfDay(day, new Date(s.start_time));
		const prev = sessions;
		// optimistic update
		sessions = sessions.map((x) =>
			x.id === id ? { ...x, start_time: newStart.toISOString() } : x
		);
		try {
			const updated = await moveSession(id, newStart);
			sessions = sessions.map((x) => (x.id === id ? updated : x));
		} catch (err) {
			sessions = prev; // revert
			error = err instanceof Error ? err.message : String(err);
		}
	}

	onMount(async () => {
		try {
			sessions = await listSessions(200);
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	});
</script>

<main>
	<header>
		<h1>🍅 Calendar</h1>
		<nav class="weeknav">
			<button onclick={() => (weekOffset -= 1)} aria-label="Previous week">‹</button>
			<button onclick={() => (weekOffset = 0)}>{weekLabel()}</button>
			<button onclick={() => (weekOffset += 1)} aria-label="Next week">›</button>
		</nav>
		<a href="/" class="home">← dashboard</a>
	</header>

	{#if error}
		<p class="error">{error}</p>
	{/if}

	{#if loading}
		<p>Loading…</p>
	{:else}
		<div class="board">
			{#each days as day (dayKey(day))}
				{@const key = dayKey(day)}
				<section
					class="col"
					class:over={dragOverKey === key}
					class:istoday={isSameDay(day, today)}
					role="list"
					ondragover={(e) => {
						e.preventDefault();
						dragOverKey = key;
					}}
					ondragleave={() => {
						if (dragOverKey === key) dragOverKey = null;
					}}
					ondrop={(e) => onDrop(e, day)}
				>
					<div class="colhead">
						<span class="dow">{WEEKDAYS[(day.getDay() + 6) % 7]}</span>
						<span class="date">{day.getDate()}</span>
					</div>

					{#each sessionsForDay(day) as s (s.id)}
						<article
							class="card"
							class:dragging={draggingId === s.id}
							class:active={s.status === 'active'}
							class:cancelled={s.status === 'cancelled'}
							draggable="true"
							role="listitem"
							ondragstart={(e) => onDragStart(e, s)}
							ondragend={onDragEnd}
						>
							<div class="topic">{s.topic}</div>
							<div class="meta">{fmtRange(s)}</div>
						</article>
					{/each}
				</section>
			{/each}
		</div>
		<p class="hint">Drag a session onto another day to reschedule it (keeps its time of day).</p>
	{/if}
</main>

<style>
	main {
		max-width: 72rem;
		margin: 1.5rem auto;
		padding: 0 1rem;
		font-family: system-ui, sans-serif;
	}
	header {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
	}
	h1 {
		font-size: 1.6rem;
		margin: 0;
	}
	.weeknav {
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}
	.weeknav button {
		border: 1px solid #ccc;
		background: #fff;
		border-radius: 6px;
		padding: 0.3rem 0.7rem;
		cursor: pointer;
		font-size: 0.95rem;
	}
	.weeknav button:hover {
		background: #f3f3f3;
	}
	.home {
		margin-left: auto;
		color: #555;
		text-decoration: none;
		font-size: 0.9rem;
	}
	.board {
		display: grid;
		grid-template-columns: repeat(7, minmax(9rem, 1fr));
		gap: 0.5rem;
		margin-top: 1rem;
		overflow-x: auto;
	}
	.col {
		background: #fafafa;
		border: 1px solid #eee;
		border-radius: 8px;
		min-height: 16rem;
		padding: 0.4rem;
		transition: background 0.12s, box-shadow 0.12s;
	}
	.col.over {
		background: #eef6ff;
		box-shadow: inset 0 0 0 2px #7ab7ff;
	}
	.col.istoday .colhead {
		color: #b00020;
		font-weight: 700;
	}
	.colhead {
		display: flex;
		justify-content: space-between;
		font-size: 0.8rem;
		color: #666;
		padding: 0.2rem 0.3rem 0.5rem;
	}
	.card {
		background: #fff;
		border: 1px solid #e2e2e2;
		border-left: 3px solid #e03a3a;
		border-radius: 6px;
		padding: 0.4rem 0.5rem;
		margin-bottom: 0.4rem;
		cursor: grab;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
	}
	.card:active {
		cursor: grabbing;
	}
	.card.dragging {
		opacity: 0.4;
	}
	.card.active {
		border-left-color: #1a9e4b;
	}
	.card.cancelled {
		border-left-color: #aaa;
		opacity: 0.7;
	}
	.topic {
		font-size: 0.85rem;
		font-weight: 600;
		word-break: break-word;
	}
	.meta {
		font-size: 0.75rem;
		color: #777;
		margin-top: 0.15rem;
	}
	.hint {
		color: #888;
		font-size: 0.8rem;
		margin-top: 0.75rem;
	}
	.error {
		color: #b00020;
	}
</style>
