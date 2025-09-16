import { goto } from '$app/navigation';
import { clsx, type ClassValue } from 'clsx';
import DOMPurify from 'dompurify';
import MarkdownIt from 'markdown-it';
import { twMerge } from 'tailwind-merge';
import type { VideoMimeType } from 'vidstack';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export function capitalizeFirstLetter(str: string) {
	return String(str).charAt(0).toUpperCase() + String(str).slice(1);
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export function buildQueryString(
	params: Record<string, string | number | boolean | undefined>
): string {
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export async function UpdateQueryParam(key: string, value: string, replaceState: boolean) {
	if (typeof window === 'undefined') return;

	const url = new URL(window.location.href);
	url.searchParams.set(key, value);

	await goto(url.toString(), { replaceState, keepFocus: true, noScroll: true });
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Enable HTML inside MD, linkify URLs, etc—tweak to taste
const md = new MarkdownIt({ html: true, linkify: true, typographer: true });

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Sanitize the rendered HTML to prevent XSS attacks
export function renderMarkdown(raw: string): string {
	return DOMPurify.sanitize(md.render(raw));
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export function toVideoMimeType(m?: string): VideoMimeType | undefined {
	if (!m) return undefined;
	const lower = m.toLowerCase();

	if (lower === 'video/mp4' || lower === 'video/quicktime') return 'video/mp4';
	if (lower === 'video/webm') return 'video/webm';
	if (lower === 'video/3gpp' || lower === 'video/3gp') return 'video/3gp';
	if (lower === 'video/ogg') return 'video/ogg';
	if (lower === 'video/avi') return 'video/avi';
	if (lower === 'video/mpeg') return 'video/mpeg';

	return undefined;
}
