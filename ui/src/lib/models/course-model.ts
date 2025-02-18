import { boolean, object, picklist, string, type InferOutput } from 'valibot';
import { BaseSchema } from './base-model';
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

// course
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
