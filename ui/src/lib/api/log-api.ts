import { APIError } from '$lib/api-error.svelte';
import {
	LogPaginationSchema,
	type LogPaginationModel,
	type LogReqParams
} from '$lib/models/log-model';
import { buildQueryString } from '$lib/utils';
import { safeParse } from 'valibot';
import { apiFetch } from './fetch';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a paginated list of logs
export async function GetLogs(params?: LogReqParams): Promise<LogPaginationModel> {
	const qs = params && buildQueryString(params);
	const response = await apiFetch(`/api/logs` + (qs ? `?${qs}` : ''));

	if (response.ok) {
		const data = (await response.json()) as LogPaginationModel;
		const result = safeParse(LogPaginationSchema, data);

		if (!result.success) throw new APIError(response.status, 'Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}
