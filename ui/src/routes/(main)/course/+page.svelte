<script lang="ts">
	import { page } from '$app/state';
	import { GetAllCourseAssets, GetCourseFromParams } from '$lib/api/course-api';

	import { PdfIcon, VideoIcon, WarningIcon } from '$lib/components/icons';
	import Spinner from '$lib/components/spinner.svelte';
	import Button from '$lib/components/ui/button.svelte';
	import type { AssetModel, ChapteredAssets } from '$lib/models/asset-model';
	import type { CourseModel } from '$lib/models/course-model';
	import { cn, UpdateQueryParam } from '$lib/utils';

	let course = $state<CourseModel>();
	let chapters = $state<ChapteredAssets>({});
	let selectedAsset = $state<AssetModel>();

	let loadPromise = $state(fetchCourse());

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch the course, then the assets for the course, then build a chapter structure from the
	// assets
	async function fetchCourse(): Promise<void> {
		try {
			const pageParams = page.url.searchParams;
			course = await GetCourseFromParams(pageParams);

			const assets = await GetAllCourseAssets(course.id, {
				q: `sort:"assets.chapter asc" sort:"assets.prefix asc"`
			});

			chapters = BuildChapterStructure(assets);

			if (!pageParams || !pageParams.get('a')) {
				let foundAsset: AssetModel | undefined = undefined;

				// If all assets are completed, default to the first asset
				if (!foundAsset && course.progress.percent === 100) {
					foundAsset = Object.values(chapters).flat()[0];
				}

				//Find the first unfinished asset
				for (const chapterAssets of Object.values(chapters)) {
					foundAsset = chapterAssets.find((asset) => !asset.progress.completed);
					if (foundAsset) break;
				}

				if (foundAsset) {
					await UpdateQueryParam('a', foundAsset.id, true);
					selectedAsset = foundAsset;
				}
			} else {
				// If the asset id is in the query params, find it and set it as the selected asset
				const assetId = pageParams.get('a');
				if (assetId) {
					selectedAsset = findAssetById(assetId, chapters);
				}
			}
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Build the course chapter structure
	function BuildChapterStructure(courseAssets: AssetModel[]): ChapteredAssets {
		const chapters: ChapteredAssets = {};

		for (const courseAsset of courseAssets) {
			const chapter = courseAsset.chapter || '(no chapter)';
			!chapters[chapter]
				? (chapters[chapter] = [courseAsset])
				: chapters[chapter]?.push(courseAsset);
		}

		return chapters;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Find an asset  by id
	function findAssetById(id: string, chapters: ChapteredAssets): AssetModel {
		const allAssets = Object.values(chapters).flat();
		const index = allAssets.findIndex((a) => a.id === id);
		if (index === -1) return chapters[Object.keys(chapters)[0]][0];
		return allAssets[index];
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// As the query param `a` changes, this will be reactively called to set the selected asset
	$effect(() => {
		if (!course || Object.keys(chapters).length === 0) return;
		const assetId = page.url.searchParams.get('a');
		if (!assetId) return;
		selectedAsset = findAssetById(assetId, chapters);
	});
</script>

{#await loadPromise}
	<div class="flex justify-center pt-10">
		<Spinner class="bg-foreground-alt-3 size-4" />
	</div>
{:then _}
	{#if course}
		<div
			class="grid grid-cols-[var(--course-menu-width)_1fr] grid-rows-1 gap-6 pt-[calc(var(--height-header)+1))]"
		>
			<div class="relative row-span-full">
				<div class="absolute inset-0">
					<nav
						class="border-foreground-alt-4 sticky top-[calc(var(--height-header)+1px)] left-0 flex h-[calc(100dvh-(var(--height-header)+1px))] w-[--course-menu-width] flex-col gap-2 overflow-x-hidden overflow-y-auto border-r pb-8"
					>
						<div
							class="bg-background border-background-alt-4 sticky top-0 z-[1] flex flex-row gap-3 border-b py-5 pr-3"
						>
							<span class="container-pl font-semibold">{course.title}</span>
						</div>

						{#each Object.keys(chapters) as chapter}
							<div class="container-pl leading-5">
								<div class="border-background-alt-6 flex flex-col gap-1.5 border-b py-1.5 pr-2">
									<span class=" text-background-primary text-sm font-semibold tracking-wide">
										{chapter}
									</span>
									<div class="flex flex-row items-center gap-1">
										<span class="text-foreground-alt-3 text-xs">
											0 / {chapters[chapter].length}
										</span>
									</div>
								</div>

								<div class="ml-auto flex flex-col gap-2 pt-4 pb-3">
									{#each chapters[chapter] as asset, index}
										<Button
											class={cn(
												'text-foreground-alt-2/80 bg-background enabled:hover:bg-background enabled:hover:text-foreground-alt-1 h-auto justify-start text-start duration-50 enabled:hover:underline',
												selectedAsset?.id === asset.id && 'text-foreground-alt-1'
											)}
											onclick={async () => {
												await UpdateQueryParam('a', asset.id, false);
											}}
										>
											{index + 1}. {asset.title}
										</Button>
									{/each}
								</div>
							</div>
						{/each}
					</nav>
				</div>
			</div>
			<main class="container-pr flex w-full py-8">
				<div class="flex w-full place-content-center">
					<div class="flex w-full max-w-5xl flex-col gap-6 pt-1">
						<div class="flex w-full flex-col gap-8">
							{#if selectedAsset}
								<div class="flex w-full flex-row items-center justify-baseline gap-2">
									{#if selectedAsset.assetType === 'video'}
										<VideoIcon class="fill-foreground-alt-2 size-6 stroke-2" />
									{:else if selectedAsset.assetType === 'pdf'}
										<PdfIcon class="fill-foreground-alt-2 size-6 stroke-2" />
									{/if}
									<span class="text-lg font-medium">
										{selectedAsset.title}
									</span>
								</div>
							{/if}
						</div>
					</div>
				</div>
			</main>
		</div>
	{/if}
{:catch error}
	<div class="flex w-full flex-col items-center gap-2 pt-10">
		<WarningIcon class="text-foreground-error size-10" />
		<span class="text-lg">Failed to fetch course</span>
		<span class="text-foreground-alt-3 text-sm">{error.message}</span>
	</div>
{/await}
