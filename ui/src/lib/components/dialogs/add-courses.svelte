<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { CreateCourse } from '$lib/api/course-api';
	import { GetFileSystem } from '$lib/api/fs-api';
	import { Oops, Spinner } from '$lib/components';
	import { BackArrowIcon, CourseIcon, PlusIcon, RefreshIcon } from '$lib/components/icons';
	import { Badge, Button, Checkbox, Dialog } from '$lib/components/ui';
	import { FsPathClassification, type FsModel } from '$lib/models/fs-model';
	import { cn } from '$lib/utils';
	import { Separator } from 'bits-ui';
	import { toast } from 'svelte-sonner';

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

	let paths: string[] = $state([]);
	let selectedPath = $state('');

	let selectedCourses: Record<string, string> = $state({});
	let selectedCoursesCount = $derived(Object.keys(selectedCourses).length);

	let isPosting = $state(false);
	let isRefreshing = $state(false);

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

	const backId = 'back-' + Math.random().toString(36);

	let loadPromise = $state<Promise<void>>();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			loadPromise = load('');
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Load the drives or the directories of the selected path
	async function load(path: string): Promise<void> {
		if (path === '' || paths.includes(path)) {
			selectedPath = backId;
		} else {
			selectedPath = path;
		}

		try {
			const flickerPromise = new Promise((resolve) => setTimeout(resolve, 200));
			const [response] = await Promise.all([GetFileSystem(path), flickerPromise]);

			if (mainEl) mainEl.scrollTop = 0;

			fs = response;
		} catch (error) {
			throw error;
		} finally {
			if (path !== '' && !paths.includes(path)) paths.push(path);
			selectedPath = '';
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Move the user back to the previous path
	async function moveBack() {
		if (paths.length === 1) {
			await load('');
		} else {
			await load(paths[paths.length - 2]);
		}

		paths.pop();
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
					CreateCourse({ path, title: selectedCourses[path] })
				)
			);
			successFn?.();
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
		open = false;
	}
</script>

<Dialog.Root
	bind:open
	onOpenChange={() => {
		selectedCourses = {};
		paths = [];
		selectedPath = '';
	}}
>
	{#snippet trigger()}
		<Dialog.Trigger class="flex h-10 w-auto flex-row items-center gap-2 px-5">
			<PlusIcon class="size-5 stroke-[1.5]" />
			Add Courses
		</Dialog.Trigger>
	{/snippet}

	<Dialog.Content class="inline-flex h-[min(calc(100vh-10rem),50rem)] max-w-2xl flex-col">
		<Dialog.Header>
			<div class="flex items-center gap-2">
				<CourseIcon class="size-5 stroke-2" />
				<span>Course Selection</span>
			</div>

			<Button
				disabled={isRefreshing}
				class="hover:bg-background-alt-4 mr-2 w-auto bg-transparent px-2 disabled:bg-transparent"
				onclick={async () => {
					isRefreshing = true;
					paths.length > 0 ? await load(paths[paths.length - 1]) : await load('');
					isRefreshing = false;
				}}
			>
				<RefreshIcon
					class={cn('text-foreground-alt-1 size-5 stroke-2', isRefreshing && 'animate-spin')}
				/>
			</Button>
		</Dialog.Header>

		<main
			bind:this={mainEl}
			class="flex min-h-[5rem] w-full flex-1 flex-col overflow-x-hidden overflow-y-auto"
		>
			{#await loadPromise}
				<div class="flex justify-center pt-10">
					<Spinner class="bg-foreground-alt-2 size-4" />
				</div>
			{:then _}
				<!-- Back button -->
				{#if paths.length > 0}
					{#key paths[paths.length - 1]}
						<div class="border-background-alt-3 flex flex-row items-center border-b">
							<Button
								class="text-foreground-alt-1 hover:bg-background disabled:text-foreground-alt-2 h-14 grow justify-start rounded-none bg-transparent p-0 px-3 text-start whitespace-normal duration-0 disabled:bg-transparent disabled:hover:cursor-default"
								disabled={selectedPath !== '' || isPosting || isRefreshing}
								onclick={async () => {
									await moveBack();
								}}
							>
								<BackArrowIcon class="size-4 stroke-2" />
								<span>Back</span>
							</Button>

							{#if backId === selectedPath}
								<div class="flex h-full w-20 shrink-0 justify-center">
									<div class="flex w-full place-content-center">
										<Spinner class="bg-foreground-alt-2 size-2.5" />
									</div>
								</div>
							{/if}
						</div>
					{/key}
				{/if}

				<!-- Filesystem directories -->
				{#each fs.directories as dir (dir.path)}
					<div class="border-background-alt-3 flex flex-row items-center border-b">
						<Button
							class="text-foreground-alt-1 hover:bg-background disabled:text-foreground-alt-2 h-full min-h-14 grow justify-start rounded-none bg-transparent p-0 px-3 py-2 text-start whitespace-normal duration-0 disabled:bg-transparent disabled:hover:cursor-default"
							disabled={isPosting ||
								isRefreshing ||
								selectedPath !== '' ||
								selectedCourses[dir.path] !== undefined ||
								dir.classification === FsPathClassification.Course}
							onclick={async () => {
								await load(dir.path);
							}}
						>
							{dir.title}
						</Button>

						<div class="flex h-full w-20 shrink-0 justify-center">
							<Separator.Root orientation="vertical" class="bg-background-alt-3 h-full w-px" />

							{#if dir.classification === FsPathClassification.Course}
								<div class="flex w-full justify-center place-self-center">
									<Badge
										class="border-foreground-alt-2 text-foreground-alt-2 h-auto border bg-transparent text-xs"
									>
										Added
									</Badge>
								</div>
							{:else if dir.path === selectedPath}
								<div class="flex w-full place-content-center">
									<Spinner class="bg-foreground-alt-2 size-2.5" />
								</div>
							{:else}
								<Button
									class={cn(
										'hover:bg-background group disabled:text-foreground-alt-2 h-full w-full rounded-none bg-transparent p-0 disabled:bg-transparent disabled:hover:cursor-default',
										dir.classification === FsPathClassification.Ancestor &&
											'cursor-default hover:bg-transparent'
									)}
									disabled={isPosting ||
										isRefreshing ||
										selectedPath !== '' ||
										dir.classification !== FsPathClassification.None ||
										Object.keys(selectedCourses).some((path) => path.startsWith(dir.path))}
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
											Object.keys(selectedCourses).some((path) => path.startsWith(dir.path))}
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

		<Dialog.Footer class="flex justify-between">
			<div class="flex justify-start gap-2">
				<Button
					class="border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground disabled:text-foreground-alt-2 w-24 cursor-pointer rounded-md border bg-transparent py-2 duration-200 select-none disabled:bg-transparent"
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
					Select All
				</Button>

				<Button
					class="border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground disabled:text-foreground-alt-2 w-28 cursor-pointer rounded-md border bg-transparent py-2 duration-200 select-none disabled:bg-transparent"
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
					Deselect All
				</Button>
			</div>

			<div class="flex justify-end gap-2">
				<Dialog.CloseButton />

				<Button
					onclick={addCourses}
					disabled={isPosting || isRefreshing || selectedCoursesCount === 0}
					class="h-10 w-25 py-2"
				>
					{#if !isPosting}
						Add
					{:else}
						<Spinner class="bg-foreground-alt-3 size-2" />
					{/if}
				</Button>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
