import { PaginationSchema, type Pagination } from '$lib/models/pagination';
import type { CreateUserModel, UserReqParams } from '$lib/models/user';
import { buildQueryString } from '$lib/utils';
import { safeParse } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a paginated list of users
export async function GetUsers(params?: UserReqParams): Promise<Pagination> {
	const qs = params && buildQueryString(params);

	const response = await fetch('/api/users' + (qs ? `?${qs}` : ''));

	if (response.ok) {
		const data = (await response.json()) as Pagination;
		const result = safeParse(PaginationSchema, data);

		if (!result.success) throw new Error('Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new Error(data.message);
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create a new user
export async function CreateUser(user: CreateUserModel): Promise<void> {
	const response = await fetch('/api/users', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(user)
	});

	if (!response.ok) {
		const data = await response.json();
		throw new Error(data.message);
	}
}
