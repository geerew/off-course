<!-- TODO Add scan button (admin) -->
<!-- TODO Edit anything/everything (admin) -->
<!-- TODO Change asset play/restart button to a menu with play/restart, clear progress  -->
<!-- TODO Don't allow clicking the start button when in maintenance -->
<!-- TODO Hide the asset play button when in maintenance -->
<script lang="ts">
	import { page } from '$app/state';
	import { GetAllCourseAssets, GetCourse, GetCourseTags } from '$lib/api/course-api';
	import { auth } from '$lib/auth.svelte';
	import { NiceDate, Spinner } from '$lib/components';
	import { ClearCourseProgressDialog } from '$lib/components/dialogs';
	import {
		DotIcon,
		LogoIcon,
		MediaPlayIcon,
		RightChevronIcon,
		WarningIcon
	} from '$lib/components/icons';
	import MediaRestart from '$lib/components/icons/media-restart.svelte';
	import { Badge, Dialog } from '$lib/components/ui';
	import Attachments from '$lib/components/ui/attachments.svelte';
	import Button from '$lib/components/ui/button.svelte';
	import type { Chapters } from '$lib/models/asset-model';
	import type { CourseModel, CourseTagsModel } from '$lib/models/course-model';
	import { BuildChapterStructure, cn } from '$lib/utils';
	import { Accordion, Avatar, Progress, useId } from 'bits-ui';
	import prettyMs from 'pretty-ms';
	import { slide } from 'svelte/transition';

	let course = $state<CourseModel>();
	let chapters = $state<Chapters>({});
	let tags = $state<CourseTagsModel>([]);

	const labelId = useId();

	let chapterCount = $derived(Object.keys(chapters).length);
	let assetCount = $derived.by(() => {
		let count = 0;
		for (const chapter of Object.values(chapters)) {
			for (const assetGroup of chapter) {
				count += assetGroup.assets.length;
			}
		}
		return count;
	});
	let attachmentCount = $derived.by(() => {
		let count = 0;
		for (const chapter of Object.values(chapters)) {
			for (const assetGroup of chapter) {
				count += assetGroup.attachments.length;
			}
		}
		return count;
	});

	let assetToResume = $derived.by(() => {
		const allChapters = Object.values(chapters);

		// Find the first asset that is not completed
		for (const chapter of allChapters) {
			for (const assetGroup of chapter) {
				for (const asset of assetGroup.assets) {
					if (!asset.progress || !asset.progress.completed) {
						return asset;
					}
				}
			}
		}

		// If all assets are completed, return the first asset
		if (
			allChapters.length > 0 &&
			allChapters[0].length > 0 &&
			allChapters[0][0].assets.length > 0
		) {
			return allChapters[0][0].assets[0];
		}

		return null;
	});

	let loadPromise = $state(fetchCourse());

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch the course, then the assets for the course, then build a chapter structure from the
	// assets
	async function fetchCourse(): Promise<void> {
		try {
			course = await GetCourse(page.params.course_id);

			tags = await GetCourseTags(course.id);

			const assets = await GetAllCourseAssets(course.id, {
				q: `sort:"assets.chapter asc" sort:"assets.prefix asc"`
			});

			chapters = BuildChapterStructure(assets);
		} catch (error) {
			throw error;
		}
	}
</script>

