import { APIError } from '$lib/api-error.svelte';
import {
	UserPaginationSchema,
	type CreateUserModel,
	type UpdateUserModel,
	type UserPaginationModel,
	type UserReqParams
} from '$lib/models/user-model';
import { buildQueryString } from '$lib/utils';
import { safeParse } from 'valibot';
import { apiFetch } from './fetch';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a paginated list of users
export async function GetUsers(params?: UserReqParams): Promise<UserPaginationModel> {
	const qs = params && buildQueryString(params);

	const response = await apiFetch('/api/users' + (qs ? `?${qs}` : ''));

	if (response.ok) {
		const data = (await response.json()) as UserPaginationModel;
		const result = safeParse(UserPaginationSchema, data);

		if (!result.success) throw new APIError(response.status, 'Invalid response from the server');

		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create a new user
export async function CreateUser(data: CreateUserModel): Promise<void> {
	const response = await apiFetch('/api/users', {
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

// Update a user
export async function UpdateUser(id: string, data: UpdateUserModel): Promise<void> {
	const response = await apiFetch(`/api/users/${id}`, {
		method: 'PUT',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify(data)
	});

	if (!response.ok) {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete a user
export async function DeleteUser(id: string): Promise<void> {
	const response = await apiFetch(`/api/users/${id}`, {
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Revoke all sessions of a user
export async function RevokeUserSessions(id: string): Promise<void> {
	const response = await apiFetch(`/api/users/${id}/sessions`, {
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
