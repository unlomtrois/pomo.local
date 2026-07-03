<script lang="ts">
	import { onMount } from 'svelte';
	import { listSessions, moveSession, type Session } from '$lib/api';
	import { startOfWeek, addDays, dayKey, isSameDay, withTimeOfDay, WEEKDAYS } from '$lib/week';

	const HH = 48; // pixels per hour (must drive both layout and card positioning)
	const HOURS = Array.from({ length: 24 }, (_, i) => i);

	let sessions = $state<Session[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let weekOffset = $state(0);
	let draggingId = $state<number | null>(null);
	let dragOverKey = $state<string | null>(null);
	let scroller: HTMLDivElement;

	const today = new Date();
	let weekStart = $derived(startOfWeek(addDays(today, weekOffset * 7)));
	let days = $derived(Array.from({ length: 7 }, (_, i) => addDays(weekStart, i)));

	function sessionsForDay(day: Date): Session[] {
		return sessions.filter((s) => isSameDay(new Date(s.start_time), day));
	}

	function topPx(s: Session): number {
		const d = new Date(s.start_time);
		return ((d.getHours() * 60 + d.getMinutes()) / 60) * HH;
	}
	function heightPx(s: Session): number {
		return Math.max(16, (s.duration / 1e9 / 3600) * HH);
	}
	function fmtTime(s: Session): string {
		return new Date(s.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
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
		sessions = sessions.map((x) =>
			x.id === id ? { ...x, start_time: newStart.toISOString() } : x
		);
		try {
			const updated = await moveSession(id, newStart);
			sessions = sessions.map((x) => (x.id === id ? updated : x));
		} catch (err) {
			sessions = prev;
			error = err instanceof Error ? err.message : String(err);
		}
	}

	onMount(async () => {
		try {
			sessions = await listSessions(500);
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
		if (scroller) scroller.scrollTop = 7 * HH; // open around the working day
	});
</script>

<main class="cal">
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

	<div class="daysrow">
		<div class="corner"></div>
		{#each days as day (dayKey(day))}
			<div class="dayhead" class:istoday={isSameDay(day, today)}>
				<span class="dow">{WEEKDAYS[(day.getDay() + 6) % 7]}</span>
				<span class="date">{day.getDate()}</span>
			</div>
		{/each}
	</div>

	<div class="scroll" bind:this={scroller}>
		<div class="body" style="--hh:{HH}px; height:{HH * 24}px">
			<div class="gutter">
				{#each HOURS as h (h)}
					<div class="hourlabel" style="top:{h * HH}px">
						{String(h).padStart(2, '0')}:00
					</div>
				{/each}
			</div>

			{#each days as day (dayKey(day))}
				{@const key = dayKey(day)}
				<section
					class="col"
					class:over={dragOverKey === key}
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
					{#each sessionsForDay(day) as s (s.id)}
						<article
							class="card"
							class:dragging={draggingId === s.id}
							class:active={s.status === 'active'}
							class:cancelled={s.status === 'cancelled'}
							style="top:{topPx(s)}px; height:{heightPx(s)}px"
							draggable="true"
							role="listitem"
							ondragstart={(e) => onDragStart(e, s)}
							ondragend={onDragEnd}
						>
							<span class="topic">{s.topic}</span>
							<span class="tt">{fmtTime(s)}</span>
						</article>
					{/each}
				</section>
			{/each}
		</div>
	</div>
</main>

<style>
	.cal {
		height: 100vh;
		display: flex;
		flex-direction: column;
		font-family: system-ui, sans-serif;
		color: #222;
	}
	header {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
		padding: 0.75rem 1rem;
		flex: none;
	}
	h1 {
		font-size: 1.4rem;
		margin: 0;
	}
	.weeknav {
		display: flex;
		gap: 0.25rem;
	}
	.weeknav button {
		border: 1px solid #ccc;
		background: #fff;
		border-radius: 6px;
		padding: 0.3rem 0.7rem;
		cursor: pointer;
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
	.error {
		color: #b00020;
		padding: 0 1rem;
		margin: 0 0 0.5rem;
	}

	/* Column template shared by the day header row and the scrolling body so
	   they stay aligned. First track is the time gutter. */
	.daysrow,
	.body {
		display: grid;
		grid-template-columns: 3.5rem repeat(7, 1fr);
	}
	.daysrow {
		flex: none;
		border-bottom: 1px solid #ddd;
	}
	.corner {
		border-right: 1px solid #eee;
	}
	.dayhead {
		display: flex;
		align-items: baseline;
		justify-content: center;
		gap: 0.35rem;
		padding: 0.4rem 0;
		font-size: 0.85rem;
		color: #666;
		border-left: 1px solid #eee;
	}
	.dayhead .date {
		font-weight: 600;
		color: #333;
	}
	.dayhead.istoday {
		color: #b00020;
	}
	.dayhead.istoday .date {
		color: #b00020;
	}

	.scroll {
		flex: 1;
		overflow-y: auto;
	}
	.body {
		position: relative;
	}

	.gutter {
		position: relative;
		border-right: 1px solid #eee;
	}
	.hourlabel {
		position: absolute;
		right: 0.35rem;
		transform: translateY(-0.5em);
		font-size: 0.7rem;
		color: #999;
	}

	.col {
		position: relative;
		border-left: 1px solid #eee;
		/* hour guide lines */
		background-image: linear-gradient(to bottom, #ececec 1px, transparent 1px);
		background-size: 100% var(--hh);
	}
	.col.over {
		background-color: #eef6ff;
	}

	.card {
		position: absolute;
		left: 2px;
		right: 2px;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		background: #fff;
		border: 1px solid #e2e2e2;
		border-left: 3px solid #e03a3a;
		border-radius: 5px;
		padding: 1px 5px;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
		cursor: grab;
		font-size: 0.75rem;
		line-height: 1.15;
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
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.tt {
		color: #888;
		font-size: 0.68rem;
	}
</style>
