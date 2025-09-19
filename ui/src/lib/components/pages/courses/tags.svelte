<script lang="ts">
	import { GetTagNames } from '$lib/api/tag-api';
	import { Spinner } from '$lib/components';
	import { RightChevronIcon, SearchIcon, TagIcon, WarningIcon, XIcon } from '$lib/components/icons';
	import { Dropdown, Input } from '$lib/components/ui';
	import Button from '$lib/components/ui/button.svelte';
	import { cn } from '$lib/utils';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		value?: string;
		defaultTags?: string[];
		disabled?: boolean;
		onApply: () => void;
	};

	let { value = $bindable(''), defaultTags = [], disabled = false, onApply }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let allTags = $state<string[]>([]);

	let tagsToRender = $state<string[]>([]);

	let selected = $state<string[]>(defaultTags);

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
			console.log('allTags', allTags);
		} catch (error) {
			throw error;
		}
	}
</script>

<div class="flex h-10 items-center gap-3 rounded-lg">
	<Dropdown.Root>
		<Dropdown.Trigger
			class={cn(
				'w-36 [&[data-state=open]>svg]:rotate-90 ',
				value &&
					'data-[state=open]:border-b-background-primary-alt-1 hover:border-b-background-primary-alt-1 border-b-background-primary-alt-1 border-b-2'
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

				{#await loadPromise}
					<div class="flex justify-center pt-10">
						<Spinner class="bg-foreground-alt-3 size-4" />
					</div>
				{:then _}
					{#if allTags.length === 0}
						<div class="text-foreground-alt-3 py-10 text-sm">No tags</div>
					{:else}
						<!--  Search -->
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

						<!-- Tags -->
						{#if tagsToRender.length === 0}
							<div class="text-foreground-alt-3 py-3 text-center text-sm">No matching tags</div>
						{:else}
							<Dropdown.CheckboxGroup
								bind:value={selected}
								onValueChange={() => {
									value = selected.map((v) => `tag:"${v}"`).join(' OR ');
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
					<div class="flex w-full flex-col items-center gap-2 pt-10">
						<WarningIcon class="text-foreground-error size-10" />
						<span class="text-lg">Failed to load tags: {error.message}</span>
					</div>
				{/await}
			</div>
		</Dropdown.Content>
	</Dropdown.Root>
</div>
