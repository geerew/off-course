<!-- TODO have a columns dropdown to hide show columns -->
<!-- TODO store selection state in localstorage -->
<script lang="ts">
	import { GetTags } from '$lib/api/tag-api';
	import { FilterBar, Pagination, SortMenu } from '$lib/components';
	import { AddTagsDialog } from '$lib/components/dialogs';
	import { WarningIcon } from '$lib/components/icons';
	import RowActionMenu from '$lib/components/pages/admin/tags/row-action-menu.svelte';
	import TableActionMenu from '$lib/components/pages/admin/tags/table-action-menu.svelte';
	import Spinner from '$lib/components/spinner.svelte';
	import { Checkbox } from '$lib/components/ui';
	import * as Table from '$lib/components/ui/table';
	import type { TagModel, TagsModel } from '$lib/models/tag-model';
	import type { SortColumns, SortDirection } from '$lib/types/sort';
	import { PersistedState } from 'runed';
	import { tick } from 'svelte';
	import { toast } from 'svelte-sonner';

	let tags: TagsModel = $state([]);

	let filterValue = $state('');
	let filterAppliedValue = $state('');

	let selectedTags: Record<string, TagModel> = $state({});
	let selectedTagsCount = $derived(Object.keys(selectedTags).length);

	let paginationPage = $state(1);
	let paginationPerPage = $state(10);
	let paginationTotal = $state(0);

	let sortColumns = [
		{ label: 'Tag', column: 'tags.tag', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Courses', column: 'course_count', asc: 'Lowest', desc: 'Highest' }
	] as const satisfies SortColumns;

	type PersistedState = {
		sort: {
			column: (typeof sortColumns)[number]['column'];
			direction: SortDirection;
		};
	};

	const persistedState = new PersistedState<PersistedState>('admin_tags', {
		sort: { column: 'tags.tag', direction: 'desc' }
	});

	let selectedSortColumn = $state(persistedState.current.sort.column);
	let selectedSortDirection = $state(persistedState.current.sort.direction);

	let isIndeterminate = $derived(selectedTagsCount > 0 && selectedTagsCount < paginationTotal);
	let isChecked = $derived(selectedTagsCount !== 0 && selectedTagsCount === paginationTotal);

	let loadPromise = $state(fetchTags());

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function fetchTags(): Promise<void> {
		try {
			const sort = `sort:"${selectedSortColumn} ${selectedSortDirection}"`;
			const q = filterValue ? `${filterValue} ${sort}` : sort;

			const data = await GetTags({
				q,
				page: paginationPage,
				perPage: paginationPerPage
			});
			paginationTotal = data.totalItems;
			tags = data.items;
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function onRowDelete(numDeleted: number) {
		const remainingTotal = paginationTotal - numDeleted;
		const totalPages = Math.max(1, Math.ceil(remainingTotal / paginationPerPage));

		if (paginationPage > totalPages) {
			paginationPage = totalPages;
		}

		loadPromise = fetchTags();
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function onCheckboxClicked(e: MouseEvent) {
		e.preventDefault();

		const allTagsSelectedOnPage = tags.every((t) => {
			return selectedTags[t.id] !== undefined;
		});

		if (allTagsSelectedOnPage) {
			tags.forEach((t) => {
				delete selectedTags[t.id];
			});
		} else {
			tags.forEach((t) => {
				selectedTags[t.id] = t;
			});
		}

		toastCount();
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function toastCount() {
		if (tags.length === 0) return;

		if (selectedTagsCount === 0) {
			toast.success('No tags selected');
		} else {
			toast.success(`${selectedTagsCount} row${selectedTagsCount > 1 ? 's' : ''} selected`);
		}
	}
</script>

<div class="flex w-full place-content-center">
	<div class="flex w-full max-w-4xl flex-col gap-6 pt-1">
		<div class="flex flex-row items-center justify-between">
			<AddTagsDialog
				successFn={() => {
					loadPromise = fetchTags();
				}}
			/>
		</div>

		<div class="flex flex-col gap-3 md:flex-row">
			<div class="flex flex-1 flex-row">
				<FilterBar
					bind:value={filterValue}
					disabled={!filterAppliedValue && tags.length === 0}
					onApply={async () => {
						if (filterValue !== filterAppliedValue) {
							filterAppliedValue = filterValue;
							paginationPage = 1;
							loadPromise = fetchTags();
						}
					}}
				/>
			</div>

			<div class="flex flex-row justify-end gap-3">
				<div class="flex h-10 items-center gap-3 rounded-lg">
					<TableActionMenu
						bind:tags={selectedTags}
						onDelete={() => {
							const numDeleted = Object.keys(selectedTags).length;
							selectedTags = {};
							onRowDelete(numDeleted);
						}}
					/>
				</div>

				<div class="flex h-10 items-center gap-3 rounded-lg">
					<SortMenu
						columns={sortColumns}
						bind:selectedColumn={selectedSortColumn}
						bind:selectedDirection={selectedSortDirection}
						onUpdate={async () => {
							await tick();

							persistedState.current = {
								...persistedState.current,
								sort: {
									column: selectedSortColumn,
									direction: selectedSortDirection
								}
							};

							loadPromise = fetchTags();
						}}
					/>
				</div>
			</div>
		</div>

		<div class="flex w-full place-content-center">
			{#await loadPromise}
				<div class="flex justify-center pt-10">
					<Spinner class="bg-foreground-alt-3 size-4" />
				</div>
			{:then _}
				<div class="flex w-full flex-col gap-8">
					<Table.Root class="grid-cols-[3.5rem_1fr_auto_3.5rem]">
						<Table.Thead>
							<Table.Tr class="text-xs font-semibold uppercase">
								<!-- Checkbox-->
								<Table.Th>
									<Checkbox
										indeterminate={isIndeterminate}
										checked={isChecked}
										onclick={onCheckboxClicked}
									/>
								</Table.Th>

								<!-- Tag -->
								<Table.Th class="justify-start">Tag</Table.Th>

								<Table.Th># Courses</Table.Th>

								<!-- Row action menu -->
								<Table.Th />
							</Table.Tr>
						</Table.Thead>

						<Table.Tbody>
							{#if tags.length === 0}
								<Table.Tr>
									<Table.Td class="col-span-full flex-col gap-3 py-5 text-center ">
										<div>No tags</div>

										{#if filterAppliedValue}
											<div class="text-foreground-alt-3">Try adjusting your filters</div>
										{/if}
									</Table.Td>
								</Table.Tr>
							{/if}

							{#each tags as tag, i (tag.id)}
								<Table.Tr class="group">
									<!-- Checkbox -->
									<Table.Td class="group-hover:bg-background-alt-1">
										<Checkbox
											checked={selectedTags[tag.id] !== undefined}
											onCheckedChange={(checked) => {
												if (checked) {
													selectedTags[tag.id] = tag;
												} else {
													delete selectedTags[tag.id];
												}

												toastCount();
											}}
										/>
									</Table.Td>

									<!-- Tag -->
									<Table.Td class="group-hover:bg-background-alt-1 justify-start px-4">
										{tag.tag}
									</Table.Td>

									<!-- Number of courses -->
									<Table.Td class="group-hover:bg-background-alt-1 px-4">
										{tag.courseCount}</Table.Td
									>

									<!-- Row action menu -->
									<Table.Td class="group-hover:bg-background-alt-1">
										<RowActionMenu
											bind:tag={tags[i]}
											onDelete={async () => {
												await onRowDelete(1);
												if (selectedTags[tag.id] !== undefined) {
													delete selectedTags[tag.id];
												}
											}}
										/>
									</Table.Td>
								</Table.Tr>
							{/each}
						</Table.Tbody>
					</Table.Root>

					{#if tags.length !== 0}
						<Pagination
							count={paginationTotal}
							bind:perPage={paginationPerPage}
							bind:page={paginationPage}
							onPageChange={fetchTags}
							onPerPageChange={fetchTags}
						/>
					{/if}
				</div>
			{:catch error}
				<div class="flex w-full flex-col items-center gap-2 pt-10">
					<WarningIcon class="text-foreground-error size-10" />
					<span class="text-lg">Failed to fetch tags: {error.message}</span>
				</div>
			{/await}
		</div>
	</div>
</div>
