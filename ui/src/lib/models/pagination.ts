import { array, number, object, union, type InferOutput } from 'valibot';
import { UserSchema } from './user';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type PaginationReqParams = {
	page?: number;
	perPage?: number;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type PaginationRespParams = {
	page: number;
	perPage: number;
	perPages: number[];
	totalItems: number;
	totalPages: number;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const PaginationSchema = object({
	page: number(),
	perPage: number(),
	totalItems: number(),
	totalPages: number(),
	items: union([array(UserSchema)])
});

export type Pagination = InferOutput<typeof PaginationSchema>;
