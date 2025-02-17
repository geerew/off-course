import { picklist, type InferOutput } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const ScanStatusSchema = picklist(['waiting', 'processing', '']);
export type ScanStatus = InferOutput<typeof ScanStatusSchema>;
