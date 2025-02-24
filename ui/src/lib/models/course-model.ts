import { array, boolean, object, picklist, string, type InferOutput } from 'valibot';
import { BaseSchema } from './base-model';
import { BasePaginationSchema, type PaginationReqParams } from './pagination-model';
import { ScanStatusSchema } from './scan-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const UserRoleSchema = picklist(['admin', 'user']);
export type UserRole = InferOutput<typeof UserRoleSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const SelectUserRoles = [
	{ value: 'user', label: 'User' },
	{ value: 'admin', label: 'Admin' }
];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course schema
// TODO add progress
export const CourseSchema = object({
	...BaseSchema.entries,
	title: string(),
	path: string(),
	hasCard: boolean(),
	available: boolean(),
	scanStatus: ScanStatusSchema
});

export type CourseModel = InferOutput<typeof CourseSchema>;
export type CoursesModel = CourseModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create course schema
export const CreateCourseSchema = object({
	title: string(),
	path: string()
});

export type CreateCourseModel = InferOutput<typeof CreateCourseSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create course tag schema
export const CreateCourseTagSchema = object({
	tag: string()
});

export type CreateCourseTagModel = InferOutput<typeof CreateCourseTagSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course tag schema
export const CourseTagSchema = object({
	id: string(),
	tag: string()
});

export type CourseTagModel = InferOutput<typeof CourseTagSchema>;
export type CourseTagsModel = CourseTagModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const CoursePaginationSchema = object({
	...BasePaginationSchema.entries,
	items: array(CourseSchema)
});

export type CoursePaginationModel = InferOutput<typeof CoursePaginationSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type CourseReqParams = PaginationReqParams & {
	orderBy?: string;
};
