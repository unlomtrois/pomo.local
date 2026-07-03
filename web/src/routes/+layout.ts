// Render as a single-page app: no SSR, and prerender the shell so adapter-static
// can emit fully static assets for the daemon to serve.
export const ssr = false;
export const prerender = true;
