<script lang="ts">
	import { goto, pushState } from '$app/navigation';
	import { page } from '$app/state';
	import type { APIError } from '$lib/api-error.svelte';
	import { CreateCourse } from '$lib/api/course-api';
	import { GetFileSystem } from '$lib/api/fs-api';
	import { Oops, Spinner } from '$lib/components';
	import {
		ActionIcon,
		ArrowBackIcon,
		CourseIcon,
		DeselectAllIcon,
		PlusIcon,
		RefreshIcon,
		RightChevronIcon,
		SelectAllIcon
	} from '$lib/components/icons';
	import { Badge, Button, Checkbox, Dialog, Drawer, Dropdown } from '$lib/components/ui';
	import type { CourseCreateModel } from '$lib/models/course-model';
	import { FsPathClassification, type FsModel } from '$lib/models/fs-model';
	import { cn, remCalc } from '$lib/utils';
	import { Separator } from 'bits-ui';
	import { toast } from 'svelte-sonner';
	import { innerWidth, outerHeight } from 'svelte/reactivity/window';
	import theme from 'tailwindcss/defaultTheme';

	type Props = {
		successFn?: () => void;
	};

	let { successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let open = $state(false);

	let fs: FsModel = $state({
		count: 0,
		directories: [],
		files: []
	});

	// When the filesystem for a path has been loaded, this will be set
	let currentPath = $state('');
	let pathHistory = $state<string[]>([]);

	// When a path is selected, this will be set
	let selectedPath = $state('');

	let selectedCourses: Record<string, string> = $state({});
	let selectedCoursesCount = $derived(Object.keys(selectedCourses).length);

	let isPosting = $state(false);
	let isRefreshing = $state(false);
	let isMovingBack = $state(false);

	const widthBreakpoint = +theme.screens.md.replace('rem', '');
	const heightBreakpoint = 520;
	let showDialog = $derived(
		remCalc(innerWidth.current ?? 0) > widthBreakpoint &&
			(outerHeight.current ?? 0) > heightBreakpoint
	);

	let deselectAllDisabled = $derived.by(() => {
		if (isPosting || isRefreshing || selectedCoursesCount === 0) return true;
		return fs.directories.find((dir) => dir.path in selectedCourses) === undefined;
	});

	let selectAllDisabled = $derived.by(() => {
		if (isPosting || isRefreshing) return true;

		return (
			fs.directories.find(
				(dir) => dir.classification === FsPathClassification.None && !(dir.path in selectedCourses)
			) === undefined
		);
	});

	let mainEl: HTMLElement | null = null;

	let loadPromise = $state<Promise<void>>();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// As pathHistory changes, this will trigger
	$effect(() => {
		if (!open) return;

		const path = pathHistory.length > 0 ? pathHistory[pathHistory.length - 1] : '';
		if (path == currentPath) return;

		if (currentPath !== '' && currentPath.startsWith(path)) {
			isMovingBack = true;
		} else {
			selectedPath = path;
		}

		load(path).then(() => {
			currentPath = path;
			selectedPath = '';
			isMovingBack = false;
		});
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// This effect triggers when the page state changes, such as the back button being
	// clicked. It will close the modal if it is open and it shouldn't be
	$effect(() => {
		const s = page.state as { modal?: 'add-courses' } | undefined;
		const shouldBeOpen = s?.modal === 'add-courses';
		if (open && !shouldBeOpen) {
			open = false;
			cleanup();
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function openModal() {
		// Only add an entry if we’re not already on the modal state
		const s = page.state as { modal?: 'add-courses' } | undefined;
		if (s?.modal !== 'add-courses') {
			pushState('', { modal: 'add-courses' as const });
		}

		loadPromise = load('');
	}

	function cleanup() {
		selectedCourses = {};
		currentPath = '';
		selectedPath = '';
		pathHistory = [];
	}

	// Handle closing the modal/drawer - removes modal state from history if present
	function handleClose(isDrawer: boolean = false) {
		cleanup();
		// When closing, check if modal state is still present
		// If it is, we're closing by clicking outside (or sliding down for drawer) → remove history entry
		// If it's not, back button was clicked → effect already handled it → do nothing
		const s = page.state as { modal?: 'add-courses' } | undefined;
		if (s?.modal === 'add-courses') {
			if (isDrawer) {
				// Drawer: remove entry and push new state to stay on page (handles slide-down gesture)
				history.back();
				setTimeout(() => {
					pushState('', {});
				}, 0);
			} else {
				// Dialog: just remove the entry (clicking outside)
				history.back();
			}
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Load the drives or the directories of the selected path
	async function load(path: string): Promise<void> {
		try {
			const flickerPromise = new Promise((resolve) => setTimeout(resolve, 200));
			const [response] = await Promise.all([GetFileSystem(path), flickerPromise]);

			if (mainEl) mainEl.scrollTop = 0;

			fs = response;
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Print the number of selected courses
	function toastCount() {
		if (selectedCoursesCount === 0) {
			toast.success('No courses selected');
		} else {
			toast.success(
				`${selectedCoursesCount} course${selectedCoursesCount > 1 ? 's' : ''} selected`
			);
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Add the selected courses
	async function addCourses(): Promise<void> {
		isPosting = true;

		try {
			await Promise.all(
				Object.keys(selectedCourses).map((path) =>
					CreateCourse({ path, title: selectedCourses[path] } satisfies CourseCreateModel)
				)
			);
			successFn?.();

			// Clear selected courses after successful addition
			selectedCourses = {};

			// Only close on success
			open = false;
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
	}
</script>

{#snippet trigger()}
	{#if showDialog}
		<Dialog.Trigger class="flex h-10 w-auto flex-row items-center gap-2 px-5">
			<PlusIcon class="size-5 stroke-[1.5]" />
			Add Courses
		</Dialog.Trigger>
	{:else}
		<Drawer.Trigger class="flex h-10 w-auto flex-row items-center gap-2 px-5 py-2">
			<PlusIcon class="size-5 stroke-[1.5]" />
			Add Courses
		</Drawer.Trigger>
	{/if}
{/snippet}

{#snippet refreshButton()}
	<Button
		variant="ghost"
		class="mr-2.5 w-auto"
		disabled={isRefreshing || isPosting || isMovingBack}
		onclick={async () => {
			isRefreshing = true;
			await load(currentPath);
			isRefreshing = false;
		}}
	>
		<RefreshIcon
			class={cn('text-foreground-alt-1 size-5 stroke-2', isRefreshing && 'animate-spin')}
		/>
	</Button>
{/snippet}

{#snippet addButton()}
	<Button
		class="w-25"
		disabled={isPosting || isRefreshing || selectedCoursesCount === 0}
		onclick={addCourses}
	>
		{#if isPosting}
			<Spinner class="bg-background-alt-4 size-2" />
		{:else}
			Add
		{/if}
	</Button>
{/snippet}

{#snippet contents()}
	<main
		bind:this={mainEl}
		class="flex min-h-20 w-full flex-1 flex-col overflow-y-auto overflow-x-hidden"
	>
		{#await loadPromise}
			<div class="flex justify-center pt-10">
				<Spinner class="bg-foreground-alt-3 size-4" />
			</div>
		{:then _}
			<!-- 
				Back button 
				
				Only render when the selected path is not empty or the user is moving
				back to prevent layout shift when navigating directories
			-->
			{#if currentPath !== '' || isMovingBack}
				<div class="border-background-alt-3 flex flex-row items-center border-b">
					<Button
						variant="ghost"
						class=" h-14 grow justify-start rounded-none px-3 py-2 text-start duration-0"
						disabled={isMovingBack || isPosting || isRefreshing || selectedPath !== ''}
						onclick={async () => {
							pathHistory.length > 0 ? pathHistory.pop() : null;
						}}
					>
						<ArrowBackIcon class="size-4 stroke-2" />
						<span>Back</span>
					</Button>

					{#if isMovingBack}
						<div class="flex h-full w-20 shrink-0 justify-center">
							<div class="flex w-full place-content-center">
								<Spinner class="bg-foreground-alt-3 size-2.5" />
							</div>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Filesystem directories -->
			{#each fs.directories as dir (dir.path)}
				<div class="border-background-alt-3 flex flex-row items-stretch border-b">
					<Button
						variant="ghost"
						class="h-auto! wrap-break-word min-h-14 min-w-0 shrink grow basis-0 items-center justify-start whitespace-normal rounded-none px-3 py-2 text-start duration-0"
						disabled={isPosting ||
							isRefreshing ||
							isMovingBack ||
							selectedPath !== '' ||
							selectedCourses[dir.path] !== undefined ||
							dir.classification === FsPathClassification.Course}
						onclick={async () => {
							pathHistory.push(dir.path);
						}}
					>
						<span class="wrap-break-word block min-w-0">{dir.title}</span>
					</Button>

					<!-- Selection -->
					<div class="flex w-20 flex-none shrink-0 basis-20 justify-center self-stretch">
						<Separator.Root
							orientation="vertical"
							class="bg-background-alt-3 h-full w-px shrink-0"
						/>

						{#if dir.classification === FsPathClassification.Course}
							<div class="flex w-full justify-center place-self-center">
								<Badge
									class="border-foreground-alt-3 text-foreground-alt-3 h-auto border bg-transparent text-xs"
								>
									Added
								</Badge>
							</div>
						{:else if dir.path === selectedPath}
							<div class="flex w-full place-content-center">
								<Spinner class="bg-foreground-alt-3 size-2.5" />
							</div>
						{:else}
							<Button
								variant="ghost"
								class={cn(
									'group h-full w-full rounded-none p-0',
									dir.classification === FsPathClassification.Ancestor &&
										'cursor-default hover:bg-transparent'
								)}
								disabled={isPosting ||
									isRefreshing ||
									selectedPath !== '' ||
									(dir.classification !== FsPathClassification.None &&
										!(dir.path in selectedCourses)) ||
									(!!(dir.path in selectedCourses)
										? false
										: Object.keys(selectedCourses).some((path) => path.startsWith(dir.path)))}
								onclick={() => {
									dir.path in selectedCourses
										? delete selectedCourses[dir.path]
										: (selectedCourses[dir.path] = dir.title);

									toastCount();
								}}
							>
								<Checkbox
									disabled={isPosting || isRefreshing || selectedPath !== ''}
									class="group-hover:border-foreground-alt-1 border-2 duration-200 data-[state=indeterminate]:cursor-default"
									checked={selectedCourses[dir.path] !== undefined}
									indeterminate={dir.classification === FsPathClassification.Ancestor ||
										(!!(dir.path in selectedCourses)
											? false
											: Object.keys(selectedCourses).some((path) => path.startsWith(dir.path)))}
									onclick={(e) => {
										e.preventDefault();
									}}
								/>
							</Button>
						{/if}
					</div>
				</div>
			{/each}
		{:catch error}
			<Oops
				class="pt-0"
				contentClass="border-0"
				message={'Failed to fetch file system information: ' + error.message}
			/>
		{/await}
	</main>
{/snippet}

{#snippet selectDeselect()}
	{#if showDialog}
		<div class="flex justify-start gap-2">
			<Button
				variant="outline"
				class="w-auto"
				disabled={selectAllDisabled}
				onclick={() => {
					// Select all courses current not selected (and can be selected)
					fs.directories.forEach((dir) => {
						if (dir.classification === FsPathClassification.None) {
							selectedCourses[dir.path] = dir.title;
						}
					});

					toastCount();
				}}
			>
				<SelectAllIcon class="size-4 stroke-[1.5]" />
				Select all
			</Button>

			<Button
				variant="outline"
				class="w-auto"
				disabled={deselectAllDisabled}
				onclick={() => {
					// Remove all selected courses
					Object.keys(selectedCourses).forEach((path) => {
						if (fs.directories.find((dir) => dir.path === path)) {
							delete selectedCourses[path];
						}
					});

					toastCount();
				}}
			>
				<DeselectAllIcon class="size-4 stroke-[1.5]" />
				Deselect All
			</Button>
		</div>
	{:else}
		<Dropdown.Root>
			<Dropdown.Trigger class="w-32 [&[data-state=open]>svg]:rotate-90">
				<div class="flex items-center gap-1.5">
					<ActionIcon class="size-4 stroke-[1.5]" />
					<span>Actions</span>
				</div>
				<RightChevronIcon class="stroke-foreground-alt-3 size-4.5 duration-200" />
			</Dropdown.Trigger>

			<Dropdown.Content class="w-42" align="start" portalProps={{ disabled: true }}>
				<Dropdown.Item
					disabled={selectAllDisabled}
					onclick={() => {
						if (selectAllDisabled) return;

						// Select all courses current not selected (and can be selected)
						fs.directories.forEach((dir) => {
							if (dir.classification === FsPathClassification.None) {
								selectedCourses[dir.path] = dir.title;
							}
						});

						toastCount();
					}}
				>
					<SelectAllIcon class="size-4 stroke-[1.5]" />
					Select All
				</Dropdown.Item>

				<Dropdown.Item
					disabled={deselectAllDisabled}
					onclick={() => {
						if (deselectAllDisabled) return;

						// Remove all selected courses
						Object.keys(selectedCourses).forEach((path) => {
							if (fs.directories.find((dir) => dir.path === path)) {
								delete selectedCourses[path];
							}
						});

						toastCount();
					}}
				>
					<DeselectAllIcon class="size-4 stroke-[1.5]" />
					Deselect All
				</Dropdown.Item>
			</Dropdown.Content>
		</Dropdown.Root>
	{/if}
{/snippet}

{#if showDialog}
	<Dialog.Root
		bind:open
		{trigger}
		onOpenChange={(open) => {
			if (open) {
				openModal();
			} else {
				handleClose(false);
			}
		}}
	>
		<Dialog.Content class="inline-flex h-[min(calc(100vh-10rem),50rem)] max-w-2xl flex-col">
			<Dialog.Header>
				<div class="flex items-center gap-2">
					<CourseIcon class="size-5 stroke-2" />
					<span>Course Selection</span>
				</div>

				{@render refreshButton()}
			</Dialog.Header>

			{@render contents()}

			<Dialog.Footer class="flex justify-between">
				{@render selectDeselect()}

				<div class="flex justify-end gap-2">
					<Dialog.CloseButton>Close</Dialog.CloseButton>

					{@render addButton()}
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{:else}
	<Drawer.Root
		bind:open
		onOpenChange={(open) => {
			if (open) {
				openModal();
			} else {
				handleClose(true);
			}
		}}
	>
		{@render trigger()}

		<Drawer.Content class="bg-background-alt-2" handleClass="bg-background-alt-4">
			<Drawer.Header>
				<div class="flex items-center gap-2">
					<CourseIcon class="size-5 stroke-2" />
					<span>Course Selection</span>
				</div>

				{@render refreshButton()}
			</Drawer.Header>

			{@render contents()}

			<Drawer.Footer class="flex justify-between">
				{@render selectDeselect()}

				<div class="flex justify-end gap-2">
					<Drawer.CloseButton>Close</Drawer.CloseButton>
					{@render addButton()}
				</div>
			</Drawer.Footer>
		</Drawer.Content>
	</Drawer.Root>
{/if}
