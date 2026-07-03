<script lang="ts">
	import { onMount } from 'svelte';
	import { activeSession, listSessions, type Session } from '$lib/api';

	let active = $state<Session | null>(null);
	let sessions = $state<Session[]>([]);
	let error = $state<string | null>(null);
	let loading = $state(true);

	function fmtDuration(nanos: number): string {
		const mins = Math.round(nanos / 1e9 / 60);
		return `${mins}m`;
	}

	function fmtTime(iso: string): string {
		return new Date(iso).toLocaleString();
	}

	onMount(async () => {
		try {
			[active, sessions] = await Promise.all([activeSession(), listSessions(20)]);
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			loading = false;
		}
	});
</script>

<main>
	<h1>🍅 pomo</h1>
	<nav><a href="/calendar">Calendar →</a></nav>

	{#if loading}
		<p>Loading…</p>
	{:else if error}
		<p class="error">Could not reach the daemon: {error}</p>
	{:else}
		<section>
			<h2>Active session</h2>
			{#if active}
				<p><strong>{active.topic}</strong> — ends {fmtTime(active.stop_time)}</p>
			{:else}
				<p>No active session.</p>
			{/if}
		</section>

		<section>
			<h2>Recent sessions</h2>
			{#if sessions.length === 0}
				<p>Nothing recorded yet.</p>
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
	{/if}
</main>

<style>
	main {
		max-width: 48rem;
		margin: 2rem auto;
		padding: 0 1rem;
	}
	h1 {
		font-size: 2rem;
	}
	section {
		margin-top: 2rem;
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	th,
	td {
		text-align: left;
		padding: 0.4rem 0.6rem;
		border-bottom: 1px solid #ddd;
	}
	.error {
		color: #b00020;
	}
</style>
