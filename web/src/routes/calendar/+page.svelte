<script lang="ts">
	import { onMount } from 'svelte';
	import { listSessions, moveSession, type Session } from '$lib/api';
	import { startOfWeek, addDays, dayKey, isSameDay, withTimeOfDay, WEEKDAYS } from '$lib/week';

	const HH = 48; // pixels per hour (must drive both layout and card positioning)
	const HOURS = Array.from({ length: 24 }, (_, i) => i);

	let sessions = $state<Session[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let draggingId = $state<number | null>(null);
	let dragOverKey = $state<string | null>(null);
	let scroller: HTMLDivElement;

	// A single cursor date drives both views: the week containing it (wide) or
	// the day itself (narrow). `narrow` tracks a mobile-width media query.
	let cursor = $state(new Date());
	let narrow = $state(false);
	const today = new Date();

	let weekStart = $derived(startOfWeek(cursor));
	let weekDays = $derived(Array.from({ length: 7 }, (_, i) => addDays(weekStart, i)));
	let selectedDay = $derived.by(() => {
		const d = new Date(cursor);
		d.setHours(0, 0, 0, 0);
		return d;
	});
	// Columns actually rendered: one day on mobile, the whole week otherwise.
	let visibleDays = $derived(narrow ? [selectedDay] : weekDays);
	let gridCols = $derived(`3.5rem repeat(${visibleDays.length}, 1fr)`);

	function shift(n: number) {
		cursor = addDays(cursor, narrow ? n : n * 7);
	}
	function goToday() {
		cursor = new Date();
	}

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
	function showTime(s: Session): boolean {
		return heightPx(s) >= 34;
	}
	function fmtTime(s: Session): string {
		return new Date(s.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}
	function weekLabel(): string {
		const end = addDays(weekStart, 6);
		const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric' };
		return `${weekStart.toLocaleDateString([], opts)} – ${end.toLocaleDateString([], opts)}`;
	}
	function dayLabel(): string {
		return cursor.toLocaleDateString([], { weekday: 'long', month: 'short', day: 'numeric' });
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

	onMount(() => {
		const mq = window.matchMedia('(max-width: 640px)');
		narrow = mq.matches;
		const onChange = (e: MediaQueryListEvent) => (narrow = e.matches);
		mq.addEventListener('change', onChange);

		(async () => {
			try {
				sessions = await listSessions(500);
			} catch (e) {
				error = e instanceof Error ? e.message : String(e);
			} finally {
				loading = false;
			}
			if (scroller) scroller.scrollTop = 7 * HH; // open around the working day
		})();

		return () => mq.removeEventListener('change', onChange);
	});
</script>

<main class="cal">
	<header>
		<h1>🍅 Calendar</h1>
		<nav class="nav">
			<button onclick={() => shift(-1)} aria-label={narrow ? 'Previous day' : 'Previous week'}>‹</button>
			<button class="label" onclick={goToday}>{narrow ? dayLabel() : weekLabel()}</button>
			<button onclick={() => shift(1)} aria-label={narrow ? 'Next day' : 'Next week'}>›</button>
		</nav>
		<a href="/" class="home">← dashboard</a>
	</header>

	{#if error}
		<p class="error">{error}</p>
	{/if}

	<div class="daysrow" style="grid-template-columns:{gridCols}">
		<div class="corner"></div>
		{#each visibleDays as day (dayKey(day))}
			<div class="dayhead" class:istoday={isSameDay(day, today)}>
				<span class="dow">{WEEKDAYS[(day.getDay() + 6) % 7]}</span>
				<span class="date">{day.getDate()}</span>
			</div>
		{/each}
	</div>

	<div class="scroll" bind:this={scroller}>
		<div class="body" style="--hh:{HH}px; height:{HH * 24}px; grid-template-columns:{gridCols}">
			<div class="gutter">
				{#each HOURS as h (h)}
					<div class="hourlabel" style="top:{h * HH}px">
						{String(h).padStart(2, '0')}:00
					</div>
				{/each}
			</div>

			{#each visibleDays as day (dayKey(day))}
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
							{#if showTime(s)}
								<span class="tt">{fmtTime(s)}</span>
							{/if}
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
		font-weight: 500;
	}
	.nav {
		display: flex;
		gap: 0.25rem;
	}
	.nav button {
		border: 1px solid #ccc;
		background: #fff;
		border-radius: 6px;
		padding: 0.3rem 0.7rem;
		cursor: pointer;
		font: inherit;
	}
	.nav button:hover {
		background: #f3f3f3;
	}
	.nav .label {
		min-width: 9rem;
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

	/* grid-template-columns is set inline so it can vary with column count. */
	.daysrow,
	.body {
		display: grid;
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
		color: #333;
	}
	.dayhead.istoday,
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
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.tt {
		color: #888;
		font-size: 0.68rem;
	}
</style>
