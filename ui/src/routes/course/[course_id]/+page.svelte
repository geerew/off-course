<!-- TODO Add scan button (admin) -->
<!-- TODO Edit anything/everything (admin) -->
<!-- TODO Change asset play/restart button to a menu with play/restart, clear progress  -->
<!-- TODO Don't allow clicking the start button when in maintenance -->
<!-- TODO Hide the asset play button when in maintenance -->
<!-- TODO Add mark Complete -->
<script lang="ts">
	import { page } from '$app/state';
	import { GetAllCourseAssets, GetCourse, GetCourseTags } from '$lib/api/course-api';
	import { auth } from '$lib/auth.svelte';
	import { NiceDate, Spinner } from '$lib/components';
	import { ClearCourseProgressDialog } from '$lib/components/dialogs';
	import {
		AddedIcon,
		ClearProgressIcon,
		DotIcon,
		DotsIcon,
		DurationIcon,
		EllipsisCircleIcon,
		FilesIcon,
		LogoIcon,
		ModulesIcon,
		PathIcon,
		PlayCircleIcon,
		TickCircleIcon,
		UpdatedIcon,
		WarningIcon
	} from '$lib/components/icons';
	import { Badge, Dropdown } from '$lib/components/ui';
	import Attachments from '$lib/components/ui/attachments.svelte';
	import Button from '$lib/components/ui/button.svelte';
	import type { Chapters } from '$lib/models/asset-model';
	import type { CourseModel, CourseTagsModel } from '$lib/models/course-model';
	import { BuildChapterStructure } from '$lib/utils';
	import { useId } from 'bits-ui';
	import prettyMs from 'pretty-ms';

	let course = $state<CourseModel>();
	let chapters = $state<Chapters>({});
	let tags = $state<CourseTagsModel>([]);
	let courseImageUrl = $state<string | null>(null);
	let courseImageLoaded = $state<boolean>(false);

	let openCourseProgressDialog = $state(false);

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
	let assetGroupCount = $derived.by(() => {
		let count = 0;
		for (const chapter of Object.values(chapters)) {
			count += chapter.length;
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

			await loadCourseImage(course.id);
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function loadCourseImage(courseId: string): Promise<void> {
		try {
			const response = await fetch(`/api/courses/${courseId}/card`);
			if (response.ok) {
				const blob = await response.blob();
				courseImageUrl = URL.createObjectURL(blob);
				courseImageLoaded = true;
			} else {
				courseImageLoaded = false;
			}
		} catch (error) {
			courseImageLoaded = false;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	const pad2 = (n: number) => String(n).padStart(2, '0');

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		return () => {
			if (courseImageUrl) {
				URL.revokeObjectURL(courseImageUrl);
			}
		};
	});
</script>

{#await loadPromise}
	<div class="flex justify-center pt-10">
		<Spinner class="bg-foreground-alt-3 size-4" />
	</div>
{:then _}
	{#if course}
		<div class="flex w-full flex-col">
			<div class="flex w-full place-content-center">
				<div class="container-px flex w-full max-w-7xl flex-col gap-6 py-10">
					<div class="grid w-full grid-cols-1 gap-6 lg:grid-cols-[1fr_minmax(0,23rem)] lg:gap-10">
						<!-- Information -->
						<div class="order-2 flex h-full w-full flex-col justify-between gap-5 lg:order-1">
							<div class="flex h-full w-full flex-col gap-5">
								<div class="text-foreground-alt-1 text-2xl font-semibold">
									{course.title}
								</div>

								<!-- Course overview -->
								<div class="flex flex-col gap-5 text-sm">
									<div class="flex flex-row items-center gap-3">
										<!-- Modules -->
										<div class="flex flex-row items-center gap-2 font-semibold">
											<ModulesIcon class="text-foreground-alt-3 size-4.5" />
											<span>
												{chapterCount} module{chapterCount != 1 ? 's' : ''}
											</span>
										</div>

										<DotIcon class="text-foreground-alt-3 text-xl" />

										<!-- Assets -->
										<div class="flex flex-row items-center gap-2 font-semibold">
											<FilesIcon class="text-foreground-alt-3 size-4.5" />
											<span>
												{assetGroupCount} lesson{assetGroupCount != 1 ? 's' : ''}
											</span>
										</div>

										<DotIcon class="text-foreground-alt-3 text-xl" />

										<!-- Duration -->
										<div class="flex flex-row items-center gap-2 font-semibold">
											<DurationIcon class="text-foreground-alt-3 size-4.5" />
											<span
												class={course.duration ? 'text-foreground-alt-1' : 'text-foreground-alt-3'}
											>
												{course.duration
													? prettyMs(course.duration * 1000, { hideSeconds: true })
													: '-'}
											</span>
										</div>
									</div>
								</div>

								<!-- Path, Created At, Updated At, Status -->
								<div class="flex flex-col gap-2 text-sm">
									{#if auth.user?.role === 'admin'}
										<div class="text-foreground-alt-2 flex flex-row items-start gap-2">
											<PathIcon class="text-foreground-alt-3 size-4.5 shrink-0" />
											<span class="wrap-anywhere whitespace-normal" title={course.path}
												>{course.path}</span
											>
										</div>
									{/if}

									<div class="text-foreground-alt-2 flex flex-row items-center gap-3">
										<!-- Added -->
										<div class="flex flex-row items-center gap-2">
											<AddedIcon class="text-foreground-alt-3 size-4.5" />
											<span>
												<NiceDate date={course.createdAt} prefix="Added:" />
											</span>
										</div>

										<DotIcon class="text-xl" />

										<!-- Updated -->
										<div class="flex flex-row items-center gap-2">
											<UpdatedIcon class="text-foreground-alt-3 size-4.5" />
											<span>
												<NiceDate date={course.updatedAt} prefix="Updated:" />
											</span>
										</div>

										{#if !course.available || course.maintenance || (course.initialScan !== undefined && !course.initialScan)}
											<DotIcon class="text-xl" />

											<div class="flex flex-row items-center gap-2">
												{#if !course.initialScan}
													<span class="text-amber-600">Initial scan</span>
												{:else if course.maintenance}
													<span class="text-background-success">Maintenance</span>
												{:else}
													<span class="text-foreground-error">Unavailable</span>
												{/if}
											</div>
										{/if}
									</div>
								</div>

								<!-- Tags -->
								<div class="flex flex-wrap gap-2">
									{#each tags as tag}
										<Badge class="text-sm  select-none">
											{tag.tag}
										</Badge>
									{/each}
								</div>
							</div>

							{#if assetCount > 0}
								<div class="flex flex-row place-items-end gap-2.5">
									<Button
										href={`/course/${course.id}/${assetToResume?.id}`}
										variant="default"
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
											Resume Course
										{:else}
											Start Course
										{/if}
									</Button>

									<Dropdown.Root>
										<Dropdown.Trigger
											class="bg-background-alt-3 data-[state=open]:bg-background-alt-4 hover:bg-background-alt-4 w-auto rounded-lg border-none"
										>
											<DotsIcon class="size-5 stroke-[1.5]" />
										</Dropdown.Trigger>

										<Dropdown.Content class="z-60 w-38">
											<Dropdown.Item
												class="data-disabled:pointer-events-none"
												disabled={!course?.progress}
												onclick={async () => {
													openCourseProgressDialog = true;
												}}
											>
												<ClearProgressIcon class="size-4 stroke-[1.5]" />
												<span>Clear Progress</span>
											</Dropdown.Item>
										</Dropdown.Content>
									</Dropdown.Root>

									<ClearCourseProgressDialog
										bind:open={openCourseProgressDialog}
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
									/>
								</div>
							{/if}
						</div>

						<!-- Card -->
						<div class="relative order-1 flex w-full overflow-hidden rounded-lg lg:order-2">
							<!-- Image Container -->
							<div
								class="relative flex h-full max-h-70 w-full items-center justify-center overflow-hidden [background-image:repeating-linear-gradient(-45deg,var(--color-background),var(--color-background)13px,var(--color-background-alt-1)13px,var(--color-background-alt-1)14px)] bg-[size:40px_40px]"
							>
								{#if courseImageLoaded && courseImageUrl}
									<img
										src={courseImageUrl}
										alt={course?.title}
										class="h-auto max-h-full w-auto max-w-full rounded-lg object-contain"
									/>
								{:else}
									<!-- Fallback -->
									<div
										class="bg-background-alt-2 flex h-50 w-full max-w-90 items-center justify-center rounded-lg"
									>
										<LogoIcon class="fill-background-alt-3 w-16 md:w-20" />
									</div>
								{/if}

								<!-- Progress Bar Overlay -->
								{#if course?.progress}
									<div class="absolute right-0 bottom-0 left-0 w-full">
										<div
											class="bg-background-alt-3/80 relative h-5 w-full overflow-hidden backdrop-blur-sm"
											aria-labelledby={labelId}
											role="progressbar"
											aria-valuenow={course.progress.percent}
											aria-valuemin="0"
											aria-valuemax="100"
										>
											<div
												class="bg-background-primary-alt-1/70 h-full transition-all duration-1000 ease-in-out"
												style={`width: ${course.progress.percent}%`}
											></div>

											<!-- Progress Text Inside Bar -->
											<div
												id={labelId}
												class="text-foreground-alt-1 absolute inset-0 flex items-center justify-center text-xs font-medium drop-shadow-sm"
											>
												{course.progress.percent}%
											</div>
										</div>
									</div>
								{/if}
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Course Content -->
			<div class="bg-background flex w-full place-content-center">
				<div class="container-px flex w-full max-w-7xl flex-col py-7">
					<div class="text-foreground-alt-1 flex flex-col gap-16">
						{#each Object.keys(chapters) as chapter, index}
							<section class="border-background-alt-2 grid grid-cols-4 border-t">
								<div class="col-span-full sm:col-span-1">
									<div class="border-foreground-alt-2 -mt-px inline-flex border-t pt-px">
										<div class="text-background-primary-alt-1 pt-6 font-semibold sm:pt-10">
											Module {pad2(index + 1)}
										</div>
									</div>
								</div>

								<div class="col-span-full pt-6 sm:col-span-3 sm:pt-10">
									<div class="max-w-2xl">
										<!-- Module title -->
										<div class="text-2xl font-medium text-pretty">
											{chapter}
										</div>

										<!-- Module details -->
										<ol class="mt-10 space-y-6">
											{#each chapters[chapter] as assetGroup}
												{@const isCollection = assetGroup.assets.length > 1}
												{@const totalVideoDuration = assetGroup.assets.reduce(
													(acc, asset) => acc + (asset.videoMetadata?.duration || 0),
													0
												)}

												<li>
													<div class="flow-root">
														<Button
															href={`/course/${course.id}/${assetGroup.assets[0].id}`}
															variant="ghost"
															class="hover:bg-background-alt-2 -mx-3 -my-2 flex h-auto justify-start gap-3 py-2 text-sm whitespace-normal"
															disabled={course.maintenance || !course.available}
															onclick={(e) => {
																if (course?.maintenance || !course?.available) {
																	e.preventDefault();
																	e.stopPropagation();
																}
															}}
														>
															<!-- Lesson status -->
															{#if assetGroup.completed}
																<TickCircleIcon
																	class="stroke-background-success fill-background-success [&_path]:stroke-foreground size-5 place-self-start stroke-1 [&_path]:stroke-1"
																/>
															{:else if assetGroup.startedAssetCount > 0}
																<EllipsisCircleIcon
																	class="[&_path]:fill-foreground-alt-1 [&_path]:stroke-foreground size-5 place-self-start fill-amber-700 stroke-amber-700 stroke-1 [&_path]:stroke-2"
																/>
															{:else}
																<PlayCircleIcon
																	class="stroke-foreground-alt-3 fill-background [&_polygon]:stroke-foreground-alt-2 [&_polygon]:fill-foreground-alt-2 size-5 place-self-start stroke-1"
																/>
															{/if}

															<div class="flex w-full flex-col gap-1.5">
																<!-- Lesson title -->
																<span class="text-foreground-alt-2 w-full font-semibold"
																	>{assetGroup.prefix}. {assetGroup.title}</span
																>

																<!-- Lesson details -->
																<div
																	class="relative flex w-full flex-col gap-0 text-sm select-none"
																>
																	<div class="flex w-full flex-row flex-wrap items-center gap-2">
																		<!-- Type -->
																		<span class="text-foreground-alt-3 whitespace-nowrap">
																			{#if isCollection}
																				collection
																			{:else}
																				{assetGroup.assets[0].assetType}
																			{/if}
																		</span>

																		<!-- Video duration -->
																		{#if totalVideoDuration > 0}
																			<DotIcon class="text-foreground-alt-3 text-xl" />

																			<span class="text-foreground-alt-3 whitespace-nowrap">
																				{prettyMs(totalVideoDuration * 1000)}
																			</span>
																		{/if}

																		<!-- Attachments -->
																		{#if assetGroup.attachments.length > 0}
																			<DotIcon class="text-foreground-alt-3 text-xl" />

																			<Attachments
																				attachments={assetGroup.attachments}
																				courseId={course?.id ?? ''}
																				assetId={assetGroup.assets[0].id}
																			/>
																		{/if}
																	</div>
																</div>
															</div>
														</Button>
													</div>
												</li>
											{/each}
										</ol>
									</div>
								</div>
							</section>
						{/each}

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
