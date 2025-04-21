<script lang="ts">
	import { page } from '$app/state';
	import type { APIError } from '$lib/api-error.svelte';
	import {
		GetAllCourseAssets,
		GetCourseFromParams,
		UpdateCourseAssetProgress
	} from '$lib/api/course-api';

	import { PdfIcon, TickIcon, VideoIcon, WarningIcon } from '$lib/components/icons';
	import Spinner from '$lib/components/spinner.svelte';
	import { Tooltip } from '$lib/components/ui';
	import Button from '$lib/components/ui/button.svelte';
	import type { AssetModel, ChapteredAssets } from '$lib/models/asset-model';
	import type { CourseModel } from '$lib/models/course-model';
	import { cn, UpdateQueryParam } from '$lib/utils';
	import { toast } from 'svelte-sonner';

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

	async function updateAssetProgress(asset: AssetModel): Promise<void> {
		if (!course) return;
		try {
			await UpdateCourseAssetProgress(course.id, asset.id, asset.progress);
		} catch (error) {
			toast.error((error as APIError).message);
		}
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
			<!-- Side navigation -->
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
											{chapters[chapter].filter((a) => a.progress.completed).length}
											/ {chapters[chapter].length}
										</span>
									</div>
								</div>

								<div class="ml-auto flex flex-col gap-3 pt-4 pb-3">
									{#each chapters[chapter] as asset, index}
										<Button
											class={cn(
												'text-foreground-alt-2/80 bg-background enabled:hover:bg-background enabled:hover:text-foreground-alt-1 h-auto justify-start text-start duration-50 enabled:hover:underline',
												selectedAsset?.id === asset.id && 'text-foreground-alt-1'
											)}
											onclick={async () => {
												if (selectedAsset && selectedAsset.id === asset.id) return;
												await UpdateQueryParam('a', asset.id, false);
											}}
										>
											<div class="flex w-full flex-row items-center justify-between gap-2 pr-2.5">
												<span>
													{index + 1}. {asset.title}
												</span>
												<Button
													class={cn(
														' flex size-4 shrink-0 items-center justify-center rounded-full border',
														asset.progress.completed
															? 'enabled:bg-background-success enabled:hover:bg-background-success border-background-success '
															: 'enabled:bg-background enabled:hover:bg-background border-foreground'
													)}
													onclick={async (e: MouseEvent) => {
														e.stopPropagation();
														asset.progress.completed = !asset?.progress.completed;
														if (selectedAsset && selectedAsset.id === asset.id) {
															selectedAsset = asset;
														}
														await updateAssetProgress(asset);
													}}
												>
													<TickIcon class="text-foreground size-2 stroke-[3]" />
												</Button>
											</div>
										</Button>
									{/each}
								</div>
							</div>
						{/each}
					</nav>
				</div>
			</div>

			<!-- Main content -->
			<main class="container-pr flex w-full py-8">
				<div class="flex w-full place-content-center">
					<div class="flex w-full max-w-5xl flex-col gap-6 pt-1">
						<!-- Header -->
						<div class="flex w-full flex-col gap-8">
							{#if selectedAsset}
								<div class="flex flex-row items-center justify-between">
									<div class="flex w-full flex-row items-center gap-2">
										{#if selectedAsset.assetType === 'video'}
											<VideoIcon class="fill-foreground-alt-2 size-8 stroke-2" />
										{:else if selectedAsset.assetType === 'pdf'}
											<PdfIcon class="fill-foreground-alt-2 size-8 stroke-2" />
										{/if}
										<span class="text-xl font-medium">
											{selectedAsset.title}
										</span>
									</div>

									<!-- Mark watched/unwatched -->
									<Tooltip delayDuration={100} contentProps={{ side: 'bottom', sideOffset: 8 }}>
										{#snippet trigger()}
											<Button
												class={cn(
													' flex size-8 shrink-0 items-center justify-center rounded-full border',
													selectedAsset?.progress.completed
														? 'enabled:bg-background-success enabled:hover:bg-background-success border-background-success'
														: 'enabled:bg-background enabled:hover:bg-background border-foreground'
												)}
												onclick={async () => {
													if (!selectedAsset) return;
													selectedAsset.progress.completed = !selectedAsset?.progress.completed;
													await updateAssetProgress(selectedAsset);
												}}
											>
												<TickIcon class="text-foreground size-4 stroke-[3]" />
											</Button>
										{/snippet}

										{#snippet content()}
											Mark as {selectedAsset?.progress.completed ? 'unwatched' : 'watched'}
										{/snippet}
									</Tooltip>
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
