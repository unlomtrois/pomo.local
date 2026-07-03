import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) => filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			// SPA mode: emit a static fallback so client-side routes work when
			// served by the Go daemon (which will embed web/build and host it
			// on pomo.local). The dashboard fetches session data from the daemon
			// API at runtime, so no server-side rendering is needed.
			adapter: adapter({ fallback: 'index.html' })
		})
	]
});
