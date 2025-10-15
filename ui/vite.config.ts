import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { vite as vidstack } from 'vidstack/plugins';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [vidstack(), tailwindcss(), sveltekit()],
	server: {
		host: '127.0.0.1', // Explicitly bind to IPv4 localhost
		port: 5173,
		strictPort: true,
		hmr: {
			// In dev mode, this forces the socket to use a port other than the default client
			// port, which will be that of the golang backend. It allows HMR to by pass the
			// backend and connect directly to the vite server
			port: 5174
		}
	}
});
