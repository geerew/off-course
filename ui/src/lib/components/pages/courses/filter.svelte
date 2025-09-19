<script lang="ts">
	import Progress from './progress.svelte';
	import Search from './search.svelte';
	import Sort from './sort.svelte';
	import Tags from './tags.svelte';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		filter?: string;
		disabled?: boolean;
		onApply: () => void | Promise<void>;
	};

	let { filter = $bindable(''), disabled = false, onApply }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let appliedFilter = $state(filter);

	let searchCourses = $state('');
	let sort = $state('');
	let progress = $state('');
	let tags = $state('');

	$inspect('filter', filter);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function applyFilter() {
		let tmpFilter = sort;

		if (searchCourses) tmpFilter += ` "${searchCourses}"`;
		if (progress) tmpFilter += ` AND (${progress})`;
		if (tags) tmpFilter += ` AND (${tags})`;

		// Do nothing when the value hasn't changed
		if (tmpFilter === appliedFilter) return;

		appliedFilter = tmpFilter;
		filter = tmpFilter;

		await onApply();
	}
</script>

<div class="flex w-full flex-1 justify-between">
	<div class="flex flex-1 gap-5">
		<Search bind:value={searchCourses} {disabled} onApply={applyFilter} />
		<Tags bind:value={tags} {disabled} onApply={applyFilter} />
		<Progress bind:value={progress} {disabled} onApply={applyFilter} />
	</div>

	<Sort
		bind:value={sort}
		defaultColumn="courses.title"
		defaultDirection="asc"
		{disabled}
		onApply={applyFilter}
	/>
</div>
