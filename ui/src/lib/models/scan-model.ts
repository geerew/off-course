import { picklist, type InferOutput } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const ScanStatusSchema = picklist(['waiting', 'processing', '']);
export type ScanStatus = InferOutput<typeof ScanStatusSchema>;
