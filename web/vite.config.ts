import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

// Where the dev server proxies API calls. Defaults to the local daemon; set
// POMO_PROXY_TARGET to point at a remote one, e.g. http://pomo.local:7420.
// (globalThis access avoids needing @types/node just for process.env.)
const env = (globalThis as { process?: { env?: Record<string, string | undefined> } }).process?.env;
const proxyTarget = env?.POMO_PROXY_TARGET ?? 'http://127.0.0.1:7420';

export default defineConfig({
	// In dev, proxy the daemon's API so the browser sees it as same-origin
	// (avoids CORS and mirrors the embedded-in-daemon production setup).
	server: {
		proxy: {
			'/api': proxyTarget,
			'/healthz': proxyTarget
		}
	},
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
