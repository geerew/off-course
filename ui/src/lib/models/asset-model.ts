import {
	array,
	boolean,
	number,
	object,
	optional,
	picklist,
	string,
	type InferOutput
} from 'valibot';
import { BaseSchema } from './base-model';
import { BasePaginationSchema, type PaginationReqParams } from './pagination-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset type schema
const AssetTypeSchema = picklist(['video', 'html', 'pdf', 'markdown', 'text']);
export type AssetType = InferOutput<typeof AssetTypeSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset progress schema
export const AssetProgressSchema = object({
	position: number(),
	completed: boolean(),
	completedAt: string()
});

export type AssetProgressModel = InferOutput<typeof AssetProgressSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset progress update schema
export const AssetProgressUpdateSchema = object({
	position: optional(number()),
	completed: boolean()
});

export type AssetProgressUpdateModel = InferOutput<typeof AssetProgressUpdateSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset video metadata schema
export const AssetVideoMetadataSchema = object({
	durationSec: number(),
	container: string(),
	mimeType: string(),
	sizeBytes: number(),
	overallBPS: number(),
	videoCodec: string(),
	width: number(),
	height: number(),
	fpsNum: number(),
	fpsDen: number()
});

export type AssetVideoMetadataModel = InferOutput<typeof AssetVideoMetadataSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset audio metadata schema
export const AssetAudioMetadataSchema = object({
	Language: string(),
	Codec: string(),
	Profile: string(),
	Channels: number(),
	ChannelLayout: string(),
	SampleRate: number(),
	BitRate: number()
});

export type AssetAudioMetadataModel = InferOutput<typeof AssetAudioMetadataSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset metadata schema
export const AssetMetadataSchema = object({
	video: optional(AssetVideoMetadataSchema),
	audio: optional(AssetAudioMetadataSchema)
});

export type AssetMetadataModel = InferOutput<typeof AssetMetadataSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset schema
export const AssetSchema = object({
	...BaseSchema.entries,
	courseId: string(),
	lessonId: string(),
	title: string(),
	prefix: number(),
	subPrefix: optional(number()),
	subTitle: optional(string()),
	module: string(),
	path: string(),
	type: AssetTypeSchema,
	metadata: AssetMetadataSchema,
	progress: AssetProgressSchema
});

export type AssetModel = InferOutput<typeof AssetSchema>;
export type AssetsModel = AssetModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AssetPaginationSchema = object({
	...BasePaginationSchema.entries,
	items: array(AssetSchema)
});

export type AssetPaginationModel = InferOutput<typeof AssetPaginationSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type AssetReqParams = PaginationReqParams & {
	q?: string;
};
