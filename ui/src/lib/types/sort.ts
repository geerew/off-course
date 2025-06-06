import { array, object, string, type InferOutput } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Sort columns schema
export const SortColumnsSchema = array(
	object({
		label: string(),
		column: string(),
		asc: string(),
		desc: string()
	})
);

export type SortColumns = InferOutput<typeof SortColumnsSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type SortDirection = 'asc' | 'desc';
