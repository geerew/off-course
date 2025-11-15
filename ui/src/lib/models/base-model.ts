import { object, string } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const BaseSchema = object({
	id: string(),
	createdAt: string(),
	updatedAt: string()
});
