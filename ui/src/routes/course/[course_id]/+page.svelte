<!-- TODO show attachments dropdown -->
<!-- TODO show a progress bar -->
<script lang="ts">
	import { page } from '$app/state';
	import { GetAllCourseAssets, GetCourse, GetCourseTags } from '$lib/api/course-api';
	import { auth } from '$lib/auth.svelte';
	import { NiceDate, Spinner } from '$lib/components';
	import {
		DotIcon,
		LogoIcon,
		MediaPlayIcon,
		RightChevronIcon,
		WarningIcon
	} from '$lib/components/icons';
	import { Badge, Button } from '$lib/components/ui';
	import type { AssetModel, ChapteredAssets } from '$lib/models/asset-model';
	import type { CourseModel, CourseTagsModel } from '$lib/models/course-model';
	import { Accordion, Avatar } from 'bits-ui';
	import prettyMs from 'pretty-ms';
	import { slide } from 'svelte/transition';

	let course = $state<CourseModel>();
	let chapters = $state<ChapteredAssets>({});
	let tags = $state<CourseTagsModel>([]);

	let chapterCount = $derived(Object.keys(chapters).length);
	let assetCount = $derived.by(() => {
		let count = 0;
		for (const chapterAssets of Object.values(chapters)) {
			count += chapterAssets.length;
		}
		return count;
	});
	let attachmentCount = $derived.by(() => {
		let count = 0;
		for (const chapterAssets of Object.values(chapters)) {
			for (const asset of chapterAssets) {
				count += asset.attachments.length;
			}
		}
		return count;
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
					<div class="grid-col-1 grid w-full gap-10 md:grid-cols-[22.6rem_1fr]">
						<!-- Card -->
						<Avatar.Root
							class="relative z-0 flex h-full max-h-50 w-full items-center justify-center"
						>
							<Avatar.Image
								src={`/api/courses/${course.id}/card`}
								class="h-auto max-h-full w-auto max-w-full object-contain"
							/>

							<Avatar.Fallback
								class="bg-background-alt-2 flex h-50 w-full place-content-center rounded-lg"
							>
								<LogoIcon class="fill-background-alt-3 w-20" />
							</Avatar.Fallback>

							<div
								class="absolute inset-0 -z-10 h-full w-full bg-[radial-gradient(circle,#73737350_1px,transparent_1px)] bg-[size:11px_11px]"
							></div>
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
										<span class="text-foreground-alt-1">{course.path} </span>
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

								<!-- Duration -->
								<div class="grid grid-cols-[6.5rem_1fr]">
									<span class="text-foreground-alt-3 font-medium">DURATION</span>
									<span class="text-foreground-alt-1">
										{course.duration
											? prettyMs(course.duration * 1000, { hideSeconds: true })
											: '-'}
									</span>
								</div>

								<!-- Progress -->
								{#if course.progress.started}
									<div class="grid grid-cols-[6.5rem_1fr]">
										<span class="text-foreground-alt-3 font-medium">PROGRESS</span>
										<!-- TODO Make a progress bar -->
										<span class="text-foreground-alt-1">{course.progress.percent}%</span>
									</div>
								{/if}
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
				</div>
			</div>

			<!-- Course Content -->
			<div class="bg-background flex w-full place-content-center">
				<div class="container-px flex w-full max-w-7xl flex-col py-7">
					<div class="flex flex-col gap-5">
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

						<Accordion.Root class="w-full" type="single" value={Object.keys(chapters)[0]}>
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
												<span class="text-foreground-alt-3"
													>{chapters[chapter].length} asset{chapters[chapter].length > 1
														? 's'
														: ''}</span
												>
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
													{#each chapters[chapter] as asset}
														<Button
															href={`/course/${course?.id}/${asset.id}`}
															class="border-background-alt-2 bg-background enabled:hover:bg-background-alt-1/60 text-foreground-alt-1 group flex h-auto flex-row items-center justify-between gap-2 rounded-none border-b px-5 py-2 last:border-none "
														>
															<div class="flex w-full flex-col items-center py-2 text-sm">
																<span class="w-full text-left">{asset.prefix}. {asset.title}</span>
																<div class="flex w-full flex-row items-center">
																	<span class="text-foreground-alt-3">{asset.assetType}</span>
																	{#if asset.videoMetadata}
																		<DotIcon class="text-foreground-alt-3 mt-0.5 size-7" />
																		<span class="text-foreground-alt-3">
																			{prettyMs(asset.videoMetadata.duration * 1000)}
																		</span>
																	{/if}
																	{#if asset.progress.completed}
																		<DotIcon class="text-foreground-alt-3 mt-0.5 size-7" />
																		<span class="text-background-success">completed</span>
																	{:else if asset.progress.videoPos > 0}
																		<DotIcon class="text-foreground-alt-3 mt-0.5 size-7" />
																		<span class="text-amber-600">in-progress</span>
																	{/if}

																	<!-- TODO TEMP -->
																	{#if asset.attachments.length > 0}
																		<DotIcon class="text-foreground-alt-3 mt-0.5 size-7" />
																		<span class="text-foreground-alt-3">
																			{asset.attachments.length} attachment
																			{asset.attachments.length > 1 ? 's' : ''}
																		</span>
																	{/if}
																</div>
															</div>

															<div
																class="bg-background-alt-2 flex items-center justify-center rounded-full p-2 opacity-0 transition-opacity duration-150 ease-in group-hover:opacity-100"
															>
																<MediaPlayIcon class="fill-foreground-alt-1 size-3.5" />
															</div>
														</Button>
													{/each}
												</div>
											{/if}
										{/snippet}
									</Accordion.Content>
								</Accordion.Item>
							{/each}
						</Accordion.Root>
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
