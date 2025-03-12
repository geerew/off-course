<script lang="ts">
	import { GetCourses } from '$lib/api/course-api';
	import { FilterBar, NiceDate, Pagination, SortMenu } from '$lib/components';
	import { AddCoursesDialog } from '$lib/components/dialogs';
	import { TickIcon, WarningIcon, XIcon } from '$lib/components/icons';
	import RowActionMenu from '$lib/components/pages/admin/courses/row-action-menu.svelte';
	import TableActionMenu from '$lib/components/pages/admin/courses/table-action-menu.svelte';
	import Spinner from '$lib/components/spinner.svelte';
	import { Checkbox } from '$lib/components/ui';
	import * as Table from '$lib/components/ui/table';
	import type { CourseModel, CoursesModel } from '$lib/models/course-model';
	import { scanMonitor } from '$lib/scans.svelte';
	import type { SortColumns, SortDirection } from '$lib/types/sort';
	import { tick } from 'svelte';
	import { toast } from 'svelte-sonner';

	let courses: CoursesModel = $state([]);

	let filterValue = $state('');

	let selectedCourses: Record<string, CourseModel> = $state({});
	let selectedCoursesCount = $derived(Object.keys(selectedCourses).length);

	let sortColumns = [
		{ label: 'Title', column: 'courses.title', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Available', column: 'courses.available', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Created', column: 'courses.created_at', asc: 'Newest', desc: 'Oldest' },
		{ label: 'Updated', column: 'courses.updated_at', asc: 'Newest', desc: 'Oldest' }
	] as const satisfies SortColumns;
	let selectedSortColumn = $state<(typeof sortColumns)[number]['column']>('courses.updated_at');
	let selectedSortDirection = $state<SortDirection>('asc');

	let paginationPage = $state(1);
	let paginationPerPage = $state(10);
	let paginationTotal = $state(0);

	let isIndeterminate = $derived(
		selectedCoursesCount > 0 && selectedCoursesCount < paginationTotal
	);
	let isChecked = $derived(selectedCoursesCount !== 0 && selectedCoursesCount === paginationTotal);

	let loadPromise = $state(fetchCourses(true));

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Stop the scan monitor when the component is destroyed
	$effect(() => {
		return () => scanMonitor.stop();
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch courses
	async function fetchCourses(doScan: boolean): Promise<void> {
		try {
			if (doScan) await scanMonitor.fetch();

			const sort = `sort:"${selectedSortColumn} ${selectedSortDirection}"`;
			const q = filterValue ? `${filterValue} ${sort}` : sort;

			const data = await GetCourses({
				q,
				page: paginationPage,
				perPage: paginationPerPage
			});
			paginationTotal = data.totalItems;
			courses = data.items;
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

		loadPromise = fetchCourses(false);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function onCheckboxClicked(e: MouseEvent) {
		e.preventDefault();

		const allCoursesSelectedOnPage = courses.every((c) => {
			return selectedCourses[c.id] !== undefined;
		});

		if (allCoursesSelectedOnPage) {
			courses.forEach((c) => {
				delete selectedCourses[c.id];
			});
		} else {
			courses.forEach((c) => {
				selectedCourses[c.id] = c;
			});
		}

		toastCount();
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function toastCount() {
		if (courses.length === 0) return;

		if (selectedCoursesCount === 0) {
			toast.success('No courses selected');
		} else {
			toast.success(`${selectedCoursesCount} user${selectedCoursesCount > 1 ? 's' : ''} selected`);
		}
	}
</script>

<div class="flex w-full place-content-center">
	<div class="flex w-full max-w-7xl min-w-4xl flex-col gap-6 pt-1">
		<div class="flex flex-row items-center justify-between">
			<AddCoursesDialog
				successFn={() => {
					loadPromise = fetchCourses(true);
				}}
			/>
		</div>

		<div class="flex flex-row gap-3">
			<div class="flex flex-1 flex-row">
				<FilterBar
					bind:value={filterValue}
					onUpdate={async () => {
						await tick();
						loadPromise = fetchCourses(true);
					}}
				/>
			</div>

			<div class="flex h-10 items-center gap-3 rounded-lg">
				<TableActionMenu
					bind:courses={selectedCourses}
					onScan={async () => {
						selectedCourses = {};
						await scanMonitor.fetch();
					}}
					onDelete={() => {
						selectedCourses = {};
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
						loadPromise = fetchCourses(true);
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
								<Table.Th class="min-w-[1%]">
									<Checkbox
										disabled={courses.length === 0}
										indeterminate={isIndeterminate}
										checked={isChecked}
										onclick={onCheckboxClicked}
									/>
								</Table.Th>
								<Table.Th class="max-w-[5rem]">Course</Table.Th>
								<Table.Th class="min-w-[1%]">Available</Table.Th>
								<Table.Th class="min-w-[1%] text-center">Created</Table.Th>
								<Table.Th class="min-w-[1%] text-center">Updated</Table.Th>
								<Table.Th class="min-w-[1%]" />
							</Table.Tr>
						</Table.Thead>

						<Table.Tbody>
							{#if courses.length === 0}
								<Table.Tr>
									<Table.Td class="text-center" colspan={9999}>No courses found</Table.Td>
								</Table.Tr>
							{/if}

							{#each courses as course (course.id)}
								<Table.Tr class="hover:bg-background-alt-1 items-center duration-200">
									<Table.Td>
										<Checkbox
											checked={selectedCourses[course.id] !== undefined}
											onCheckedChange={(checked) => {
												if (checked) {
													selectedCourses[course.id] = course;
												} else {
													delete selectedCourses[course.id];
												}

												toastCount();
											}}
										/>
									</Table.Td>

									<Table.Td>
										<div class="flex items-center gap-2">
											<span>{course.title}</span>
											{#if scanMonitor.scans[course.id] !== undefined}
												{#if scanMonitor.scans[course.id] === 'processing'}
													<div
														class="bg-background-primary mt-0.5 size-2 shrink-0 rounded-full"
													></div>
												{:else}
													<div
														class="bg-background-alt-6 mt-0.5 size-2 shrink-0 rounded-full"
													></div>
												{/if}
											{/if}
										</div>
									</Table.Td>

									<Table.Td class="min-w-[1%]">
										<div class="flex w-full place-content-center">
											{#if course.available}
												<div class="bg-background-success size-5 place-self-center rounded-md p-1">
													<TickIcon class="text-foreground size-3 stroke-2" />
												</div>
											{:else}
												<div class="bg-background-error size-5 place-self-center rounded-md p-1">
													<XIcon class="text-foreground size-3 stroke-2" />
												</div>
											{/if}
										</div>
									</Table.Td>

									<Table.Td class="min-w-[1%] whitespace-nowrap">
										<NiceDate date={course.createdAt} />
									</Table.Td>
									<Table.Td class="w-[1%] whitespace-nowrap">
										<NiceDate date={course.updatedAt} />
									</Table.Td>

									<Table.Td class="flex items-center justify-center">
										<RowActionMenu
											{course}
											onScan={async () => {
												await scanMonitor.fetch();
											}}
											onDelete={async () => {
												await onRowDelete();
												if (selectedCourses[course.id] !== undefined) {
													delete selectedCourses[course.id];
												}
											}}
										/>
									</Table.Td>
								</Table.Tr>
							{/each}
						</Table.Tbody>
					</Table.Root>

					<div class="flex flex-row gap-3 text-sm">
						<span>Scan Status:</span>
						<div class="flex flex-row gap-3">
							<div class="flex flex-row items-center gap-2">
								<div class="bg-background-primary mt-px size-4 rounded-md"></div>
								<span>Processing</span>
							</div>
							<div class="flex flex-row items-center gap-2">
								<div class="bg-background-alt-4 mt-px size-4 rounded-md"></div>
								<span>Waiting</span>
							</div>
						</div>
					</div>

					{#if courses.length !== 0}
						<Pagination
							count={paginationTotal}
							bind:perPage={paginationPerPage}
							bind:page={paginationPage}
							onPageChange={() => fetchCourses(false)}
							onPerPageChange={() => fetchCourses(false)}
						/>
					{/if}
				</div>
			{:catch error}
				<div class="flex w-full flex-col items-center gap-2 pt-10">
					<WarningIcon class="text-foreground-error size-10" />
					<span class="text-lg">Failed to fetch courses: {error.message}</span>
				</div>
			{/await}
		</div>
	</div>
</div>
