import { goto } from '$app/navigation';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { AssetModel, Chapters } from './models/asset-model';

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

// Build the course chapter structure
export function BuildChapterStructure(courseAssets: AssetModel[]): Chapters {
	const chapters: Chapters = {};

	// Group by chapter
	const assetsByChapter: Record<string, AssetModel[]> = {};
	for (const asset of courseAssets) {
		const chapter = asset.chapter || '(no chapter)';
		if (!assetsByChapter[chapter]) assetsByChapter[chapter] = [];
		assetsByChapter[chapter].push(asset);
	}

	// Group by prefix within each chapter to create lessons
	for (const [chapterName, chapterAssets] of Object.entries(assetsByChapter)) {
		console.log(`Processing chapter: ${chapterName} with ${chapterAssets.length} assets`);
		const lessonMap: Record<number, AssetModel[]> = {};

		// Group assets by prefix
		for (const asset of chapterAssets) {
			if (!lessonMap[asset.prefix]) lessonMap[asset.prefix] = [];
			lessonMap[asset.prefix].push(asset);
		}

		// Convert to lessons array
		chapters[chapterName] = Object.entries(lessonMap)
			.map(([prefix, assets]) => {
				// Sort assets by subPrefix
				const sortedAssets = assets.sort((a, b) => {
					if (a.subPrefix === undefined && b.subPrefix === undefined) return 0;
					if (a.subPrefix === undefined) return -1;
					if (b.subPrefix === undefined) return 1;
					return a.subPrefix - b.subPrefix;
				});

				// The group title is the title of the first asset in the sorted list
				let groupTitle = sortedAssets[0].title;

				const allAttachments = sortedAssets.flatMap((asset) => asset.attachments);
				const completedAssets = sortedAssets.filter((asset) => asset.progress?.completed);
				const startedAssets = sortedAssets.filter(
					(asset) =>
						(asset.progress?.videoPos && asset.progress?.videoPos > 0) || asset.progress?.completed
				);

				console.log(
					`Lesson ${prefix} (${groupTitle}) has ${sortedAssets.length} assets, ` +
						`${completedAssets.length} completed, ${startedAssets.length} started`
				);

				return {
					prefix: parseInt(prefix),
					title: groupTitle,
					assets: sortedAssets,
					completed: completedAssets.length === sortedAssets.length,
					startedAssetCount: startedAssets.length,
					completedAssetCount: completedAssets.length,
					attachments: allAttachments
				};
			})
			.sort((a, b) => a.prefix - b.prefix);
	}

	return chapters;
}
