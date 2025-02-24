import { type DeleteSelfModel, type UpdateSelfModel } from '$lib/models/user-model';

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
