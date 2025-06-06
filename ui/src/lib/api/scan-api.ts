import { APIError } from '$lib/api-error.svelte';
import {
	ScanPaginationSchema,
	type ScanModel,
	type ScanPaginationModel,
	type ScanReqParams,
	type StartScanModel
} from '$lib/models/scan-model';
import { buildQueryString } from '$lib/utils';
import { safeParse } from 'valibot';
import { apiFetch } from './fetch';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a paginated list of scans
export async function GetScans(params?: ScanReqParams): Promise<ScanPaginationModel> {
	const qs = params && buildQueryString(params);
	const response = await apiFetch(`/api/scans` + (qs ? `?${qs}` : ''));

	if (response.ok) {
		const data = (await response.json()) as ScanPaginationModel;
		const result = safeParse(ScanPaginationSchema, data);

		if (!result.success) throw new APIError(response.status, 'Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get all the scans
export async function GetAllScans(params?: ScanReqParams): Promise<ScanModel[]> {
	let assets: ScanModel[] = [];
	let page = 1;
	let totalPages = 1;

	do {
		try {
			const response = await GetScans({ ...params, page, perPage: 100 });

			if (response.totalItems > 0) {
				assets.push(...response.items);
				totalPages = response.totalPages;
				page++;
			} else {
				break;
			}
		} catch (error) {
			throw error;
		}
	} while (page <= totalPages);

	return assets;
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Start a scan
export async function StartScan(data: StartScanModel): Promise<void> {
	const response = await apiFetch('/api/scans', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});

	if (!response.ok) {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete a scan
export async function DeleteScan(id: string): Promise<void> {
	const response = await apiFetch(`/api/scans/${id}`, {
		method: 'DELETE',
		headers: {
			'Content-Type': 'application/json'
		}
	});

	if (!response.ok) {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}
