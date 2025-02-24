<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { CreateCourseTag, DeleteCourseTag, GetCourseTags } from '$lib/api/course-api';
	import { Oops } from '$lib/components/';
	import { ScanIcon, UndoIcon, XIcon } from '$lib/components/icons';
	import { Badge, Button, Dialog, Input } from '$lib/components/ui';
	import type { CourseModel, CoursesModel, CourseTagsModel } from '$lib/models/course-model';
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';
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

	let toAdd = $state<string[]>([]);
	let toDelete = $state<CourseTagsModel>([]);
	let currentTags = $state<CourseTagsModel>([]);
	let currentTag = $state('');

	let inputEl = $state<HTMLInputElement>();
	let tagsEl = $state<HTMLElement>();

	let isPosting = $state(false);

	const isArray = Array.isArray(value);

	let loadCurrentTagsPromise = $state<Promise<void>>();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			loadCurrentTagsPromise = loadCurrentTags();
		}
	});
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function loadCurrentTags(): Promise<void> {
		if (isArray) return;

		try {
			const flickerPromise = new Promise((resolve) => setTimeout(resolve, 200));
			const [response] = await Promise.all([GetCourseTags(value.id), flickerPromise]);
			currentTags = response;
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function handleTagInput(e: KeyboardEvent): Promise<void> {
		if (e.key !== 'Enter') return;

		e.preventDefault();

		const cleanTag = currentTag.trim();

		if (toAdd.length !== 0 && !cleanTag) {
			await updateTags();
			return;
		}

		if (!cleanTag) return;

		// When the tag already exists in the list, shake the tag
		if (toAdd.includes(cleanTag) || currentTags.find((t) => t.tag === cleanTag)) {
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
		currentTag = '';
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function updateTags(): Promise<void> {
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

<Dialog.Root
	bind:open
	onOpenChange={() => {
		isPosting = false;
	}}
	{trigger}
>
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
		<Dialog.Header class="relative px-0">
			<Button
				class="absolute h-full w-auto cursor-text rounded-none bg-transparent px-3 hover:bg-transparent hover:brightness-100"
				onfocusin={() => {
					inputEl?.focus();
				}}
			>
				<ScanIcon class="text-foreground-alt-1 size-5" />
			</Button>

			<Input
				type="text"
				bind:ref={inputEl}
				bind:value={currentTag}
				placeholder="Add tag..."
				class="bg-background-alt-2 focus:bg-background-alt-3 h-full rounded-none ps-12"
				disabled={isPosting}
				onkeydown={(e) => {
					handleTagInput(e);
				}}
			/>
		</Dialog.Header>

		<main
			bind:this={tagsEl}
			class="flex max-h-60 min-h-40 w-full flex-1 shrink-0 flex-wrap place-content-start gap-2.5 overflow-x-hidden overflow-y-auto p-5"
		>
			{#if !isArray}
				{#await loadCurrentTagsPromise}
					<div class="flex w-full items-center justify-center pt-3">
						<Spinner class="bg-foreground-alt-2 size-3" />
					</div>
				{:then _}
					{#each currentTags as tag}
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
				onclick={updateTags}
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
