<script lang="ts">
	import { SortAscendingIcon, SortDescendingIcon, WarningIcon } from '$lib/components/icons';
	import RightChevron from '$lib/components/icons/right-chevron.svelte';
	import { Dropdown } from '$lib/components/ui';
	import type { SortColumns, SortDirection } from '$lib/types/sort';

	type Props = {
		columns: SortColumns;
		selectedColumn: string;
		selectedDirection: SortDirection;
		disabled?: boolean;
		onUpdate?: () => void;
	};

	let {
		columns,
		selectedColumn = $bindable(),
		selectedDirection = $bindable(),
		disabled = false,
		onUpdate
	}: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	selectedDirection = selectedDirection || 'asc';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// If a selected column is not provided or the selected column is not in the list of columns,
	// select the first column
	$effect(() => {
		if (columns.length === 0) return;

		if (!selectedColumn || !columns.find((column) => column.column === selectedColumn)) {
			selectedColumn = columns[0].column;
		}
	});
</script>

<Dropdown.Root>
	<Dropdown.Trigger class="w-36 [&[data-state=open]>svg]:rotate-90" {disabled}>
		{#if columns.length === 0}
			<div class="flex items-center gap-1.5">
				<WarningIcon class="size-4 stroke-[1.5]" />
				<span>No columns</span>
			</div>
			<RightChevron class="stroke-foreground-alt-3 size-4.5 duration-200" />
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
			<RightChevron class="stroke-foreground-alt-3 size-4.5 duration-200" />
		{/if}
	</Dropdown.Trigger>

	<Dropdown.Content class="min-w-40">
		{#if columns.length === 0}
			<div class="flex flex-col gap-1.5 p-1">
				<p class="text-foreground-alt-3 text-center text-sm">No columns</p>
			</div>
		{:else}
			<div class="flex flex-col gap-1.5">
				<Dropdown.RadioGroup bind:value={selectedColumn}>
					{#each columns as column}
						<Dropdown.RadioItem
							value={column.column}
							onclick={() => {
								if (selectedColumn === column.column) return;
								onUpdate?.();
							}}
						>
							{column.label}
						</Dropdown.RadioItem>
					{/each}
				</Dropdown.RadioGroup>
			</div>

			<Dropdown.Separator />

			<div class="flex flex-col gap-1.5">
				<Dropdown.RadioGroup bind:value={selectedDirection}>
					<Dropdown.RadioItem
						value="asc"
						onclick={() => {
							if (selectedDirection === 'asc') return;
							onUpdate?.();
						}}
					>
						<SortAscendingIcon class="text-foreground-alt-3 size-4" />
						{columns.find((column) => column.column === selectedColumn)?.asc || 'Ascending'}
					</Dropdown.RadioItem>

					<Dropdown.RadioItem
						value="desc"
						onclick={() => {
							if (selectedDirection === 'desc') return;
							onUpdate?.();
						}}
					>
						<SortDescendingIcon class="text-foreground-alt-3 size-4" />
						{columns.find((column) => column.column === selectedColumn)?.desc || 'Descending'}
					</Dropdown.RadioItem>
				</Dropdown.RadioGroup>
			</div>
		{/if}
	</Dropdown.Content>
</Dropdown.Root>
