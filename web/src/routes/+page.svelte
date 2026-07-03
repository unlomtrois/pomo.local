<script lang="ts">
	import { onMount } from 'svelte';
	import { activeSession, listSessions, stopActive, type Session } from '$lib/api';
	import QuickAdd from '$lib/QuickAdd.svelte';

	let active = $state<Session | null>(null);
	let sessions = $state<Session[]>([]);
	let error = $state<string | null>(null);
	let loading = $state(true);
	let stopping = $state(false);

	function fmtDuration(nanos: number): string {
		return `${Math.round(nanos / 1e9 / 60)}m`;
	}
	function fmtTime(iso: string): string {
		return new Date(iso).toLocaleString();
	}

	async function reload() {
		[active, sessions] = await Promise.all([activeSession(), listSessions(20)]);
	}

	async function stop() {
		stopping = true;
		try {
			await stopActive();
			await reload();
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			stopping = false;
		}
	}

	onMount(async () => {
		try {
			await reload();
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	});
</script>

<main>
	<header>
		<h1>🍅 pomo</h1>
		<nav><a href="/calendar">Calendar →</a></nav>
	</header>

	{#if error}
		<p class="error">{error}</p>
	{/if}

	{#if loading}
		<p>Loading…</p>
	{:else if active}
		<section class="activecard">
			<div class="dot"></div>
			<div class="grow">
				<div class="topic">{active.topic || 'Focus session'}</div>
				<div class="meta">ends {fmtTime(active.stop_time)}</div>
			</div>
			<button class="stop" onclick={stop} disabled={stopping}>
				{stopping ? '…' : 'Stop'}
			</button>
		</section>
	{:else}
		<section class="quick">
			<QuickAdd onstarted={reload} />
		</section>
	{/if}

	<section>
		<h2>Recent sessions</h2>
		{#if sessions.length === 0}
			<p class="muted">Nothing recorded yet.</p>
		{:else}
			<table>
				<thead>
					<tr><th>Topic</th><th>Started</th><th>Duration</th><th>Status</th><th>Source</th></tr>
				</thead>
				<tbody>
					{#each sessions as s (s.id)}
						<tr>
							<td>{s.topic}</td>
							<td>{fmtTime(s.start_time)}</td>
							<td>{fmtDuration(s.duration)}</td>
							<td>{s.status}</td>
							<td>{s.source}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</section>
</main>

<style>
	main {
		max-width: 48rem;
		margin: 2rem auto;
		padding: 0 1rem;
	}
	header {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
	}
	h1 {
		font-size: 2rem;
		font-weight: 500;
	}
	nav a {
		color: #555;
		text-decoration: none;
	}
	h2 {
		font-size: 1.1rem;
		font-weight: 500;
	}
	section {
		margin-top: 2rem;
	}
	.quick {
		margin-top: 2.5rem;
	}
	.activecard {
		display: flex;
		align-items: center;
		gap: 0.9rem;
		border: 1px solid #eee;
		border-left: 3px solid #1a9e4b;
		border-radius: 10px;
		padding: 1rem 1.1rem;
	}
	.activecard .dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		background: #1a9e4b;
	}
	.activecard .grow {
		flex: 1;
	}
	.activecard .topic {
		font-size: 1.1rem;
	}
	.activecard .meta {
		color: #888;
		font-size: 0.85rem;
	}
	.stop {
		border: 1px solid #e03a3a;
		color: #e03a3a;
		background: #fff;
		border-radius: 8px;
		padding: 0.45rem 1rem;
		cursor: pointer;
		font: inherit;
	}
	.stop:disabled {
		opacity: 0.6;
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	th,
	td {
		text-align: left;
		padding: 0.4rem 0.6rem;
		border-bottom: 1px solid #eee;
		font-weight: 400;
	}
	th {
		color: #888;
		font-size: 0.85rem;
	}
	.muted {
		color: #888;
	}
	.error {
		color: #b00020;
	}
</style>
