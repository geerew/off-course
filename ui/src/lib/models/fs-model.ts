import { array, enum_, number, object, string, type InferOutput } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const FsPathClassification = {
	None: 0,
	Ancestor: 1,
	Course: 2,
	Descendant: 3
} as const;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const FsFileSchema = object({
	title: string(),
	path: string()
});

export type FsFileModel = InferOutput<typeof FsFileSchema>;
export type FsFilesModel = FsFileModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const FsDirSchema = object({
	title: string(),
	path: string(),
	classification: enum_(FsPathClassification)
});

export type FsDirModel = InferOutput<typeof FsDirSchema>;
export type FsDirsModel = FsDirModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const FsSchema = object({
	count: number(),
	directories: array(FsDirSchema),
	files: array(FsFileSchema)
});

export type FsModel = InferOutput<typeof FsSchema>;
