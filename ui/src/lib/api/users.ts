import {
	UserPaginationSchema,
	type CreateUserModel,
	type DeleteSelfModel,
	type UpdateSelfModel,
	type UpdateUserModel,
	type UserPaginationModel,
	type UserReqParams
} from '$lib/models/user';
import { buildQueryString } from '$lib/utils';
import { safeParse } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a paginated list of users
export async function GetUsers(params?: UserReqParams): Promise<UserPaginationModel> {
	const qs = params && buildQueryString(params);

	const response = await fetch('/api/users' + (qs ? `?${qs}` : ''));

	if (response.ok) {
		const data = (await response.json()) as UserPaginationModel;
		const result = safeParse(UserPaginationSchema, data);

		if (!result.success) throw new Error('Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new Error(data.message);
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create a new user
export async function CreateUser(data: CreateUserModel): Promise<void> {
	const response = await fetch('/api/users', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});

	if (!response.ok) {
		const data = await response.json();
		throw new Error(data.message);
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update a user
export async function UpdateUser(id: string, data: UpdateUserModel): Promise<void> {
	const response = await fetch(`/api/users/${id}`, {
		method: 'PUT',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify(data)
	});

	if (!response.ok) {
		const data = await response.json();
		throw new Error(data.message);
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete a user
export async function DeleteUser(id: string): Promise<void> {
	const response = await fetch(`/api/users/${id}`, {
		method: 'DELETE',
		headers: {
			'Content-Type': 'application/json'
		}
	});

	if (!response.ok) {
		const data = await response.json();
		throw new Error(data.message);
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update self
export async function UpdateSelf(data: UpdateSelfModel): Promise<void> {
	const response = await fetch('/api/auth/me', {
		method: 'PUT',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify(data)
	});

	if (!response.ok) {
		const data = await response.json();
		throw new Error(data.message);
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete self
export async function DeleteSelf(data: DeleteSelfModel): Promise<void> {
	const response = await fetch('/api/auth/me', {
		method: 'DELETE',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify(data)
	});

	if (!response.ok) {
		const data = await response.json();
		throw new Error(data.message);
	}
}
