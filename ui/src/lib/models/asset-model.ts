import {
	array,
	boolean,
	number,
	object,
	optional,
	picklist,
	record,
	string,
	type InferOutput
} from 'valibot';
import { AttachmentSchema } from './attachment-model';
import { BaseSchema } from './base-model';
import { BasePaginationSchema, type PaginationReqParams } from './pagination-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset type schema
const AssetTypeSchema = picklist(['video', 'html', 'pdf', 'md', 'txt']);
export type AssetType = InferOutput<typeof AssetTypeSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AssetVideoMetadataSchema = object({
	duration: number(),
	width: number(),
	height: number(),
	codec: string(),
	resolution: string()
});

export type AssetVideoMetadataModel = InferOutput<typeof AssetVideoMetadataSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AssetProgressSchema = object({
	videoPos: number(),
	completed: boolean(),
	completedAt: string()
});

export type AssetProgressModel = InferOutput<typeof AssetProgressSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset schema
export const AssetSchema = object({
	...BaseSchema.entries,
	title: string(),
	prefix: number(),
	subPrefix: optional(number()),
	subTitle: optional(string()),
	chapter: string(),
	path: string(),
	assetType: AssetTypeSchema,
	hasDescription: boolean(),
	videoMetadata: optional(AssetVideoMetadataSchema),
	attachments: array(AttachmentSchema),
	progress: optional(AssetProgressSchema)
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const AssetGroupSchema = object({
	prefix: number(),
	title: string(),
	assets: array(AssetSchema),
	attachments: array(AttachmentSchema),
	completed: boolean(),
	startedAssetCount: number(),
	completedAssetCount: number()
});

export type AssetGroup = InferOutput<typeof AssetGroupSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const ChaptersSchema = record(string(), array(AssetGroupSchema));

export type Chapters = InferOutput<typeof ChaptersSchema>;
