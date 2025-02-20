import { FsSchema, type FsModel } from '$lib/models/fs-model';
import { safeParse } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get filesystem information. When the path is not provided, the backend will return the
// available drives
export async function GetFileSystem(path?: string): Promise<FsModel> {
	let response: Response;
	if (path) {
		response = await fetch(`/api/filesystem/${window.btoa(encodeURIComponent(path))}`);
	} else {
		response = await fetch(`/api/filesystem`);
	}

	if (response.ok) {
		const data = (await response.json()) as FsModel;
		const result = safeParse(FsSchema, data);

		if (!result.success) throw new Error('Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new Error(data.message);
	}
}
