// $lib/api/fetch.ts

import { auth } from '$lib/auth.svelte';

export async function apiFetch(input: RequestInfo, init?: RequestInit): Promise<Response> {
	const response = await fetch(input, init);

	// Handle forbidden / unauthorized (session expired or revoked)
	if (response.status === 403 || response.status === 401) {
		try {
			const data = await response.json();

			// Optional: Show user-friendly message (but avoid infinite loops)
			console.warn('Session expired or unauthorized:', data.message);

			// Clear auth state and redirect to login
			auth.empty();
			window.location.href = '/auth/login';
		} catch {
			// Handle edge case where server returns HTML (e.g., redirect or error page)
			auth.empty();
			window.location.href = '/auth/login';
		}
	}

	return response;
}
