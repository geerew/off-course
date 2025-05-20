<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { CreateTag, GetTag } from '$lib/api/tag-api';
	import { Spinner } from '$lib/components';
	import { PlusIcon, ScanIcon, XIcon } from '$lib/components/icons';
	import { Badge, Button, Dialog, Drawer, Input } from '$lib/components/ui';
	import { remCalc } from '$lib/utils';
	import { toast } from 'svelte-sonner';
	import { innerWidth } from 'svelte/reactivity/window';
	import theme from 'tailwindcss/defaultTheme';

	type Props = {
		successFn?: () => void;
	};

	let { successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let open = $state(false);
	let toAdd = $state<string[]>([]);
	let toAddCount = $derived(toAdd.length);
	let currentTag = $state('');
	let isPosting = $state(false);
	let inputEl = $state<HTMLInputElement>();
	let tagsEl = $state<HTMLElement>();

	const mdBreakpoint = +theme.screens.md.replace('rem', '');
	let isDesktop = $derived(remCalc(innerWidth.current ?? 0) > mdBreakpoint);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			toAdd = [];
			currentTag = '';
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function handleTagInput(e: KeyboardEvent): Promise<void> {
		if (e.key !== 'Enter') return;

		e.preventDefault();

		const cleanTag = currentTag.trim();

		if (toAdd.length !== 0 && !cleanTag) {
			await addTags();
			return;
		}

		if (!cleanTag) return;

		// When the tag already exists in the list, shake the tag
		if (toAdd.includes(cleanTag)) {
			if (!tagsEl) return;

			const tagEl = tagsEl.querySelector(`[data-tag="${cleanTag}"]`);
			if (!tagEl || tagEl.classList.contains('animate-shake')) return;

			tagEl.classList.add('animate-shake');
			setTimeout(() => {
				tagEl.classList.remove('animate-shake');
			}, 1000);

			return;
		}

		// Check if the tag already exists in the database
		try {
			await GetTag(cleanTag);
			toast.error(`Tag already exists`);
			return;
		} catch (error) {
			if ((error as APIError).status !== 404) toast.error((error as APIError).message);
		}

		toAdd.push(cleanTag);
		currentTag = '';
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function addTags(): Promise<void> {
		isPosting = true;

		try {
			await Promise.all(toAdd.map((name) => CreateTag({ tag: name })));
			toast.success('Tags added');
			successFn?.();
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
		open = false;
	}
</script>

{#snippet trigger()}
	{#if isDesktop}
		<Dialog.Trigger class="flex h-10 w-auto flex-row items-center gap-2 px-5">
			<PlusIcon class="size-5 stroke-[1.5]" />
			Add Tags
		</Dialog.Trigger>
	{:else}
		<Drawer.Trigger class="flex h-10 w-auto flex-row items-center gap-2 px-5">
			<PlusIcon class="size-5 stroke-[1.5]" />
			Add Tags
		</Drawer.Trigger>
	{/if}
{/snippet}

{#snippet header()}
	<Button
		class="absolute h-full w-auto cursor-text rounded-none bg-transparent px-3 enabled:hover:bg-transparent"
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
		onkeydown={(e) => {
			handleTagInput(e);
		}}
	/>
{/snippet}

{#snippet contents()}
	<main
		bind:this={tagsEl}
		class="flex max-h-60 min-h-40 w-full flex-1 shrink-0 flex-wrap place-content-start gap-2.5 overflow-x-hidden overflow-y-auto p-5"
		data-vaul-no-drag=""
	>
		{#each toAdd as tag}
			<Badge
				class="bg-background-success text-foreground h-6 overflow-hidden p-0 text-sm"
				data-tag={tag}
			>
				<span class="mt-px h-full px-2.5 font-semibold">
					{tag}
				</span>

				<Button
					class="border-background-alt-3 text-foreground enabled:hover:bg-background-error h-full rounded-none border-l bg-transparent px-1"
					onclick={() => {
						toAdd = toAdd.filter((t) => t !== tag);
					}}
				>
					<XIcon class="size-3 stroke-2" />
				</Button>
			</Badge>
		{/each}
	</main>
{/snippet}

{#snippet action()}
	<Button disabled={isPosting || toAddCount === 0} onclick={addTags} class="h-10 w-25 py-2">
		{#if isPosting}
			<Spinner class="bg-background-alt-4 size-2" />
		{:else}
			Add
		{/if}
	</Button>
{/snippet}

{#if isDesktop}
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
			<Dialog.Header class="relative px-0">
				{@render header()}
			</Dialog.Header>

			{@render contents()}

			<Dialog.Footer>
				<Dialog.CloseButton>Close</Dialog.CloseButton>
				{@render action()}
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{:else}
	<Drawer.Root bind:open>
		{@render trigger()}

		<Drawer.Content handleClass="bg-background-alt-4">
			<Drawer.Header class="border-y px-0">
				{@render header()}
			</Drawer.Header>

			{@render contents()}

			<Drawer.Footer>
				<Drawer.CloseButton>Close</Drawer.CloseButton>
				{@render action()}
			</Drawer.Footer>
		</Drawer.Content>
	</Drawer.Root>
{/if}
