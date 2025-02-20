import { object, picklist, string, type InferOutput } from 'valibot';
import { BaseSchema } from './base-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const ScanStatusSchema = picklist(['waiting', 'processing', '']);
export type ScanStatus = InferOutput<typeof ScanStatusSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan schema
export const ScanSchema = object({
	...BaseSchema.entries,
	courseId: string(),
	status: ScanStatusSchema
});

export type ScanModel = InferOutput<typeof ScanSchema>;
export type ScansModel = ScanModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create scan schema
export const StartScanSchema = object({
	courseId: string()
});

export type StartScanModel = InferOutput<typeof StartScanSchema>;
