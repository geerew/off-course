<script lang="ts">
	import { SortMenu } from '$lib/components';
	import type { SortColumns, SortDirection } from '$lib/types/sort';
	import { tick } from 'svelte';

	let sortColumns = [
		{ label: 'Title', column: 'courses.title', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Available', column: 'courses.available', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Added', column: 'courses.created_at', asc: 'Oldest', desc: 'Newest' },
		{ label: 'Updated', column: 'courses.updated_at', asc: 'Oldest', desc: 'Newest' }
	] as const satisfies SortColumns;

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		value?: string;
		defaultColumn?: (typeof sortColumns)[number]['column'];
		defaultDirection?: SortDirection;
		disabled?: boolean;
		onApply: () => void;
	};

	let {
		value = $bindable(''),
		defaultColumn = sortColumns[0].column,
		defaultDirection = 'asc',
		disabled = false,
		onApply
	}: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let selectedSortColumn: (typeof sortColumns)[number]['column'] = $state(defaultColumn);
	let selectedSortDirection: SortDirection = $state(defaultDirection);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the default sort value when the component is initialized
	$effect(() => {
		if (defaultColumn && defaultDirection && (value === '' || value === undefined)) {
			value = `sort:"${defaultColumn} ${defaultDirection}"`;
		}
	});
</script>

<div class="flex h-10 items-center gap-3 rounded-lg">
	<SortMenu
		{disabled}
		columns={sortColumns}
		bind:selectedColumn={selectedSortColumn}
		bind:selectedDirection={selectedSortDirection}
		onUpdate={async () => {
			await tick();
			value = `sort:"${selectedSortColumn} ${selectedSortDirection}"`;
			await onApply();
		}}
	/>
</div>
