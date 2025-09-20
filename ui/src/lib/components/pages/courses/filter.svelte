<script lang="ts">
	import { FilterIcon } from '$lib/components/icons';
	import { Button } from '$lib/components/ui';
	import type { SortDirection } from '$lib/types/sort';
	import { cn, remCalc } from '$lib/utils';
	import { Accordion, Dialog } from 'bits-ui';
	import { innerWidth } from 'svelte/reactivity/window';
	import theme from 'tailwindcss/defaultTheme';
	import Progress, { type ProgressState } from './progress.svelte';
	import Search from './search.svelte';
	import Sort, { type SortColumn } from './sort.svelte';
	import Tags from './tags.svelte';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		filter?: string;
		disabled?: boolean;
		onApply: () => void | Promise<void>;
	};

	let { filter = $bindable(''), disabled = false, onApply }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let menuPopupMode = $state(false);
	let windowWidth = $derived(remCalc(innerWidth.current ?? 0));
	let dialogOpen = $state(false);

	let appliedFilter = $state(filter);

	let searchCourses = $state('');

	let sort = $state('');
	let selectedSortColumn: SortColumn = $state('courses.title');
	let selectedSortDirection: SortDirection = $state('asc');

	let tags = $state('');
	let selectedTags = $state<string[]>([]);

	let progress = $state('');
	let selectedProgress = $state<ProgressState[]>([]);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function applyFilter() {
		// Do nothing when the value hasn't changed
		if (filter === appliedFilter) return;
		appliedFilter = filter;
		await onApply();
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		let tmpFilter = sort;

		if (searchCourses) tmpFilter += ` "${searchCourses}"`;
		if (progress) tmpFilter += ` AND (${progress})`;
		if (tags) tmpFilter += ` AND (${tags})`;
		filter = tmpFilter;
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the menu popup mode based on the screen size
	$effect(() => {
		menuPopupMode = windowWidth >= +theme.screens.xl.replace('rem', '') ? false : true;
		if (!menuPopupMode) dialogOpen = false;
	});
</script>

<div class="flex w-full flex-1 justify-between">
	{#if menuPopupMode}
		<div class="flex flex-1 gap-5">
			<Search bind:value={searchCourses} {disabled} onApply={applyFilter} />

			<Button
				class={cn(
					'relative flex h-10 w-auto flex-row items-center gap-2 px-4 py-0',
					(progress || tags) &&
						'after:bg-background-primary-alt-1 after:absolute after:bottom-0 after:left-0 after:h-0.5 after:w-full after:rounded-b-lg'
				)}
				variant="outline"
				{disabled}
				onclick={() => {
					dialogOpen = true;
				}}
			>
				<FilterIcon class="size-4 stroke-2" />
				Filter
			</Button>

			<Dialog.Root bind:open={dialogOpen}>
				<Dialog.Overlay
					class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/60"
				/>

				<Dialog.Content
					class="border-foreground-alt-4 bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left fixed top-0 left-0 z-50 h-full w-[var(--settings-menu-width)] border-r px-4 pt-4"
				>
					<nav class="flex h-full w-full flex-col gap-3 overflow-x-hidden overflow-y-auto pb-8">
						<!-- Title -->
						<div class="flex flex-row items-center justify-between px-1.5">
							<div
								class="text-background-primary-alt-1 flex flex-row items-center gap-2 text-lg font-semibold"
							>
								<FilterIcon class="size-5 stroke-2" />
								<span class="text-base font-semibold">Filter</span>
							</div>

							<Button
								variant="ghost"
								class={cn(
									'text-foreground-alt-3 hover:text-foreground-alt-2 p-0 text-sm hover:bg-transparent',
									!tags && !progress && 'invisible'
								)}
								onclick={() => {
									tags = '';
									selectedTags = [];
									progress = '';
									selectedProgress = [];
									onApply();
								}}
							>
								clear
							</Button>
						</div>

						<Accordion.Root class="flex w-full flex-col gap-5 pt-5" type="multiple">
							<Tags
								type="accordion"
								bind:value={tags}
								bind:selected={selectedTags}
								onApply={applyFilter}
							/>
							<Progress
								type="accordion"
								bind:value={progress}
								bind:selected={selectedProgress}
								onApply={applyFilter}
							/>

							<Sort
								type="accordion"
								bind:value={sort}
								bind:selectedColumn={selectedSortColumn}
								bind:selectedDirection={selectedSortDirection}
								{disabled}
								onApply={applyFilter}
							/>
						</Accordion.Root>
					</nav>
				</Dialog.Content>
			</Dialog.Root>
		</div>
	{:else}
		<div class="flex flex-1 gap-5">
			<Search bind:value={searchCourses} {disabled} onApply={applyFilter} />
			<Tags
				type="dropdown"
				bind:value={tags}
				bind:selected={selectedTags}
				{disabled}
				onApply={applyFilter}
			/>
			<Progress
				type="dropdown"
				bind:value={progress}
				bind:selected={selectedProgress}
				{disabled}
				onApply={applyFilter}
			/>
		</div>

		<Sort
			type="dropdown"
			bind:value={sort}
			bind:selectedColumn={selectedSortColumn}
			bind:selectedDirection={selectedSortDirection}
			{disabled}
			onApply={applyFilter}
		/>
	{/if}
</div>
