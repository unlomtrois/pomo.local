<script lang="ts">
	import '@fontsource-variable/geist';
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}

<style>
	:global(body) {
		margin: 0;
		font-family: 'Geist Variable', system-ui, sans-serif;
		font-weight: 400;
	}
</style>
