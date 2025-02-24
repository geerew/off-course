import { APIError } from '$lib/api-error.svelte';
import { ScanSchema, type ScansModel, type StartScanModel } from '$lib/models/scan-model';
import { array, safeParse } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get all scans
export async function GetScans(): Promise<ScansModel> {
	const response = await fetch('/api/scans');

	if (response.ok) {
		const data = (await response.json()) as ScansModel;
		const result = safeParse(array(ScanSchema), data);

		if (!result.success) throw new APIError(response.status, 'Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Start a scan
export async function StartScan(data: StartScanModel): Promise<void> {
	const response = await fetch('/api/scans', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});

	if (!response.ok) {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}
