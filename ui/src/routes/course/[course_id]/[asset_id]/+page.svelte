<!-- TODO When page contains a group of assets, only allow 1 video to play -->
<!-- TODO find a way to show which assets are completed when page contains a group of assets -->
<!-- TODO rework description to support description type so we can render md vs txt -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { APIError } from '$lib/api-error.svelte';
	import {
		GetAllCourseAssets,
		GetCourse,
		ServeCourseAsset,
		ServeCourseAssetDescription,
		UpdateCourseAssetProgress
	} from '$lib/api/course-api';
	import { Spinner } from '$lib/components';
	import { BurgerMenuIcon, DotIcon, TickIcon, WarningIcon } from '$lib/components/icons';
	import { Button, Checkbox, Tooltip } from '$lib/components/ui';
	import Attachments from '$lib/components/ui/attachments.svelte';
	import { VideoPlayer } from '$lib/components/ui/media';
	import type { AssetGroup, AssetModel, Chapters } from '$lib/models/asset-model';
	import type { CourseModel } from '$lib/models/course-model';
	import { BuildChapterStructure, cn, renderMarkdown } from '$lib/utils';
	import { Dialog } from 'bits-ui';
	import prettyMs from 'pretty-ms';
	import { ElementSize } from 'runed';
	import { toast } from 'svelte-sonner';

	let course = $state<CourseModel>();
	let chapters = $state<Chapters>({});

	let selectedAssetGroup = $state<AssetGroup>();
	let selectedAsset = $state<AssetModel>();

	let renderedContent = $state<string>('');
	let renderedDescription = $state<string>();

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

			const result = findAssetGroup(page.params.asset_id, chapters);
			if (!result) {
				throw new Error('Asset not found');
			}

			selectedAssetGroup = result.group;
			selectedAsset = findAssetInGroup(page.params.asset_id, result.group);
			if (!selectedAsset) throw new Error('Asset not found');

			if (selectedAsset.assetType === 'markdown' || selectedAsset.assetType === 'text') {
				renderedContent = await setRenderedContent(selectedAsset);
			}

			renderedDescription = await setRenderedDescription(selectedAsset);

			initDone = true;
		} catch (error) {
			throw error;
		}
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

	// Set the rendered description for the asset
	async function setRenderedContent(asset: AssetModel): Promise<string> {
		if (!course || !asset || (asset.assetType !== 'markdown' && asset.assetType !== 'text')) {
			return '';
		}

		const content = await ServeCourseAsset(course.id, asset.id);
		if (!content) {
			return '';
		} else if (asset.assetType === 'text') {
			return content;
		} else {
			return renderMarkdown(content);
		}
	}
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Get the rendered description for the asset and render as markdown
	async function setRenderedDescription(asset: AssetModel): Promise<string> {
		if (!course || !asset || !asset.hasDescription) return '';

		const description = await ServeCourseAssetDescription(course.id, asset.id);
		if (!description) {
			return '';
		} else if (asset.descriptionType === 'text') {
			return description;
		} else {
			return renderMarkdown(description);
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	export function findAssetGroup(
		assetId: string,
		chapters: Chapters
	): { group: AssetGroup; chapter: string } | undefined {
		for (const [chapterName, assetGroups] of Object.entries(chapters)) {
			for (const group of assetGroups) {
				if (group.assets.some((asset) => asset.id === assetId)) {
					return { group, chapter: chapterName };
				}
			}
		}
		return undefined;
	}
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	export function findAssetInGroup(assetId: string, group: AssetGroup): AssetModel | undefined {
		return group.assets.find((asset) => asset.id === assetId);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update the selected asset when the page changes
	$effect(() => {
		const assetId = page.params.asset_id;
		if (!initDone || !selectedAssetGroup) return;

		// Ensure the body is scrolled to the top when the asset changes
		if (mainEl) mainEl.scrollTo({ top: 0, behavior: 'smooth' });

		selectedAsset = findAssetInGroup(assetId, selectedAssetGroup);

		// If asset not found in current group, need to find new group
		if (!selectedAsset) {
			const result = findAssetGroup(assetId, chapters);
			if (result) {
				selectedAssetGroup = result.group;
				selectedAsset = findAssetInGroup(assetId, result.group);
			}
		}

		if (!selectedAsset) throw new Error('Asset not found');

		if (selectedAsset.assetType === 'markdown' || selectedAsset.assetType === 'text') {
			setRenderedContent(selectedAsset).then((content) => {
				renderedContent = content;
			});
		}

		setRenderedDescription(selectedAsset).then((description) => {
			renderedDescription = description;
		});
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
				variant="ghost"
				class="bg-background hover:bg-background-alt-1 text-foreground-alt-1 hover:text-background-primary h-auto w-full items-start justify-start rounded-none py-5 pr-3 pl-0 text-start whitespace-normal"
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
							{chapters[chapter].filter((a) => a.completed).length}
							/ {chapters[chapter].length}
						</span>
					</div>
				</div>

				<div class="ml-auto flex flex-col pt-4 pb-3">
					{#each chapters[chapter] as assetGroup, index}
						<Button
							variant="ghost"
							class={cn(
								'text-foreground-alt-2 hover:bg-background-alt-2 hover:text-foreground-alt-1 h-auto w-full justify-start rounded-none text-start whitespace-normal',
								selectedAsset?.id === assetGroup.assets[0].id &&
									'text-foreground-alt-1 bg-background-alt-2'
							)}
							onclick={async () => {
								if (!course || assetGroup.assets[0].id === selectedAsset?.id) return;
								if (menuPopupMode) dialogOpen = false;
								goto(`/course/${course.id}/${assetGroup.assets[0].id}`, {});
							}}
						>
							<div class="flex w-full flex-row gap-3 py-2 pr-2.5 pl-1.5">
								<Checkbox
									class="hover:border-background-primary-alt-1 mt-px shrink-0 border-2"
									checked={assetGroup.completed}
									onclick={async (e: MouseEvent) => {
										e.stopPropagation();
										e.preventDefault();

										assetGroup.completed = !assetGroup.completed;

										assetGroup.assets.forEach((asset) => {
											if (!asset.progress) {
												asset.progress = {
													completed: true,
													completedAt: '',
													videoPos: 0
												};
											} else {
												asset.progress.completed = assetGroup?.completed || false;
											}
										});

										// promise await for all assets to update
										await Promise.all(assetGroup.assets.map((asset) => updateAssetProgress(asset)));
									}}
								/>

								<div class="flex w-full flex-col gap-2 text-sm">
									<span>
										{index + 1}. {assetGroup.title}
									</span>

									<div class="relative flex w-full flex-col gap-0 text-sm select-none">
										<!-- Attachments -->
										{#if assetGroup.attachments.length > 0}
											<div class="flex h-7 w-full flex-row flex-wrap items-center">
												<Attachments
													attachments={assetGroup.attachments}
													courseId={course?.id ?? ''}
													assetId={assetGroup.assets[0].id}
												/>
											</div>
										{/if}

										{#each assetGroup.assets as asset}
											<div class="flex w-full flex-row flex-wrap items-center">
												<DotIcon class="text-foreground-alt-3 mt-0.5 -ml-2.5 size-7" />

												<!-- Asset Title -->
												<span class="text-foreground-alt-3 whitespace-nowrap">
													{asset.assetType}
												</span>

												<!-- Video duration -->
												{#if asset.videoMetadata}
													<DotIcon class="text-foreground-alt-3 mt-0.5 size-7" />
													<span class="text-foreground-alt-3 whitespace-nowrap">
														{prettyMs(asset.videoMetadata.duration * 1000)}
													</span>
												{/if}
											</div>
										{/each}
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
	{#if course && selectedAsset && selectedAssetGroup}
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
						variant="ghost"
						class="text-foreground-alt-2 hover:text-foreground h-auto hover:bg-transparent"
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
							<div class="flex flex-row items-center justify-between">
								<div class="flex w-full flex-row items-center gap-2">
									<span class="text-xl font-medium">
										{selectedAssetGroup.title}
									</span>
								</div>

								<!-- Mark watched/unwatched -->
								<Tooltip delayDuration={100} contentProps={{ side: 'bottom', sideOffset: 8 }}>
									{#snippet trigger()}
										<Button
											variant="ghost"
											class={cn(
												'flex size-8 shrink-0 items-center justify-center rounded-full border',
												selectedAssetGroup?.completed
													? 'bg-background-success hover:bg-background-success border-background-success'
													: 'bg-background border-foreground'
											)}
											onclick={async () => {
												if (!selectedAssetGroup) return;

												selectedAssetGroup.completed = !selectedAssetGroup.completed;

												selectedAssetGroup.assets.forEach((asset) => {
													if (!asset.progress) {
														asset.progress = {
															completed: true,
															completedAt: '',
															videoPos: 0
														};
													} else {
														asset.progress.completed = selectedAssetGroup?.completed || false;
													}
												});

												// promise await for all assets to update
												await Promise.all(
													selectedAssetGroup.assets.map((asset) => updateAssetProgress(asset))
												);
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
						</div>

						<!-- Asset(s) -->
						{#each selectedAssetGroup.assets as asset}
							<div class="flex w-full flex-col gap-4">
								{#if selectedAssetGroup.assets.length > 1}
									<span class="text-lg font-medium">
										{asset.subPrefix}. {asset.subTitle ? asset.subTitle : asset.title}
									</span>
								{/if}

								{#if asset.assetType === 'video'}
									<VideoPlayer
										src={`/api/courses/${course.id}/assets/${asset.id}/serve`}
										startTime={asset.progress?.videoPos || 0}
										onTimeChange={(time: number) => {
											if (!asset.progress) {
												asset.progress = {
													completed: false,
													completedAt: '',
													videoPos: time
												};
											} else {
												asset.progress.videoPos = time;
											}

											updateAssetProgress(asset);
										}}
										onCompleted={(time: number) => {
											if (!asset.progress) {
												asset.progress = {
													completed: true,
													completedAt: '',
													videoPos: time
												};
											} else {
												asset.progress.videoPos = time;
												asset.progress.completed = true;
											}

											updateAssetProgress(asset);

											if (!selectedAssetGroup) return;
											selectedAssetGroup.completedAssetCount += 1;

											if (
												selectedAssetGroup.completedAssetCount >= selectedAssetGroup.assets.length
											)
												selectedAssetGroup.completed = true;
										}}
									/>
								{:else if asset.assetType === 'markdown'}
									<div class="typography">
										{@html renderedContent}
									</div>
								{:else if asset.assetType === 'text'}
									<div class="whitespace-pre-wrap">
										{renderedContent}
									</div>
								{/if}
							</div>
						{/each}

						<!-- Description -->
						{#if renderedDescription}
							<div
								class={cn(
									'typography',
									selectedAsset.descriptionType === 'text' && 'whitespace-pre-wrap'
								)}
							>
								<h3>Description</h3>
								{@html renderedDescription}
							</div>
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
