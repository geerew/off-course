<script lang="ts">
	import { afterNavigate, goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		GetCourse,
		GetCourseModules,
		ServeCourseAsset,
		UpdateCourseAssetProgress
	} from '$lib/api/course-api';
	import { Spinner } from '$lib/components';
	import {
		BurgerMenuIcon,
		DotIcon,
		DotsIcon,
		EllipsisCircleIcon,
		LeftChevronIcon,
		OverviewIcon,
		RightChevronIcon,
		TickCircleIcon,
		TickIcon,
		WarningIcon
	} from '$lib/components/icons';
	import { Button, Dropdown, Tooltip } from '$lib/components/ui';
	import AssetError from '$lib/components/ui/asset-error.svelte';
	import Attachments from '$lib/components/ui/attachments.svelte';
	import { VideoPlayer } from '$lib/components/ui/media';
	import type { AssetModel, AssetProgressUpdateModel } from '$lib/models/asset-model';
	import type { CourseModel } from '$lib/models/course-model';
	import type { LessonModel, ModulesModel } from '$lib/models/module-model';
	import { cn, renderMarkdown, toVideoMimeType } from '$lib/utils';
	import { Dialog } from 'bits-ui';
	import prettyMs from 'pretty-ms';
	import { ElementSize } from 'runed';
	import { tick } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';

	let course = $state<CourseModel>();
	let modules = $state<ModulesModel>();

	let selectedLesson = $state<LessonModel>();

	let previousLesson = $state<LessonModel>();
	let nextLesson = $state<LessonModel>();

	const contentCache = new SvelteMap<string, string>();
	const loadingErrors = new SvelteMap<string, string>();

	let loadPromise = $state(fetcher());
	let initDone = false;

	let mainEl = $state() as HTMLElement;
	const mainSize = new ElementSize(() => mainEl);

	let staticMenuEl = $state<HTMLElement>();
	let dialogMenuEl = $state<HTMLElement | null>(null);

	let menuPopupMode = $state(false);
	let dialogOpen = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch the course, then the assets for the course
	async function fetcher(): Promise<void> {
		try {
			const courseId = page.params.course_id;
			if (!courseId) throw new Error('No course ID provided');

			const lessonId = page.params.lesson_id;
			if (!lessonId) throw new Error('No lesson ID provided');

			course = await GetCourse(courseId);
			if (!course) throw new Error('Course not found');

			modules = await GetCourseModules(course.id, { withUserProgress: true });

			selectedLesson = findLesson(lessonId, modules);
			if (!selectedLesson) throw new Error('Failed to find lesson');

			for (const a of selectedLesson.assets) {
				if (a.type === 'markdown' || a.type === 'text') {
					void loadAndRenderContent(a);
				}
			}

			previousLesson = findPreviousLesson();
			nextLesson = findNextLesson();

			initDone = true;
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// After navigating, make sure to scroll to the top of the page
	afterNavigate(() => mainEl?.scrollTo({ top: 0, behavior: 'smooth' }));

	// After navigating, make sure to scroll the selected lesson into view
	afterNavigate(() => scrollSelectedIntoView());

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Find the previous lesson
	function findPreviousLesson(): LessonModel | undefined {
		if (!modules || !selectedLesson) return undefined;

		const allLessons = modules.modules.flatMap((m) => m.lessons);
		const current = selectedLesson; // narrowed non-null

		const idx = allLessons.findIndex((l) => l.id === current.id);
		return idx > 0 ? allLessons[idx - 1] : undefined;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Find the next lesson
	function findNextLesson(): LessonModel | undefined {
		if (!modules || !selectedLesson) return undefined;

		const allLessons = modules.modules.flatMap((m) => m.lessons);
		const current = selectedLesson;

		const idx = allLessons.findIndex((l) => l.id === current.id);
		return idx >= 0 && idx < allLessons.length - 1 ? allLessons[idx + 1] : undefined;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the rendered description for the asset
	async function loadAndRenderContent(asset: AssetModel): Promise<string> {
		if (!asset || (asset.type !== 'markdown' && asset.type !== 'text')) return '';

		if (contentCache.has(asset.id)) {
			return contentCache.get(asset.id)!;
		}

		try {
			const raw = await ServeCourseAsset(asset.courseId, asset.lessonId, asset.id);
			if (!raw) {
				loadingErrors.set(asset.id, 'No content available');
				return '';
			}

			const rendered = asset.type === 'text' ? raw : renderMarkdown(raw);

			// Clear any previous error and update reactive state map
			loadingErrors.delete(asset.id);
			contentCache.set(asset.id, rendered);

			return rendered;
		} catch (error) {
			// Store the error message for display
			const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
			loadingErrors.set(asset.id, errorMessage);
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Retry loading a failed asset
	async function retryAsset(asset: AssetModel): Promise<void> {
		if (!asset) return;

		// Clear any existing error and cache
		loadingErrors.delete(asset.id);
		contentCache.delete(asset.id);

		if (asset.type === 'markdown' || asset.type === 'text') {
			try {
				await loadAndRenderContent(asset);
			} catch (error) {
				// Error is already handled in loadAndRenderContent
			}
		}
		// For video assets, clearing the error will cause the VideoPlayer to re-render
		// and attempt to load the video again
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Find the lesson from the modules based upon the lesson ID
	function findLesson(lessonId: string, modules: ModulesModel): LessonModel | undefined {
		for (const mod of modules.modules) {
			const lesson = mod.lessons.find((l) => l.id === lessonId);
			if (lesson) return lesson;
		}
		return undefined;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Get the active menu element (static or dialog)
	function getActiveMenuEl(): HTMLElement | null {
		return (menuPopupMode && dialogOpen ? dialogMenuEl : staticMenuEl) ?? null;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function raf() {
		return new Promise<void>((r) => requestAnimationFrame(() => r()));
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Scroll the selected lesson into view
	async function scrollSelectedIntoView() {
		const container = getActiveMenuEl();
		if (!container || !selectedLesson) return;
		await tick();

		const item = container.querySelector<HTMLElement>(`[data-lesson-id="${selectedLesson.id}"]`);
		if (!item) return;

		const top = item.offsetTop - container.offsetTop;
		const target = Math.max(0, top + item.offsetHeight / 2 - container.clientHeight / 2);
		container.scrollTo({ top: target, behavior: 'smooth' });
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Scroll the menu when it is ready (after dialog open animation)
	async function scrollMenuWhenReady() {
		// pick the active menu element
		const container = menuPopupMode && dialogOpen ? dialogMenuEl : staticMenuEl;
		if (!container || !selectedLesson) return;

		// wait for DOM mount & initial paint(s)
		await tick();
		await raf(); // 1st paint
		await raf(); // 2nd paint (after slide-in kicks in)

		// now scroll
		const item = container.querySelector<HTMLElement>(`[data-lesson-id="${selectedLesson.id}"]`);
		if (!item) return;

		item.scrollIntoView({ block: 'center', behavior: 'smooth' });
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update the selected asset when the page changes
	$effect(() => {
		const lessonId = page.params.lesson_id;
		if (!initDone || !selectedLesson || !modules) return;

		if (!lessonId) {
			throw new Error('No lesson ID provided');
		}

		selectedLesson = findLesson(lessonId, modules);
		if (!selectedLesson) {
			throw new Error('Lesson not found');
		}

		previousLesson = findPreviousLesson();
		nextLesson = findNextLesson();
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

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// When the menu changes (e.g. on initial load) or the selected lesson changes, scroll the selected
	$effect(() => {
		if (!selectedLesson || !modules) return;
		scrollSelectedIntoView();
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// When the dialog opens in popup mode, scroll the selected lesson into view
	$effect(() => {
		if (menuPopupMode && dialogOpen) {
			scrollMenuWhenReady();
		}
	});
</script>

{#snippet menuContents()}
	{#if course}
		<div
			class="bg-background border-background-alt-4 sticky top-0 z-1 flex flex-row gap-3 border-b"
		>
			<div class="flex w-full items-center justify-between gap-4 py-4 pr-3">
				<span class="container-pl font-semibold select-none">{course.title}</span>

				<Dropdown.Root>
					<Dropdown.Trigger
						class="hover:bg-background-alt-3 data-[state=open]:bg-background-alt-3 w-auto rounded-lg border-none"
					>
						<DotsIcon class="size-5 stroke-[1.5]" />
					</Dropdown.Trigger>

					<Dropdown.Content class="z-60 w-38">
						<Dropdown.Item
							onclick={async () => {
								if (!course) return;
								goto(`/course/${course.id}`);
							}}
						>
							<OverviewIcon class="size-4 stroke-[1.5]" />
							<span>Overview</span>
						</Dropdown.Item>
					</Dropdown.Content>
				</Dropdown.Root>
			</div>
		</div>

		{#if modules}
			{#each modules.modules as m}
				<div class="container-pl pt-3 leading-5">
					<div class="flex justify-between gap-2.5 py-1.5 pr-5">
						<span class="text-background-primary text-sm font-semibold tracking-wide">
							{m.module}
						</span>

						<span class="text-foreground-alt-3 shrink-0 py-0.5 text-xs">
							{m.lessons.filter((l) => l.completed).length}
							/ {m.lessons.length}
						</span>
					</div>

					<div class="border-background-alt-4 mt-2 ml-auto flex flex-col gap-3 border-l">
						{#each m.lessons as lesson}
							{@const isCollection = lesson.assets.length > 1}
							{@const totalVideoDuration = lesson.totalVideoDuration}

							<Button
								variant="ghost"
								data-lesson-id={lesson.id}
								data-selected={selectedLesson && selectedLesson.id === lesson.id}
								class={cn(
									'hover:text-foreground-alt-1 relative h-auto w-full justify-start rounded-none text-start whitespace-normal before:absolute before:duration-200 hover:before:top-0 hover:before:-left-px hover:before:h-full hover:before:w-px',
									selectedLesson && selectedLesson.id === lesson.id
										? 'text-foreground-alt-1 before:bg-foreground-alt-1 before:top-0 before:-left-px before:h-full before:w-px'
										: 'text-foreground-alt-2 hover:before:bg-foreground-alt-3'
								)}
								onclick={async () => {
									if (!course || selectedLesson?.id === lesson.id) return;
									if (menuPopupMode) dialogOpen = false;
									goto(`/course/${course.id}/${lesson.id}`);
								}}
							>
								<!-- Lesson status -->
								{#if lesson.completed}
									<TickCircleIcon
										class="stroke-background-success fill-background-success [&_path]:stroke-foreground absolute -left-2.5 size-5 place-self-start stroke-1 [&_path]:stroke-1"
									/>
								{:else if lesson.started}
									<EllipsisCircleIcon
										class="[&_path]:fill-foreground-alt-1 [&_path]:stroke-foreground-alt-1 absolute -left-2.5 size-5 place-self-start fill-amber-700 stroke-amber-700 stroke-1 [&_path]:stroke-2"
									/>
								{/if}

								<div class="flex w-full flex-row gap-3 pr-2.5 pl-2.5">
									<div class="flex w-full flex-col gap-2 text-sm">
										<span>{lesson.prefix}. {lesson.title}</span>

										<div class="relative flex w-full flex-col gap-0 text-sm select-none">
											<div class="flex w-full flex-row flex-wrap items-center gap-2">
												<!-- Type -->
												<span class="text-foreground-alt-3 whitespace-nowrap">
													{#if isCollection}
														collection
													{:else}
														{lesson.assets[0].type}
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
												{#if lesson.attachments.length > 0}
													<DotIcon class="text-foreground-alt-3 text-xl" />
													<Attachments
														attachments={lesson.attachments}
														courseId={course?.id ?? ''}
														lessonId={lesson.id}
													/>
												{/if}
											</div>
										</div>
									</div>
								</div>
							</Button>
						{/each}
					</div>
				</div>
			{/each}
		{/if}
	{/if}
{/snippet}

{#await loadPromise}
	<div class="flex justify-center pt-10">
		<Spinner class="bg-foreground-alt-3 size-4" />
	</div>
{:then _}
	{#if course && selectedLesson}
		<div
			class={cn(
				'grid min-h-0 grid-rows-1 gap-6 pt-[calc(var(--header-height)+1)]',
				menuPopupMode ? 'grid-cols-1' : 'grid-cols-[var(--course-menu-width)_1fr]'
			)}
		>
			<!-- Menu -->
			{#if menuPopupMode}
				<Dialog.Root bind:open={dialogOpen}>
					<Dialog.Portal>
						<Dialog.Overlay
							class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-40 bg-black/50"
						/>

						<Dialog.Content
							bind:ref={dialogMenuEl}
							class="border-background-alt-4 bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left fixed top-0 left-0 z-50 h-full w-(--course-menu-width) border-r"
						>
							<nav class="flex h-full w-full flex-col gap-2 overflow-x-hidden overflow-y-auto pb-8">
								{@render menuContents()}
							</nav>
						</Dialog.Content>
					</Dialog.Portal>
				</Dialog.Root>
			{:else}
				<div class="relative row-span-full min-h-0">
					<nav
						bind:this={staticMenuEl}
						class="border-background-alt-4 bg-background sticky top-[calc(var(--header-height)+1px)] h-[calc(100dvh-(var(--header-height)+1px))] w-[--course-menu-width] overflow-x-hidden overflow-y-auto overscroll-contain border-r pb-8"
					>
						{@render menuContents()}
					</nav>
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
										{selectedLesson.title}
									</span>
								</div>

								<Tooltip
									delayDuration={100}
									contentProps={{ side: 'bottom', sideOffset: 8 }}
									contentClass="text-sm"
								>
									{#snippet trigger()}
										<Button
											variant="ghost"
											class={cn(
												'flex size-8 shrink-0 items-center justify-center rounded-full border border-none',
												selectedLesson?.completed
													? 'bg-background-success hover:bg-background-success text-foreground'
													: 'bg-background-alt-5 text-foreground-alt-3 hover:bg-background-alt-6 hover:text-foreground-alt-2'
											)}
											onclick={async () => {
												if (!selectedLesson || !course) return;

												// Update local state by inverting the completed state and updating started
												// and assetsCompleted
												selectedLesson.completed = !selectedLesson.completed;

												if (selectedLesson.completed) {
													selectedLesson.started = true;
													selectedLesson.assetsCompleted = selectedLesson.assets.length;
												} else {
													selectedLesson.assetsCompleted = 0;
												}

												// Update this lessons assets to match the lesson completed state (local and
												// backend)
												await Promise.all(
													selectedLesson.assets.map(async (asset) => {
														if (!selectedLesson) return;

														asset.progress.completed = selectedLesson.completed;

														const progress: AssetProgressUpdateModel = {
															completed: asset.progress.completed
														};

														await UpdateCourseAssetProgress(
															asset.courseId,
															asset.lessonId,
															asset.id,
															progress
														);
													})
												);
											}}
										>
											<TickIcon class="size-4 stroke-3" />
										</Button>
									{/snippet}

									{#snippet content()}
										{#if selectedLesson}
											Mark {selectedLesson.assets.length > 1 ? 'collection' : ''}
											{selectedLesson.completed ? 'unwatched' : 'watched'}
										{/if}
									{/snippet}
								</Tooltip>
							</div>
						</div>

						<!-- Asset(s) -->
						{#each selectedLesson.assets as asset}
							<div class="flex w-full flex-col gap-4">
								{#if selectedLesson.assets.length > 1}
									<span class="text-lg font-medium">
										{asset.subPrefix}. {asset.subTitle ? asset.subTitle : asset.title}
									</span>
								{/if}

								{#if asset.type === 'video'}
									<VideoPlayer
										playerId={`${selectedLesson.id}-${asset.id}`}
										src={`/api/hls/${asset.id}/master.m3u8`}
										srcType={toVideoMimeType(asset.metadata.video?.mimeType) || 'video/object'}
										useHls={true}
										startTime={asset.progress.position || 0}
										onTimeChange={async (time: number) => {
											if (!selectedLesson) return;

											asset.progress.position = time;

											const progress: AssetProgressUpdateModel = {
												completed: asset.progress.completed,
												position: time
											};

											await UpdateCourseAssetProgress(
												asset.courseId,
												asset.lessonId,
												asset.id,
												progress
											);

											selectedLesson.started = true;
										}}
										onCompleted={async (time: number) => {
											if (!selectedLesson) return;

											asset.progress.position = time;
											asset.progress.completed = true;

											const progress: AssetProgressUpdateModel = {
												completed: true,
												position: time
											};

											await UpdateCourseAssetProgress(
												asset.courseId,
												asset.lessonId,
												asset.id,
												progress
											);

											// Update the local state
											selectedLesson.started = true;
											selectedLesson.assetsCompleted += 1;

											if (selectedLesson.assetsCompleted === selectedLesson.assets.length)
												selectedLesson.completed = true;
										}}
									/>
								{:else if asset.type === 'markdown'}
									<div class="typography">
										{#if contentCache.has(asset.id)}
											<div class="typography">{@html contentCache.get(asset.id)}</div>
										{:else if loadingErrors.has(asset.id)}
											<AssetError
												assetType="markdown"
												assetTitle={asset.subTitle || asset.title}
												error={loadingErrors.get(asset.id)}
												onRetry={() => retryAsset(asset)}
											/>
										{:else}
											{#await loadAndRenderContent(asset)}
												<div class="text-foreground-alt-3 text-sm">Loading…</div>
											{:then html}
												<div class="typography">{@html html}</div>
											{:catch e}
												<AssetError
													assetType="markdown"
													assetTitle={asset.subTitle || asset.title}
													error={e.message}
													onRetry={() => retryAsset(asset)}
												/>
											{/await}
										{/if}
									</div>
								{:else if asset.type === 'text'}
									{#if contentCache.has(asset.id)}
										<div class="whitespace-pre-wrap">{contentCache.get(asset.id)}</div>
									{:else if loadingErrors.has(asset.id)}
										<AssetError
											assetType="text"
											assetTitle={asset.subTitle || asset.title}
											error={loadingErrors.get(asset.id)}
											onRetry={() => retryAsset(asset)}
										/>
									{:else}
										{#await loadAndRenderContent(asset)}
											<div class="text-foreground-alt-3 text-sm">Loading…</div>
										{:then text}
											<div class="whitespace-pre-wrap">{text}</div>
										{:catch e}
											<AssetError
												assetType="text"
												assetTitle={asset.subTitle || asset.title}
												error={e.message}
												onRetry={() => retryAsset(asset)}
											/>
										{/await}
									{/if}
								{:else}
									(TODO PDF)
								{/if}
							</div>
						{/each}

						<div class="flex w-full flex-col gap-4 md:flex-row md:gap-6">
							<div class="flex w-full md:w-1/2">
								{#if previousLesson}
									<Button
										variant="outline"
										class="text-foreground-alt-2 hover:text-foreground hover:border-background-alt-6 flex h-auto w-full flex-row justify-start gap-4 p-4 text-left whitespace-normal hover:bg-transparent"
										onclick={() => {
											if (!course || !previousLesson) return;
											goto(`/course/${course.id}/${previousLesson.id}`);
										}}
									>
										<LeftChevronIcon class="size-5 stroke-[1.5]" />
										<span class="text-base leading-tight font-medium">
											{previousLesson.prefix}. {previousLesson.title}
										</span>
									</Button>
								{/if}
							</div>

							<!-- Next Button -->
							<div class="flex w-full md:w-1/2">
								{#if nextLesson}
									<Button
										variant="outline"
										class="text-foreground-alt-2 hover:text-foreground hover:border-background-alt-6 flex h-auto w-full flex-row justify-end gap-4 p-4 text-left whitespace-normal hover:bg-transparent"
										onclick={() => {
											if (!course || !nextLesson) return;
											goto(`/course/${course.id}/${nextLesson.id}`);
										}}
									>
										<span class="text-base leading-tight font-medium">
											{nextLesson.prefix}.
											{nextLesson.title}
										</span>
										<RightChevronIcon class="size-5 stroke-[1.5]" />
									</Button>
								{/if}
							</div>
						</div>
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
