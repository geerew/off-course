<!-- TODO fix bug when video is at the end, I see an parsing data error -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { APIError } from '$lib/api-error.svelte';
	import { GetAllCourseAssets, GetCourse, UpdateCourseAssetProgress } from '$lib/api/course-api';
	import { Spinner } from '$lib/components';
	import { BurgerMenuIcon, DotIcon, TickIcon, WarningIcon } from '$lib/components/icons';
	import { Button, Checkbox, Tooltip } from '$lib/components/ui';
	import Attachments from '$lib/components/ui/attachments.svelte';
	import { VideoPlayer } from '$lib/components/ui/media';
	import type { AssetModel, ChapteredAssets } from '$lib/models/asset-model';
	import type { CourseModel } from '$lib/models/course-model';
	import { cn } from '$lib/utils';
	import { Dialog } from 'bits-ui';
	import prettyMs from 'pretty-ms';
	import { ElementSize } from 'runed';
	import { toast } from 'svelte-sonner';

	let course = $state<CourseModel>();
	let chapters = $state<ChapteredAssets>({});
	let selectedAsset = $state<AssetModel>();

	let loadPromise = $state(fetchCourseAndAsset());
	let initDone = false;

	let mainEl = $state() as HTMLElement;
	const mainSize = new ElementSize(() => mainEl);

	let menuPopupMode = $state(false);
	let dialogOpen = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch the course, then the assets for the course, then build a chapter structure from the
	// assets
	async function fetchCourseAndAsset(): Promise<void> {
		try {
			course = await GetCourse(page.params.course_id);

			const assets = await GetAllCourseAssets(course.id, {
				q: `sort:"assets.chapter asc" sort:"assets.prefix asc"`
			});

			chapters = BuildChapterStructure(assets);

			selectedAsset = findAsset(page.params.asset_id, chapters);

			if (!selectedAsset) {
				throw new Error('Asset not found');
			}

			initDone = true;
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

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update the asset progress in the database
	async function updateAssetProgress(asset: AssetModel): Promise<void> {
		if (!course || !asset.progress) return;
		try {
			await UpdateCourseAssetProgress(course.id, asset.id, asset.progress);
		} catch (error) {
			toast.error((error as APIError).message);
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Find the asset in the chapters
	function findAsset(assetId: string, chapters: ChapteredAssets): AssetModel | undefined {
		for (const chapter of Object.values(chapters)) {
			const found = chapter.find((asset) => asset.id === assetId);
			if (found) {
				return found;
			}
		}

		return undefined;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update the selected asset when the page changes
	$effect(() => {
		const assetId = page.params.asset_id;
		if (!initDone) return;
		selectedAsset = findAsset(assetId, chapters);
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// As the main content of the page resizes, control whether we show the popup menu or the
	// static menu
	$effect(() => {
		const width = mainSize.width;
		if (width < 700) {
			dialogOpen = false;
			menuPopupMode = true;
		} else if (width > 1100) {
			menuPopupMode = false;
		}
	});
</script>

{#snippet menuContents()}
	{#if course}
		<div
			class="bg-background border-background-alt-4 sticky top-0 z-[1] flex flex-row gap-3 border-b"
		>
			<Button
				href={`/course/${course.id}`}
				class={cn(
					'bg-background hover:bg-background-alt-1 text-foreground-alt-1 hover:text-background-primary flex h-auto items-start justify-start rounded-none py-5 pr-3 text-start duration-200'
				)}
			>
				<span class="container-pl font-semibold">{course.title}</span>
			</Button>
		</div>

		{#each Object.keys(chapters) as chapter}
			<div class="container-pl leading-5">
				<div class="border-background-alt-6 flex flex-col gap-1.5 border-b py-1.5 pr-2">
					<span class=" text-background-primary text-sm font-semibold tracking-wide">
						{chapter}
					</span>
					<div class="flex flex-row items-center gap-1">
						<span class="text-foreground-alt-3 text-xs">
							{chapters[chapter].filter((a) => a.progress?.completed).length}
							/ {chapters[chapter].length}
						</span>
					</div>
				</div>

				<div class="ml-auto flex flex-col pt-4 pb-3">
					{#each chapters[chapter] as asset, index}
						<Button
							class={cn(
								'text-foreground-alt-2 bg-background enabled:hover:bg-background-alt-2 enabled:hover:text-foreground-alt-1 h-auto justify-start rounded-none text-start duration-50',
								selectedAsset?.id === asset.id && 'text-foreground-alt-1 bg-background-alt-2'
							)}
							onclick={() => {
								console.log('in here');
								if (!course || asset.id === selectedAsset?.id) return;
								if (menuPopupMode) dialogOpen = false;
								goto(`/course/${course.id}/${asset.id}`, {});
								selectedAsset = asset;
							}}
						>
							<div class="flex w-full flex-row gap-3 py-2 pr-2.5 pl-1.5">
								<Checkbox
									class="hover:border-background-primary-alt-1 mt-px shrink-0 border-2"
									checked={asset.progress?.completed}
									onclick={async (e: MouseEvent) => {
										e.stopPropagation();
										e.preventDefault();

										if (!asset.progress) {
											asset.progress = {
												completed: true,
												completedAt: '',
												videoPos: 0
											};
										} else {
											asset.progress.completed = !asset.progress.completed;
										}

										await updateAssetProgress(asset);
									}}
								/>

								<div class="flex w-full flex-col gap-2.5 text-sm">
									<span>
										{index + 1}. {asset.title}
									</span>

									<div class="flex w-full flex-row items-center">
										<span class="text-foreground-alt-3">{asset.assetType}</span>

										{#if asset.videoMetadata}
											<DotIcon class="text-foreground-alt-3 mt-0.5 size-7" />
											<span class="text-foreground-alt-3">
												{prettyMs(asset.videoMetadata.duration * 1000)}
											</span>
										{/if}

										{#if asset.attachments.length > 0}
											<DotIcon class="text-foreground-alt-3 mt-0.5 size-7" />
											<Attachments
												attachments={asset.attachments}
												courseId={course?.id ?? ''}
												assetId={asset.id}
											/>
										{/if}
									</div>
								</div>
							</div>
						</Button>
					{/each}
				</div>
			</div>
		{/each}
	{/if}
{/snippet}

{#await loadPromise}
	<div class="flex justify-center pt-10">
		<Spinner class="bg-foreground-alt-3 size-4" />
	</div>
{:then _}
	{#if course && selectedAsset}
		<div
			class={cn(
				'grid grid-rows-1 gap-6 pt-[calc(var(--header-height)+1))]',
				menuPopupMode ? 'grid-cols-1' : 'grid-cols-[var(--course-menu-width)_1fr]'
			)}
		>
			<!-- Menu -->
			{#if menuPopupMode}
				<Dialog.Root bind:open={dialogOpen}>
					<Dialog.Portal>
						<Dialog.Overlay
							class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-40 bg-black/30"
						/>

						<Dialog.Content
							class="border-foreground-alt-4 bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left fixed top-0 left-0 z-50 h-full w-[var(--course-menu-width)] border-r"
						>
							<nav class="flex h-full w-full flex-col gap-2 overflow-x-hidden overflow-y-auto pb-8">
								{@render menuContents()}
							</nav>
						</Dialog.Content>
					</Dialog.Portal>
				</Dialog.Root>
			{:else}
				<div class="relative row-span-full">
					<div class="absolute inset-0">
						<nav
							class="border-foreground-alt-4 sticky top-[calc(var(--header-height)+1px)] left-0 flex h-[calc(100dvh-(var(--header-height)+1px))] w-[--course-menu-width] flex-col gap-2 overflow-x-hidden overflow-y-auto border-r pb-8"
						>
							{@render menuContents()}
						</nav>
					</div>
				</div>
			{/if}

			<!-- Dialog trigger -->
			<div
				class={cn(
					'border-background-alt-3 flex h-12 border-b',
					menuPopupMode ? 'visible' : 'hidden'
				)}
			>
				<div class="container-pl flex h-full items-center">
					<Button
						class="bg-background text-foreground-alt-2 hover:text-foreground-alt-1 flex h-auto items-start justify-start gap-1.5 text-start duration-200 enabled:hover:bg-transparent"
						onclick={() => {
							dialogOpen = !dialogOpen;
						}}
					>
						<BurgerMenuIcon class="size-6 stroke-[1.5]" />
						<span>Menu</span>
					</Button>
				</div>
			</div>

			<!-- Main content -->
			<main
				class={cn('container-px flex w-full pb-8', !menuPopupMode && 'pt-8')}
				bind:this={mainEl}
			>
				<div class="flex w-full place-content-center">
					<div class="flex w-full max-w-5xl flex-col gap-6 pt-1">
						<!-- Header -->
						<div class="flex w-full flex-col gap-8">
							{#if selectedAsset}
								<div class="flex flex-row items-center justify-between">
									<div class="flex w-full flex-row items-center gap-2">
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
													selectedAsset?.progress?.completed
														? 'enabled:bg-background-success enabled:hover:bg-background-success border-background-success'
														: 'enabled:bg-background enabled:hover:bg-background border-foreground'
												)}
												onclick={async () => {
													if (!selectedAsset) return;

													if (!selectedAsset.progress) {
														selectedAsset.progress = {
															completed: true,
															completedAt: '',
															videoPos: 0
														};
													} else {
														selectedAsset.progress.completed = !selectedAsset.progress.completed;
													}

													await updateAssetProgress(selectedAsset);
												}}
											>
												<TickIcon class="text-foreground size-4 stroke-[3]" />
											</Button>
										{/snippet}

										{#snippet content()}
											Mark as {selectedAsset?.progress?.completed ? 'unwatched' : 'watched'}
										{/snippet}
									</Tooltip>
								</div>
							{/if}
						</div>

						{#if selectedAsset}
							{#if selectedAsset.assetType === 'video'}
								<VideoPlayer
									src={`/api/courses/${course.id}/assets/${selectedAsset.id}/serve`}
									startTime={selectedAsset.progress?.videoPos || 0}
									onTimeChange={(time: number) => {
										if (!selectedAsset) return;

										if (!selectedAsset.progress) {
											selectedAsset.progress = {
												completed: false,
												completedAt: '',
												videoPos: time
											};
										} else {
											selectedAsset.progress.videoPos = time;
										}

										updateAssetProgress(selectedAsset);
									}}
									onCompleted={(time: number) => {
										if (!selectedAsset) return;
										if (!selectedAsset.progress) {
											selectedAsset.progress = {
												completed: true,
												completedAt: '',
												videoPos: time
											};
										} else {
											selectedAsset.progress.videoPos = time;
											selectedAsset.progress.completed = true;
										}

										updateAssetProgress(selectedAsset);
									}}
								/>
							{/if}
						{/if}
					</div>
				</div>
			</main>
		</div>
	{/if}
{:catch error}
	<div class="flex w-full flex-col items-center gap-2 pt-10">
		<WarningIcon class="text-foreground-error size-10" />
		<span class="text-lg">Failed to load page</span>
		<span class="text-foreground-alt-3 text-sm">{error.message}</span>
	</div>
{/await}