{#await loadPromise}
	<div class="flex justify-center pt-10">
		<Spinner class="bg-foreground-alt-3 size-4" />
	</div>
{:then _}
	{#if course}
		<div class="flex w-full flex-col">
			<div class="bg-background-alt-1 flex w-full place-content-center">
				<div class="container-px flex w-full max-w-7xl flex-col gap-6 py-10">
					<div class="grid w-full grid-cols-1 gap-6 md:grid-cols-[minmax(0,22.6rem)_1fr] md:gap-10">
						<!-- Card -->
						<Avatar.Root
							class="relative z-0 flex h-full max-h-70 w-full items-center justify-center overflow-hidden rounded-lg [background-image:repeating-linear-gradient(-45deg,var(--color-background-alt-1),var(--color-background-alt-1)13px,var(--color-background-alt-2)13px,var(--color-background-alt-2)14px)] bg-[size:40px_40px]"
						>
							<Avatar.Image
								src={`/api/courses/${course.id}/card`}
								class="h-auto max-h-full w-auto max-w-full object-contain"
							/>

							<Avatar.Fallback
								class="bg-background-alt-2 flex h-50 w-full max-w-90 place-content-center rounded-lg"
							>
								<LogoIcon class="fill-background-alt-3 w-16 md:w-20" />
							</Avatar.Fallback>
						</Avatar.Root>

						<!-- Information -->
						<div class="flex h-full w-full flex-col gap-5">
							<div class="text-foreground-alt-1 text-2xl font-semibold">
								{course.title}
							</div>

							<div class="flex flex-col gap-1.5 text-sm">
								<!-- Path -->
								{#if auth.user?.role === 'admin'}
									<div class="grid grid-cols-[6.5rem_1fr]">
										<span class="text-foreground-alt-3 font-medium">PATH</span>
										<span
											class="text-foreground-alt-1 wrap-anywhere whitespace-normal"
											title={course.path}>{course.path}</span
										>
									</div>
								{/if}

								<!-- Created At -->
								<div class="grid grid-cols-[6.5rem_1fr]">
									<span class="text-foreground-alt-3 font-medium">ADDED</span>
									<span class="text-foreground-alt-1"><NiceDate date={course.createdAt} /></span>
								</div>

								<!-- Updated At -->
								<div class="grid grid-cols-[6.5rem_1fr]">
									<span class="text-foreground-alt-3 font-medium">UPDATED</span>
									<span class="text-foreground-alt-1"><NiceDate date={course.updatedAt} /></span>
								</div>

								<!-- Status -->
								{#if !course.available || course.maintenance || (course.initialScan !== undefined && !course.initialScan)}
									<div class="grid grid-cols-[6.5rem_1fr]">
										<span class="text-foreground-alt-3 font-medium">STATUS</span>

										{#if !course.initialScan}
											<span class="text-amber-600">Initial scan</span>
										{:else if course.maintenance}
											<span class="text-background-success">Maintenance</span>
										{:else}
											<span class="text-foreground-error">Unavailable</span>
										{/if}
									</div>
								{/if}

								<!-- Duration -->
								<div class="grid grid-cols-[6.5rem_1fr]">
									<span class="text-foreground-alt-3 font-medium">DURATION</span>
									<span class={course.duration ? 'text-foreground-alt-1' : 'text-foreground-alt-3'}>
										{course.duration
											? prettyMs(course.duration * 1000, { hideSeconds: true })
											: '-'}
									</span>
								</div>

								<!-- Progress -->
								<div class="grid grid-cols-[6.5rem_1fr]">
									<span class="text-foreground-alt-3 font-medium">PROGRESS</span>
									{#if course.progress}
										<div class="flex flex-row gap-2">
											<div class="relative flex w-30 items-center">
												<Progress.Root
													aria-labelledby={labelId}
													value={course.progress.percent}
													max={100}
													class="bg-background-alt-3 relative h-2.5 w-full overflow-hidden rounded-full"
												>
													<div
														class="bg-background-primary-alt-1 h-full w-full flex-1 rounded-full transition-all duration-1000 ease-in-out"
														style={`transform: translateX(-${100 - (100 * (course.progress.percent ?? 0)) / 100}%)`}
													></div>
												</Progress.Root>
											</div>
											<span id={labelId} class="text-foreground-alt-1 text-sm">
												{course.progress.percent}%
											</span>
										</div>
									{:else}
										<span class="text-foreground-alt-3">-</span>
									{/if}
								</div>
							</div>
						</div>
					</div>

					<!-- Tags -->
					<div class="flex flex-col gap-2.5">
						<span class="text-lg font-medium">Tags</span>
						<div class="flex flex-wrap gap-2">
							{#if tags.length === 0}
								<span class="text-foreground-alt-3">-</span>
							{:else}
								{#each tags as tag}
									<Badge class="text-sm  select-none">
										{tag.tag}
									</Badge>
								{/each}
							{/if}
						</div>
					</div>

					{#if assetCount > 0}
						<div class="flex flex-row gap-2.5">
							<Button
								href={`/course/${course.id}/${assetToResume?.id}`}
								class="w-auto px-4"
								disabled={course.maintenance || !course.available}
								onclick={(e) => {
									if (course?.maintenance || !course?.available) {
										e.preventDefault();
										e.stopPropagation();
									}
								}}
							>
								{#if course.progress}
									Resume
								{:else}
									Start
								{/if}
							</Button>

							<ClearCourseProgressDialog
								{course}
								successFn={() => {
									if (!course) return;
									course.progress = undefined;

									// Clear progress for all assets in all chapters
									const allChapters = Object.values(chapters);
									for (const chapter of allChapters) {
										for (const assetGroup of chapter) {
											assetGroup.completed = false;
											assetGroup.startedAssetCount = 0;
											assetGroup.completedAssetCount = 0;
											assetGroup.assets.forEach((asset) => {
												asset.progress = undefined;
											});
										}
									}
								}}
							>
								{#snippet trigger()}
									<Dialog.Trigger class="w-auto px-4" disabled={!course?.progress}>
										Clear Progress
									</Dialog.Trigger>
								{/snippet}
							</ClearCourseProgressDialog>
						</div>
					{/if}
				</div>
			</div>

			<!-- Course Content -->
			<div class="bg-background flex w-full place-content-center">
				<div class="container-px flex w-full max-w-7xl flex-col py-7">
					<div class="flex flex-col gap-5">
						<div class="flex flex-col gap-2.5">
							<div class="flex flex-col gap-1.5">
								<span class="text-xl font-medium">Course Content</span>
								<span class="text-foreground-alt-3 flex items-center text-sm font-medium">
									{chapterCount} chapter{chapterCount != 1 ? 's' : ''}
									<DotIcon class="text-foreground-alt-3 mt-0.5 size-7" />
									{assetCount} asset{assetCount != 1 ? 's' : ''}
									<DotIcon class="text-foreground-alt-3 mt-0.5 size-7" />
									{attachmentCount} attachment{attachmentCount != 1 ? 's' : ''}
								</span>
							</div>
						</div>

						<Accordion.Root class="w-full" type="multiple" value={[Object.keys(chapters)[0]]}>
							{#each Object.keys(chapters) as chapter}
								<Accordion.Item
									value={chapter}
									class="bg-background-alt-1 border-background last:border-background-alt-2 overflow-hidden border-b duration-150 first:rounded-t-lg last:rounded-b-lg"
								>
									<Accordion.Header>
										<Accordion.Trigger
											class="group/trigger hover:bg-background-alt-2 flex w-full flex-1 items-center justify-between p-5 font-medium transition-all select-none hover:cursor-pointer"
										>
											<span class="w-full text-left">
												{chapter}
											</span>

											<div class="flex shrink-0 flex-row items-center gap-2.5">
												<span class="text-foreground-alt-3 text-xs">
													{chapters[chapter].reduce(
														(acc, assetGroup) => acc + assetGroup.completedAssetCount,
														0
													)}
													/ {chapters[chapter].reduce(
														(acc, assetGroup) => acc + assetGroup.assets.length,
														0
													)}
												</span>
												<RightChevronIcon
													class="size-[18px] rotate-90 stroke-2 transition-transform duration-200 group-data-[state=open]/trigger:-rotate-90"
												/>
											</div>
										</Accordion.Trigger>
									</Accordion.Header>

									<Accordion.Content
										forceMount={true}
										class="bg-background border-background-alt-2 flex flex-col border-x"
									>
										{#snippet child({ props, open })}
											{#if open}
												<div {...props} transition:slide={{ duration: 200 }}>
													{#each chapters[chapter] as assetGroup}
														<div
															class="border-background-alt-2 text-foreground-alt-1 group relative flex flex-row items-center justify-between gap-2 overflow-hidden rounded-none border-b px-5 py-2 last:border-none"
														>
															{#if assetGroup.completed || assetGroup.startedAssetCount > 0}
																<div
																	class={cn(
																		'absolute top-1/2 left-1 inline-block h-[calc(100%-30px)] w-0.5 -translate-y-1/2 opacity-60',
																		assetGroup.completed
																			? 'bg-background-success'
																			: assetGroup.startedAssetCount > 0
																				? 'bg-amber-600'
																				: ''
																	)}
																></div>
															{/if}

															<div class="flex w-full flex-col gap-2 py-2 text-sm">
																<span class="w-full">{assetGroup.prefix}. {assetGroup.title}</span>

																<!-- Main metadata row -->
																<div
																	class="relative flex w-full flex-col gap-0 text-sm select-none"
																>
																	<!-- Attachments -->
																	{#if assetGroup.attachments.length > 0}
																		<div
																			class="flex h-7 w-full flex-row flex-wrap items-center pl-2.5"
																		>
																			<Attachments
																				attachments={assetGroup.attachments}
																				courseId={course?.id ?? ''}
																				assetId={assetGroup.assets[0].id}
																			/>
																		</div>
																	{/if}

																	{#each assetGroup.assets as asset}
																		<div class="flex w-full flex-row flex-wrap items-center">
																			<DotIcon class="text-foreground-alt-3 mt-0.5 size-7" />

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

															<!-- Play button -->
															<Button
																href={`/course/${course?.id}/${assetGroup.assets[0].id}`}
																class={cn(
																	'bg-background-alt-2  hover:bg-background-alt-3 flex h-auto w-auto shrink-0 items-center justify-center rounded-full p-2 opacity-0 transition-all duration-150 ease-in',
																	course?.maintenance || !course?.available
																		? 'group-hover:opacity-0 pointer-coarse:opacity-0'
																		: 'group-hover:opacity-100 pointer-coarse:opacity-100'
																)}
															>
																{#if assetGroup.completed}
																	<MediaRestart
																		class="stroke-foreground-alt-1 size-5.5 fill-transparent stroke-[1.5] pointer-coarse:size-4"
																	/>
																{:else}
																	<MediaPlayIcon
																		class="fill-foreground-alt-1 size-5.5 pointer-coarse:size-4"
																	/>
																{/if}
															</Button>
														</div>
													{/each}
												</div>
											{/if}
										{/snippet}
									</Accordion.Content>
								</Accordion.Item>
							{/each}
						</Accordion.Root>

						<div class="flex flex-row gap-3 text-sm">
							<span>Asset Status:</span>
							<div class="flex flex-row gap-3">
								<div class="flex flex-row items-center gap-2">
									<div class="bg-background-success mt-px size-4 rounded-md"></div>
									<span>Completed</span>
								</div>
								<div class="flex flex-row items-center gap-2">
									<div class="mt-px size-4 rounded-md bg-amber-600"></div>
									<span>In-progress</span>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
{:catch error}
	<div class="flex w-full flex-col items-center gap-2 pt-10">
		<WarningIcon class="text-foreground-error size-10" />
		<span class="text-lg">Failed to load page</span>
		<span class="text-foreground-alt-3 text-sm">{error.message}</span>
	</div>
{/await}
