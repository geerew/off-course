<script lang="ts">
	import {
		SortAscendingIcon,
		SortDescendingIcon,
		TickIcon,
		WarningIcon
	} from '$lib/components/icons';
	import RightChevron from '$lib/components/icons/right-chevron.svelte';
	import { Dropdown } from '$lib/components/ui';
	import type { SortColumns, SortDirection } from '$lib/types/sort';
	import { DropdownMenu } from 'bits-ui';

	type Props = {
		columns: SortColumns;
		selectedColumn: string;
		selectedDirection: SortDirection;
		onUpdate?: () => void;
	};

	let {
		columns,
		selectedColumn = $bindable(),
		selectedDirection = $bindable(),
		onUpdate
	}: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	selectedDirection = selectedDirection || 'asc';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// If a selected column is not provided or the selected column is not in the list of columns,
	// select the first column
	$effect(() => {
		if (columns.length === 0) return;

		if (!selectedColumn || !columns.find((column) => column.column === selectedColumn)) {
			selectedColumn = columns[0].column;
		}
	});
</script>

<Dropdown triggerClass="w-36 [&[data-state=open]>svg]:rotate-90" contentClass="min-w-40">
	{#snippet trigger()}
		{#if columns.length === 0}
			<div class="flex items-center gap-1.5">
				<WarningIcon class="size-4 stroke-[1.5]" />
				<span>No columns</span>
			</div>
			<RightChevron class="stroke-foreground-alt-2 size-4.5 duration-200" />
		{:else}
			<div class="flex items-center gap-1.5">
				{#if selectedDirection === 'asc'}
					<SortAscendingIcon class="size-4 stroke-[1.5]" />
				{:else}
					<SortDescendingIcon class="size-4 stroke-[1.5]" />
				{/if}
				{#if columns.find((column) => column.column === selectedColumn)}
					{columns.find((column) => column.column === selectedColumn)?.label}
				{:else}
					<span>Sort</span>
				{/if}
			</div>
			<RightChevron class="stroke-foreground-alt-2 size-4.5 duration-200" />
		{/if}
	{/snippet}

	{#snippet content()}
		{#if columns.length === 0}
			<div class="flex flex-col gap-1.5 p-1">
				<p class="text-foreground-alt-2 text-center text-sm">No columns</p>
			</div>
		{:else}
			<div class="flex flex-col gap-1.5 p-1">
				<DropdownMenu.RadioGroup bind:value={selectedColumn}>
					{#each columns as column}
						<DropdownMenu.RadioItem
							class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 text-sm duration-200 select-none"
							value={column.column}
							onclick={() => {
								if (selectedColumn === column.column) return;
								onUpdate?.();
							}}
						>
							{#snippet children({ checked })}
								{#if checked}
									<TickIcon class="text-foreground-alt-2 size-3.5 stroke-2" />
								{:else}
									<span class="size-3.5"></span>
								{/if}

								{column.label}
							{/snippet}
						</DropdownMenu.RadioItem>
					{/each}
				</DropdownMenu.RadioGroup>
			</div>
			<DropdownMenu.Separator class="bg-background-alt-5 h-px w-full" />
			<div class="flex flex-col gap-1.5 p-1">
				<DropdownMenu.RadioGroup bind:value={selectedDirection}>
					<DropdownMenu.RadioItem
						class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 text-sm duration-200 select-none"
						value="asc"
						onclick={() => {
							if (selectedDirection === 'asc') return;
							onUpdate?.();
						}}
					>
						{#snippet children({ checked })}
							{#if checked}
								<TickIcon class="text-foreground-alt-2 size-3.5 stroke-2" />
							{:else}
								<span class="size-3.5"></span>
							{/if}

							<SortAscendingIcon class="text-foreground-alt-2 size-4" />

							{columns.find((column) => column.column === selectedColumn)?.asc || 'Ascending'}
						{/snippet}
					</DropdownMenu.RadioItem>

					<DropdownMenu.RadioItem
						class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 text-sm duration-200 select-none"
						value="desc"
						onclick={() => {
							if (selectedDirection === 'desc') return;
							onUpdate?.();
						}}
					>
						{#snippet children({ checked })}
							{#if checked}
								<TickIcon class="text-foreground-alt-2 size-3.5 stroke-2" />
							{:else}
								<span class="size-3.5"></span>
							{/if}

							<SortDescendingIcon class="text-foreground-alt-2 size-4" />

							{columns.find((column) => column.column === selectedColumn)?.desc || 'Descending'}
						{/snippet}
					</DropdownMenu.RadioItem>
				</DropdownMenu.RadioGroup>
			</div>
		{/if}
	{/snippet}
</Dropdown>
