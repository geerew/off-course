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
	import { tick } from 'svelte';
	import { toast } from 'svelte-sonner';

	let tags: TagsModel = $state([]);

	let filterValue = $state('');

	let selectedTags: Record<string, TagModel> = $state({});
	let selectedTagsCount = $derived(Object.keys(selectedTags).length);

	let paginationPage = $state(1);
	let paginationPerPage = $state(10);
	let paginationTotal = $state(0);

	let sortColumns = [
		{ label: 'Tag', column: 'tags.tag', asc: 'Ascending', desc: 'Descending' }
	] as const satisfies SortColumns;
	let selectedSortColumn = $state<(typeof sortColumns)[number]['column']>('tags.tag');
	let selectedSortDirection = $state<SortDirection>('asc');

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

	async function onRowDelete() {
		// If the current page is greater than the new total, set it to the last
		// page
		if (paginationPage > Math.ceil(paginationTotal / paginationPerPage)) {
			paginationPage = Math.ceil(paginationTotal / paginationPerPage);
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
			toast.success(`${selectedTagsCount} tag${selectedTagsCount > 1 ? 's' : ''} selected`);
		}
	}
</script>

<div class="flex w-full place-content-center">
	<div class="flex w-full max-w-4xl min-w-2xl flex-col gap-6 pt-1">
		<div class="flex flex-row items-center justify-between">
			<AddTagsDialog
				successFn={() => {
					loadPromise = fetchTags();
				}}
			/>
		</div>
		<div class="flex flex-row gap-3">
			<div class="flex flex-1 flex-row">
				<FilterBar
					bind:value={filterValue}
					onUpdate={async () => {
						await tick();
						loadPromise = fetchTags();
					}}
				/>
			</div>

			<div class="flex h-10 items-center gap-3 rounded-lg">
				<TableActionMenu
					bind:tags={selectedTags}
					onDelete={() => {
						selectedTags = {};
						onRowDelete();
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
						loadPromise = fetchTags();
					}}
				/>
			</div>
		</div>

		<div class="flex w-full place-content-center">
			{#await loadPromise}
				<div class="flex justify-center pt-10">
					<Spinner class="bg-foreground-alt-2 size-4" />
				</div>
			{:then _}
				<div class="flex w-full flex-col gap-8">
					<Table.Root>
						<Table.Thead>
							<Table.Tr>
								<Table.Th class="w-[1%]">
									<Checkbox
										indeterminate={isIndeterminate}
										checked={isChecked}
										onclick={onCheckboxClicked}
									/>
								</Table.Th>
								<Table.Th>Tag</Table.Th>
								<Table.Th class="min-w-[1%]">Courses</Table.Th>
								<Table.Th class="min-w-[1%]" />
							</Table.Tr>
						</Table.Thead>

						<Table.Tbody>
							{#if tags.length === 0}
								<Table.Tr>
									<Table.Td class="text-center" colspan={9999}>No tags found</Table.Td>
								</Table.Tr>
							{/if}
							{#each tags as tag, i (tag.id)}
								<Table.Tr class="hover:bg-background-alt-1 items-center duration-200">
									<Table.Td>
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

									<Table.Td>
										{tag.tag}
									</Table.Td>

									<Table.Td class="text-center">{tag.courseCount}</Table.Td>

									<Table.Td class="flex items-center justify-center">
										<RowActionMenu bind:tag={tags[i]} onDelete={onRowDelete} />
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
