import { type SelfDeleteModel, type SelfUpdateModel } from '$lib/models/user-model';
import { apiFetch } from './fetch';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update self
export async function UpdateSelf(data: SelfUpdateModel): Promise<void> {
	const response = await apiFetch('/api/auth/me', {
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
export async function DeleteSelf(data: SelfDeleteModel): Promise<void> {
	const response = await apiFetch('/api/auth/me', {
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
