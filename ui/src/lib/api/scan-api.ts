import { APIError } from '$lib/api-error.svelte';
import type { ScanCreateModel, ScanModel } from '$lib/models/scan-model';
import { apiFetch } from './fetch';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Start a scan
export async function StartScan(data: ScanCreateModel): Promise<void> {
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type ScanUpdateEvent = {
	type: 'all_scans' | 'scan_update' | 'scan_deleted' | 'error';
	data: ScanModel[] | ScanModel | { id: string } | { message: string };
};

export type ScanSubscriptionCallbacks = {
	onUpdate?: (event: ScanUpdateEvent) => void;
	onError?: (error: Error) => void;
	onClose?: () => void;
};

// Subscribe to scan updates via Server-Sent Events
export function subscribeToScans(callbacks: ScanSubscriptionCallbacks): () => void {
	const eventSource = new EventSource('/api/scans/stream');
	let isClosed = false;

	const close = () => {
		if (!isClosed) {
			isClosed = true;
			eventSource.close();
			callbacks.onClose?.();
		}
	};

	eventSource.onmessage = (event) => {
		try {
			const data = JSON.parse(event.data) as ScanUpdateEvent;
			callbacks.onUpdate?.(data);
		} catch (error) {
			callbacks.onError?.(error as Error);
		}
	};

	eventSource.onerror = (error) => {
		callbacks.onError?.(new Error('SSE connection error'));
		// Auto-reconnect on error (EventSource handles this automatically)
	};

	return close;
}
