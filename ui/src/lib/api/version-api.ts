import { APIError } from '$lib/api-error.svelte';
import { apiFetch } from './fetch';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type VersionModel = {
	version: string;
	latestRelease?: string;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get the application version
export async function GetVersion(): Promise<VersionModel> {
	const response = await apiFetch('/api/version');

	if (response.ok) {
		const data = (await response.json()) as VersionModel;
		return data;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}
