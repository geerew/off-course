import { APIError } from '$lib/api-error.svelte';
import { FsSchema, type FsModel } from '$lib/models/fs-model';
import { safeParse } from 'valibot';
import { apiFetch } from './fetch';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get filesystem information. When the path is not provided, the backend will return the
// available drives
export async function GetFileSystem(path?: string): Promise<FsModel> {
	let response: Response;
	if (path) {
		response = await apiFetch(`/api/filesystem/${window.btoa(encodeURIComponent(path))}`);
	} else {
		response = await apiFetch(`/api/filesystem`);
	}

	if (response.ok) {
		const data = (await response.json()) as FsModel;
		const result = safeParse(FsSchema, data);

		if (!result.success) throw new APIError(response.status, 'Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}
