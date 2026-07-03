<script lang="ts">
	import { startSession, type Session } from '$lib/api';

	// Called after a session is successfully started so the parent can refresh.
	let { onstarted }: { onstarted?: (s: Session) => void } = $props();

	const CX = 100;
	const CY = 100;
	const R = 80;
	const STEP = 5; // snap minutes
	const MIN = 5;
	const MAX = 60;
	const PRESETS = [5, 15, 25, 50];

	let minutes = $state(25);
	let topic = $state('');
	let busy = $state(false);
	let error = $state<string | null>(null);
	let dragging = $state(false);
	let svgEl: SVGSVGElement;

	// Angle 0 min = 12 o'clock (top); clockwise.
	function pointFor(min: number): { x: number; y: number } {
		const a = (min / 60) * 2 * Math.PI - Math.PI / 2;
		return { x: CX + R * Math.cos(a), y: CY + R * Math.sin(a) };
	}
	let handle = $derived(pointFor(minutes));

	// Arc path from the top, clockwise to the handle.
	let arc = $derived.by(() => {
		const p = pointFor(minutes);
		const large = minutes > 30 ? 1 : 0;
		return `M ${CX} ${CY - R} A ${R} ${R} 0 ${large} 1 ${p.x} ${p.y}`;
	});

	// Clock tick marks every 5 minutes.
	const ticks = Array.from({ length: 12 }, (_, i) => {
		const a = (i / 12) * 2 * Math.PI - Math.PI / 2;
		const outer = R + 8;
		const inner = R + (i % 3 === 0 ? 2 : 5);
		return {
			x1: CX + inner * Math.cos(a),
			y1: CY + inner * Math.sin(a),
			x2: CX + outer * Math.cos(a),
			y2: CY + outer * Math.sin(a),
			major: i % 3 === 0
		};
	});

	function setFromEvent(e: PointerEvent) {
		const rect = svgEl.getBoundingClientRect();
		const vx = ((e.clientX - rect.left) / rect.width) * 200;
		const vy = ((e.clientY - rect.top) / rect.height) * 200;
		const deg = (Math.atan2(vy - CY, vx - CX) * 180) / Math.PI;
		const fromTop = (deg + 90 + 360) % 360;
		let m = Math.round(fromTop / 360 * 60 / STEP) * STEP;
		if (m === 0) m = 60; // top maps to a full hour, not zero
		minutes = Math.min(MAX, Math.max(MIN, m));
	}

	function onPointerDown(e: PointerEvent) {
		dragging = true;
		svgEl.setPointerCapture(e.pointerId);
		setFromEvent(e);
	}
	function onPointerMove(e: PointerEvent) {
		if (dragging) setFromEvent(e);
	}
	function onPointerUp(e: PointerEvent) {
		dragging = false;
		svgEl.releasePointerCapture(e.pointerId);
	}

	async function start() {
		busy = true;
		error = null;
		try {
			const s = await startSession({ topic: topic.trim(), duration: `${minutes}m` });
			topic = '';
			onstarted?.(s);
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			busy = false;
		}
	}
</script>

<div class="quickadd">
	<svg
		viewBox="0 0 200 200"
		class="dial"
		class:dragging
		bind:this={svgEl}
		onpointerdown={onPointerDown}
		onpointermove={onPointerMove}
		onpointerup={onPointerUp}
		role="slider"
		aria-label="Session duration in minutes"
		aria-valuemin={MIN}
		aria-valuemax={MAX}
		aria-valuenow={minutes}
		tabindex="0"
		onkeydown={(e) => {
			if (e.key === 'ArrowUp' || e.key === 'ArrowRight') minutes = Math.min(MAX, minutes + STEP);
			if (e.key === 'ArrowDown' || e.key === 'ArrowLeft') minutes = Math.max(MIN, minutes - STEP);
		}}
	>
		{#each ticks as t}
			<line x1={t.x1} y1={t.y1} x2={t.x2} y2={t.y2} class="tick" class:major={t.major} />
		{/each}
		<circle cx={CX} cy={CY} r={R} class="track" />
		{#if minutes >= MAX}
			<circle cx={CX} cy={CY} r={R} class="progress" />
		{:else}
			<path d={arc} class="progress" />
		{/if}
		<circle cx={handle.x} cy={handle.y} r="9" class="knob" />
		<text x={CX} y={CY - 8} class="big">{minutes}</text>
		<text x={CX} y={CY + 18} class="unit">min</text>
	</svg>

	<div class="presets">
		{#each PRESETS as p}
			<button type="button" class:sel={minutes === p} onclick={() => (minutes = p)}>{p}m</button>
		{/each}
	</div>

	<input
		class="topic"
		type="text"
		placeholder="What are you working on?"
		bind:value={topic}
		onkeydown={(e) => e.key === 'Enter' && start()}
	/>

	<button class="start" onclick={start} disabled={busy}>
		{busy ? 'Starting…' : `Start ${minutes}m`}
	</button>

	{#if error}
		<p class="error">{error}</p>
	{/if}
</div>

<style>
	.quickadd {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.9rem;
		max-width: 20rem;
		margin: 0 auto;
	}
	.dial {
		width: 15rem;
		height: 15rem;
		touch-action: none;
		cursor: pointer;
		user-select: none;
	}
	.dial.dragging {
		cursor: grabbing;
	}
	.track {
		fill: none;
		stroke: #eee;
		stroke-width: 10;
	}
	.progress {
		fill: none;
		stroke: #e03a3a;
		stroke-width: 10;
		stroke-linecap: round;
	}
	.knob {
		fill: #fff;
		stroke: #e03a3a;
		stroke-width: 3;
	}
	.tick {
		stroke: #ccc;
		stroke-width: 1.5;
	}
	.tick.major {
		stroke: #999;
		stroke-width: 2;
	}
	.big {
		text-anchor: middle;
		dominant-baseline: central;
		font-size: 2.4rem;
		fill: #222;
	}
	.unit {
		text-anchor: middle;
		dominant-baseline: central;
		font-size: 0.8rem;
		fill: #999;
	}
	.presets {
		display: flex;
		gap: 0.4rem;
	}
	.presets button {
		border: 1px solid #ddd;
		background: #fff;
		border-radius: 999px;
		padding: 0.25rem 0.7rem;
		cursor: pointer;
		font: inherit;
		color: #555;
	}
	.presets button.sel {
		border-color: #e03a3a;
		color: #e03a3a;
		background: #fff5f5;
	}
	.topic {
		width: 100%;
		box-sizing: border-box;
		padding: 0.55rem 0.7rem;
		border: 1px solid #ddd;
		border-radius: 8px;
		font: inherit;
	}
	.start {
		width: 100%;
		padding: 0.6rem;
		border: none;
		border-radius: 8px;
		background: #e03a3a;
		color: #fff;
		font: inherit;
		font-size: 1rem;
		cursor: pointer;
	}
	.start:disabled {
		opacity: 0.6;
		cursor: default;
	}
	.error {
		color: #b00020;
		margin: 0;
		font-size: 0.85rem;
	}
</style>
