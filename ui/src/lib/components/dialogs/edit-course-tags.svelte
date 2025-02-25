<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { CreateCourseTag, DeleteCourseTag, GetCourseTags } from '$lib/api/course-api';
	import { GetTagNames } from '$lib/api/tag-api';
	import { Oops } from '$lib/components/';
	import { ScanIcon, UndoIcon, XIcon } from '$lib/components/icons';
	import { Badge, Button, Dialog } from '$lib/components/ui';
	import type { CourseModel, CoursesModel, CourseTagsModel } from '$lib/models/course-model';
	import { cn } from '$lib/utils';
	import { Combobox } from 'bits-ui';
	import { Debounced } from 'runed';
	import { type Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';
	import Spinner from '../spinner.svelte';

	type Props = {
		open?: boolean;
		value: CourseModel | CoursesModel;
		trigger?: Snippet;
		successFn?: () => void;
	};

	let { open = $bindable(false), value = $bindable(), trigger, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let comboboxOpen = $state(false);

	let selectedValue = $state<string>();

	let toAdd = $state<string[]>([]);
	let toDelete = $state<CourseTagsModel>([]);
	let existingTags = $state<CourseTagsModel>([]);
	let availableTags = $state<string[]>([]);
	let filteredTags = $state<string[]>([]);

	let inputEl = $state<HTMLInputElement | null>(null);
	let inputValue = $state('');
	const inputDebounced = new Debounced(() => inputValue, 50);

	let tagsEl = $state<HTMLElement>();

	let isPosting = $state(false);

	const isArray = Array.isArray(value);

	let loadTagsPromise = $state<Promise<void>>();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// When the dialog is opened, load the existing tags
	$effect(() => {
		if (open) {
			toAdd = [];
			toDelete = [];
			existingTags = [];
			availableTags = [];
			filteredTags = [];
			inputValue = '';
			selectedValue = '';
			isPosting = false;

			loadTagsPromise = loadTags();
		}
	});

	// Open the combobox when there are tags to show
	$effect(() => {
		if (filteredTags.length === 0) {
			comboboxOpen = false;
		} else {
			comboboxOpen = true;
		}
	});

	// When the selectedValue is populated, add the tag
	$effect(() => {
		if (!selectedValue) return;

		addTag(selectedValue).then(() => {
			// Make sure the values are cleared
			selectedValue = '';
			inputValue = '';
			if (inputEl) inputEl.value = '';
		});
	});

	// After the input value has been debounced, filter the tags
	$effect(() => {
		filteredTags = filterTags(inputDebounced.current);
	});

	// As the inputValue changes, update the first entry in the filteredTags list, ignoring
	// the debounce logic
	$effect(() => {
		if (filteredTags.length > 0) {
			filteredTags[0] = inputValue;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Load all available tags. If this is a single course, load the existing tags for that course
	// as well
	async function loadTags(): Promise<void> {
		const flickerPromise = new Promise((resolve) => setTimeout(resolve, 200));

		try {
			if (isArray) {
				const [resp] = await Promise.all([GetTagNames(), flickerPromise]);
				availableTags = resp;
			} else {
				const [resp1, resp2] = await Promise.all([
					GetCourseTags(value.id),
					GetTagNames(),
					flickerPromise
				]);
				existingTags = resp1;
				availableTags = resp2;
			}
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function filterTags(filterOn: string): string[] {
		if (filterOn === '') return [];

		let selectedTags = availableTags.filter((t) =>
			t.toLowerCase().includes(filterOn.toLowerCase())
		);

		if (selectedTags.length === 0) return [];

		// filter out the selected tags that are already existing, in the toAdd list, or in the
		// toDelete list
		selectedTags = selectedTags.filter(
			(tag) =>
				!existingTags.find((t) => t.tag.toLowerCase() === tag.toLowerCase()) &&
				!toAdd.find((t) => t.toLowerCase() === tag.toLowerCase()) &&
				!toDelete.find((t) => t.tag.toLowerCase() === tag.toLowerCase())
		);

		if (selectedTags.length === 0) return [];

		// Add the current input (if it's not already in the list)
		if (!selectedTags.find((item) => item.toLowerCase() === filterOn.toLowerCase()))
			selectedTags.unshift(filterOn);

		return selectedTags;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function handleInput(e: KeyboardEvent) {
		// Handle an edge case where the user presses the down arrow key when there
		// are no tags to show
		if (e.key === 'ArrowDown' && filteredTags.length === 0) {
			e.preventDefault();
			return;
		}

		// Ignore any key presses when the combobox is open
		if (e.key !== 'Enter' || comboboxOpen) return;

		e.preventDefault();

		// When there are tags to add and there is nothing in the input, call createTags()
		if (toAdd.length !== 0 && !inputValue) {
			await createTags();
			return;
		}

		await addTag(inputValue);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Add a tag to the toAdd list. If the tag is already in the list, shake the existing tag
	async function addTag(tag: string): Promise<void> {
		const cleanTag = tag.trim();
		if (!cleanTag) return;

		// When the tag already exists in the list, shake the tag
		if (toAdd.includes(cleanTag) || existingTags.find((t) => t.tag === cleanTag)) {
			if (!tagsEl) return;

			const tagEl = tagsEl.querySelector(`[data-tag="${cleanTag}"]`);
			if (!tagEl || tagEl.classList.contains('animate-shake')) return;

			tagEl.classList.add('animate-shake');
			setTimeout(() => {
				tagEl.classList.remove('animate-shake');
			}, 1000);

			return;
		}

		toAdd.push(cleanTag);

		inputValue = '';
		if (inputEl) inputEl.value = '';
		selectedValue = '';

		return;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Add/remove tags to/from the course(s)
	async function createTags(): Promise<void> {
		isPosting = true;

		try {
			if (isArray) {
				// For each course, add the tags
				await Promise.all(
					value.map(async (c) => {
						await Promise.all(
							toAdd.map(async (tag) => {
								await CreateCourseTag(c.id, { tag });
							})
						);
					})
				);

				toast.success('Tags added');
			} else {
				// Add the tags for this course
				await Promise.all(
					toAdd.map(async (tag) => {
						await CreateCourseTag(value.id, { tag });
					})
				);

				// Delete the tags for this course
				await Promise.all(
					toDelete.map(async (tag) => {
						await DeleteCourseTag(value.id, tag.id);
					})
				);
				toast.success('Tags updated');
			}
			successFn?.();
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
		open = false;
	}
</script>

<Dialog.Root bind:open {trigger}>
	<Dialog.Content
		class="inline-flex max-w-md flex-col"
		onOpenAutoFocus={(e) => {
			e.preventDefault();
			inputEl?.focus();
		}}
		onCloseAutoFocus={(e) => {
			e.preventDefault();
		}}
	>
		<Combobox.Root
			type="single"
			bind:open={() => comboboxOpen, (newOpen) => (comboboxOpen = newOpen)}
			bind:value={selectedValue}
		>
			<Dialog.Header class="relative px-0">
				<Button
					class="absolute h-full w-auto cursor-text rounded-none bg-transparent px-3 hover:bg-transparent hover:brightness-100"
					onfocusin={() => {
						inputEl?.focus();
					}}
				>
					<ScanIcon class="text-foreground-alt-1 size-5" />
				</Button>

				<Combobox.Input
					bind:ref={inputEl}
					oninput={(e) => (inputValue = e.currentTarget.value)}
					onkeydown={handleInput}
					disabled={isPosting}
					class="bg-background-alt-2 focus:bg-background-alt-3 placeholder:text-foreground-alt-2 h-full w-full rounded-none px-2.5 ps-12 ring-0 duration-250 ease-in-out placeholder:tracking-wide focus:outline-none"
					placeholder="Add tag..."
					aria-label="Add a tag"
				/>

				<Combobox.Portal>
					{#if filteredTags.length > 0}
						<Combobox.Content
							class="bg-background border-background-alt-5 data-[side=bottom]:slide-in-from-top-2 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 z-[60] max-h-60 w-[calc(var(--bits-combobox-anchor-width)-1rem)] overflow-x-hidden overflow-y-auto rounded-lg py-3 outline-hidden select-none data-[side=bottom]:translate-y-1"
							sideOffset={10}
						>
							{#each filteredTags as tag, i (i + tag)}
								<Combobox.Item
									class="rounded-button data-highlighted:bg-background-alt-2 flex h-10 w-full items-center py-3 ps-9 pr-1.5 text-sm outline-hidden select-none"
									value={tag}
									label={tag}
								>
									{#snippet children()}
										{tag}
									{/snippet}
								</Combobox.Item>
							{/each}
						</Combobox.Content>
					{/if}
				</Combobox.Portal>
			</Dialog.Header>
		</Combobox.Root>

		<main
			bind:this={tagsEl}
			class="flex max-h-60 min-h-40 w-full flex-1 shrink-0 flex-wrap place-content-start gap-2.5 overflow-x-hidden overflow-y-auto p-5"
		>
			{#if !isArray}
				{#await loadTagsPromise}
					<div class="flex w-full items-center justify-center pt-3">
						<Spinner class="bg-foreground-alt-2 size-3" />
					</div>
				{:then _}
					{#each existingTags as tag}
						<Badge
							class={cn(
								'bg-background-alt-3 text-foreground h-6 p-0 text-sm select-none',
								toDelete.find((t) => t === tag) && 'text-foreground-alt-2'
							)}
							data-tag={tag.tag}
						>
							<span class="mt-px h-full px-2.5 font-semibold">
								{tag.tag}
							</span>

							<Button
								disabled={isPosting}
								class={cn(
									'border-background-alt-6 text-foreground hover:bg-background-alt-4 h-full rounded-none rounded-r-md border-l bg-transparent px-1 disabled:bg-transparent'
								)}
								onclick={() => {
									if (toDelete.find((t) => t === tag)) {
										toDelete = toDelete.filter((t) => t.tag !== tag.tag);
									} else {
										toDelete.push(tag);
									}
								}}
							>
								{#if toDelete.find((t) => t === tag)}
									<UndoIcon class="fill-foreground size-3 stroke-2" />
								{:else}
									<XIcon class="size-3 stroke-2" />
								{/if}
							</Button>
						</Badge>
					{/each}
				{:catch error}
					<div class="container-px flex w-full">
						<Oops class="pt-0" contentClass="border-0" message={error.message} />
					</div>
				{/await}
			{/if}

			{#each toAdd as tag}
				<Badge class="bg-background-success text-foreground h-6 p-0 text-sm" data-tag={tag}>
					<span class="mt-px h-full px-2.5 font-semibold">
						{tag}
					</span>

					<Button
						class="border-background-alt-3 text-foreground h-full rounded-none border-l bg-transparent px-1"
						onclick={() => {
							toAdd = toAdd.filter((t) => t !== tag);
						}}
					>
						<XIcon class="size-3 stroke-2" />
					</Button>
				</Badge>
			{/each}
		</main>

		<Dialog.Footer>
			<Dialog.CloseButton />

			<Button
				disabled={isPosting || (toAdd.length === 0 && toDelete.length === 0)}
				onclick={async () => {
					await createTags();
				}}
				class="h-10 w-25 py-2"
			>
				{#if !isPosting}
					{isArray ? 'Add' : 'Update'}
				{:else}
					<Spinner class="bg-foreground-alt-3 size-2" />
				{/if}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
