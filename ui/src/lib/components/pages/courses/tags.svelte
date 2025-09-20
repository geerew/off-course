<script lang="ts">
	import { GetTagNames } from '$lib/api/tag-api';
	import { Spinner } from '$lib/components';
	import {
		RightChevronIcon,
		SearchIcon,
		TagIcon,
		TickIcon,
		WarningIcon,
		XIcon
	} from '$lib/components/icons';
	import { Dropdown, Input } from '$lib/components/ui';
	import Button from '$lib/components/ui/button.svelte';
	import { cn } from '$lib/utils';
	import { Accordion, Checkbox, Label, useId } from 'bits-ui';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		type: 'dropdown' | 'accordion';
		value?: string;
		selected?: string[];
		disabled?: boolean;
		onApply: () => void;
	};

	let {
		type,
		value = $bindable(''),
		selected = $bindable([]),
		disabled = false,
		onApply
	}: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let allTags = $state<string[]>([]);

	let tagsToRender = $state<string[]>([]);

	// Reference to the search input element
	let searchTagsEl = $state<HTMLInputElement>();
	let searchValue = $state('');

	let loadPromise = $state(fetcher());

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (!searchValue) {
			tagsToRender = allTags;
			return;
		}

		const lowerSearch = searchValue.toLowerCase();
		tagsToRender = allTags.filter((tag) => tag.toLowerCase().includes(lowerSearch));
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function fetcher(): Promise<void> {
		try {
			allTags = await GetTagNames();
			tagsToRender = allTags;
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// As the selected tags change, update the value
	$effect(() => {
		value = selected.map((v) => `tag:"${v}"`).join(' OR ');
	});
</script>

{#snippet title()}
	<div class="flex flex-row items-center justify-between px-1.5">
		<span class="text-background-primary-alt-1 text-base font-semibold">Tags</span>
		<Button
			variant="ghost"
			class="text-foreground-alt-3 hover:text-foreground-alt-2 p-0 text-sm hover:bg-transparent"
			onclick={() => {
				selected = [];
				value = '';
				onApply();
			}}
		>
			clear
		</Button>
	</div>
{/snippet}

{#snippet search()}
	<div class="group relative mb-2 w-full">
		<Button
			variant="ghost"
			class="text-foreground-alt-3 group-focus-within:text-foreground-alt-1 absolute top-1/2 left-2 -translate-y-1/2 transform cursor-text rounded-full p-0 hover:bg-transparent"
			{disabled}
			onclick={() => {
				if (!searchTagsEl) return;
				searchTagsEl.focus();
			}}
		>
			<SearchIcon class="size-3.5" />
		</Button>

		<Input
			bind:ref={searchTagsEl}
			bind:value={searchValue}
			placeholder="Search tags..."
			class="placeholder:text-foreground-alt-3 bg-background-alt-2 focus:bg-alt-3 h-8 border-b-2 ps-8 pe-8 text-sm placeholder:text-xs"
			{disabled}
		/>

		{#if searchValue}
			<Button
				variant="ghost"
				class="hover:bg-background-alt-2 text-foreground-alt-2 hover:text-foreground absolute top-1/2 right-1 h-auto -translate-y-1/2 transform rounded-md p-1"
				onclick={() => {
					searchValue = '';
				}}
			>
				<XIcon class="size-4" />
			</Button>
		{/if}
	</div>
{/snippet}

{#snippet awaitLoader()}
	<div class="flex justify-center pt-5">
		<Spinner class="bg-foreground-alt-3 size-4" />
	</div>
{/snippet}

{#snippet awaitError(error: Error)}
	<div class="flex w-full flex-col items-center gap-2 pt-10">
		<WarningIcon class="text-foreground-error size-10" />
		<span class="text-lg">Failed to load tags: {error.message}</span>
	</div>
{/snippet}

{#if type === 'dropdown'}
	<div class="flex h-10 items-center gap-3 rounded-lg">
		<Dropdown.Root>
			<Dropdown.Trigger
				class={cn(
					'relative w-36 [&[data-state=open]>svg]:rotate-90 ',
					value &&
						'after:bg-background-primary-alt-1 after:absolute after:bottom-0 after:left-0 after:h-0.5 after:w-full after:rounded-b-lg'
				)}
				{disabled}
			>
				<div class="flex items-center gap-1.5">
					<TagIcon class="size-4 stroke-2" />
					<span>Tags</span>
				</div>

				<RightChevronIcon class="stroke-foreground-alt-3 size-4.5 duration-200" />
			</Dropdown.Trigger>

			<Dropdown.Content class="max-h-70 w-60 overflow-y-scroll" align="start">
				<div class="flex flex-col gap-1">
					{@render title()}

					{#await loadPromise}
						{@render awaitLoader()}
					{:then _}
						{#if allTags.length === 0}
							<div class="text-foreground-alt-3 py-5 text-center text-sm">No tags</div>
						{:else}
							<!--  Search -->
							{@render search()}

							<!-- Tags -->
							{#if tagsToRender.length === 0}
								<div class="text-foreground-alt-3 py-3 text-center text-sm">No matching tags</div>
							{:else}
								<Dropdown.CheckboxGroup
									bind:value={selected}
									onValueChange={() => {
										onApply();
									}}
								>
									{#each tagsToRender as tag}
										<Dropdown.CheckboxItem value={tag} closeOnSelect={false}>
											{tag}
										</Dropdown.CheckboxItem>
									{/each}
								</Dropdown.CheckboxGroup>
							{/if}
						{/if}
					{:catch error}
						{@render awaitError(error)}
					{/await}
				</div>
			</Dropdown.Content>
		</Dropdown.Root>
	</div>
{:else}
	<Accordion.Item value="tags" class="bg-background-alt-1 overflow-hidden rounded-lg">
		<Accordion.Header>
			<Accordion.Trigger
				class={cn(
					'group data-[state=open]:border-b-foreground-alt-4 flex w-full flex-1 items-center justify-between border-b border-transparent px-2.5 py-2.5 font-medium transition-transform select-none hover:cursor-pointer',
					value &&
						'data-[state=open]:border-b-background-primary-alt-1 data-[state=closed]:border-b-background-primary-alt-1 data-[state=closed]:border-b-2'
				)}
			>
				<div class="flex items-center gap-1.5">
					<TagIcon class="size-6 stroke-2" />
					<span class="w-full text-left">Tags</span>
				</div>

				<div class="flex flex-row items-center gap-3">
					<Button
						variant="ghost"
						class={cn(
							'text-foreground-alt-3 hover:text-foreground-alt-2 p-0 text-sm hover:bg-transparent',
							!value && 'invisible'
						)}
						onclick={(e: MouseEvent) => {
							e.preventDefault();
							e.stopPropagation();
							selected = [];
							value = '';
							onApply();
						}}
					>
						clear
					</Button>

					<RightChevronIcon
						class="size-4.5 stroke-2 transition-transform duration-100 group-data-[state=open]:rotate-90"
					/>
				</div>
			</Accordion.Trigger>
		</Accordion.Header>

		<Accordion.Content
			class="data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down max-h-72 overflow-hidden overflow-y-scroll px-2.5 py-3 text-sm tracking-[-0.01em]"
		>
			{#await loadPromise}
				{@render awaitLoader()}
			{:then _}
				{#if allTags.length === 0}
					<div class="text-foreground-alt-3 py-5 text-center text-sm">No tags</div>
				{:else}
					<!--  Search -->
					{@render search()}

					<!-- Tags -->
					{#if tagsToRender.length === 0}
						<div class="text-foreground-alt-3 py-5 text-center text-sm">No matching tags</div>
					{:else}
						<Checkbox.Group
							class="flex flex-col"
							bind:value={selected}
							name="tags"
							onValueChange={() => {
								onApply();
							}}
						>
							{#each tagsToRender as tag}
								{@const id = useId()}
								<div
									class="hover:bg-background-alt-3 flex flex-row items-center overflow-hidden rounded-md hover:cursor-pointer"
								>
									<Checkbox.Root
										{id}
										aria-labelledby="{id}-label"
										class="inline-flex size-3.5 h-full shrink-0 items-center justify-center py-1.5 hover:cursor-pointer"
										name={tag}
										value={tag}
									>
										{#snippet children({ checked })}
											<div class="inline-flex pl-2.5">
												{#if checked}
													<TickIcon class="size-3.5 stroke-2" />
												{:else}
													<span class="size-3.5"></span>
												{/if}
											</div>
										{/snippet}
									</Checkbox.Root>

									<Label.Root
										id="{id}-label"
										for={id}
										class="inline-flex w-full py-1 pr-1.5 pl-3.5 text-sm select-none hover:cursor-pointer"
									>
										{tag}
									</Label.Root>
								</div>
							{/each}
						</Checkbox.Group>
					{/if}
				{/if}
			{:catch error}
				{@render awaitError(error)}
			{/await}
		</Accordion.Content>
	</Accordion.Item>
{/if}
