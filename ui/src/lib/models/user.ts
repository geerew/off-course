import { object, picklist, string, type InferOutput } from 'valibot';
import type { PaginationReqParams } from './pagination';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const UserRoleSchema = picklist(['admin', 'user']);
export type UserRole = InferOutput<typeof UserRoleSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const UserSchema = object({
	id: string(),
	username: string(),
	displayName: string(),
	role: UserRoleSchema
});

export type UserModel = InferOutput<typeof UserSchema>;
export type UsersModel = UserModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const CreateUserSchema = object({
	username: string(),
	displayName: string(),
	password: string(),
	role: UserRoleSchema
});

export type CreateUserModel = InferOutput<typeof CreateUserSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type UserReqParams = PaginationReqParams & {
	orderBy?: string;
};
