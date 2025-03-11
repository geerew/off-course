import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export function capitalizeFirstLetter(str: string) {
	return String(str).charAt(0).toUpperCase() + String(str).slice(1);
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export function buildQueryString(params: Record<string, string | number | undefined>): string {
	const searchParams = new URLSearchParams();

	Object.entries(params).forEach(([key, value]) => {
		if (value !== undefined) {
			searchParams.append(key, value.toString());
		}
	});

	return searchParams.toString();
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export function remCalc(px: number | string, base: number = 16): number {
	const tempPx = `${px}`.replace('px', '');
	return (1 / base) * parseInt(tempPx);
}
