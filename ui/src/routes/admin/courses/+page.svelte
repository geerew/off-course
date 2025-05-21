<script lang="ts">
	import { GetCourses } from '$lib/api/course-api';
	import { FilterBar, NiceDate, Pagination, SortMenu } from '$lib/components';
	import { AddCoursesDialog } from '$lib/components/dialogs';
	import { RightChevronIcon, TickIcon, WarningIcon, XIcon } from '$lib/components/icons';
	import RowActionMenu from '$lib/components/pages/admin/courses/row-action-menu.svelte';
	import TableActionMenu from '$lib/components/pages/admin/courses/table-action-menu.svelte';
	import Spinner from '$lib/components/spinner.svelte';
	import { Button, Checkbox } from '$lib/components/ui';
	import * as Table from '$lib/components/ui/table';
	import type { CourseModel, CoursesModel } from '$lib/models/course-model';
	import { scanMonitor } from '$lib/scans.svelte';
	import type { SortColumns, SortDirection } from '$lib/types/sort';
	import { cn, remCalc } from '$lib/utils';
	import { ElementSize } from 'runed';
	import { tick } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { slide } from 'svelte/transition';
	import theme from 'tailwindcss/defaultTheme';

	let courses: CoursesModel = $state([]);

	let filterValue = $state('');
	let filterAppliedValue = $state('');
	let filterOptions = {
		available: ['true', 'false'],
		tag: [],
		progress: ['not started', 'started', 'completed']
	};

	let expandedCourses: Record<string, boolean> = $state({});

	let selectedCourses: Record<string, CourseModel> = $state({});
	let selectedCoursesCount = $derived(Object.keys(selectedCourses).length);

	let sortColumns = [
		{ label: 'Title', column: 'courses.title', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Available', column: 'courses.available', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Card', column: 'courses.card_path', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Added', column: 'courses.created_at', asc: 'Oldest', desc: 'Newest' },
		{ label: 'Updated', column: 'courses.updated_at', asc: 'Oldest', desc: 'Newest' }
	] as const satisfies SortColumns;
	let selectedSortColumn = $state<(typeof sortColumns)[number]['column']>('courses.updated_at');
	let selectedSortDirection = $state<SortDirection>('desc');

	let paginationPage = $state(1);
	let paginationPerPage = $state(10);
	let paginationTotal = $state(0);

	let isIndeterminate = $derived(
		selectedCoursesCount > 0 && selectedCoursesCount < paginationTotal
	);
	let isChecked = $derived(selectedCoursesCount !== 0 && selectedCoursesCount === paginationTotal);

	let mainEl = $state() as HTMLElement;
	const mainSize = new ElementSize(() => mainEl);
	let smallTable = $state(false);

	let loadPromise = $state(fetchCourses(true));

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Stop the scan monitor when the component is destroyed (if running)
	$effect(() => {
		return () => scanMonitor.clearAll();
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch courses
	async function fetchCourses(doScan: boolean): Promise<void> {
		try {
			const sort = `sort:"${selectedSortColumn} ${selectedSortDirection}"`;
			const q = filterValue ? `${filterValue} ${sort}` : sort;

			const data = await GetCourses({
				q,
				page: paginationPage,
				perPage: paginationPerPage
			});
			paginationTotal = data.totalItems;
			courses = data.items;
			expandedCourses = {};

			const coursesToTrack = courses.filter(
				(course) => course.scanStatus === 'processing' || course.scanStatus === 'waiting'
			);

			scanMonitor.trackCourses(coursesToTrack);
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function onRowDelete(numDeleted: number) {
		const remainingTotal = paginationTotal - numDeleted;
		const totalPages = Math.max(1, Math.ceil(remainingTotal / paginationPerPage));

		if (paginationPage > totalPages && totalPages > 0) {
			paginationPage = totalPages;
		} else if (remainingTotal === 0) {
			paginationPage = 1;
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

	function toggleRowExpansion(userId: string) {
		expandedCourses[userId] = !expandedCourses[userId];
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

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	// Flip between table and card mode based on screen size
	$effect(() => {
		smallTable = remCalc(mainSize.width) <= +theme.columns['4xl'].replace('rem', '') ? true : false;
	});
</script>

<div class="flex w-full place-content-center" bind:this={mainEl}>
	<div class="flex w-full max-w-7xl flex-col gap-6 pt-1">
		<div class="flex flex-row items-center justify-between">
			<AddCoursesDialog
				successFn={() => {
					loadPromise = fetchCourses(true);
				}}
			/>
		</div>

		<div class="flex flex-col gap-3 md:flex-row">
			<div class="flex flex-1 flex-row">
				<FilterBar
					bind:value={filterValue}
					disabled={!filterAppliedValue && courses.length === 0}
					{filterOptions}
					onApply={async () => {
						if (filterValue !== filterAppliedValue) {
							filterAppliedValue = filterValue;
							paginationPage = 1;
							loadPromise = fetchCourses(true);
						}
					}}
				/>
			</div>

			<div class="flex flex-row justify-end gap-3">
				<div class="flex h-10 items-center gap-3 rounded-lg">
					<TableActionMenu
						bind:courses={selectedCourses}
						onScan={async () => {
							selectedCourses = {};
						}}
						onDelete={() => {
							const numDeleted = Object.keys(selectedCourses).length;
							selectedCourses = {};
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
							loadPromise = fetchCourses(true);
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
					<Table.Root
						class={smallTable
							? 'grid-cols-[2.5rem_2.5rem_1fr_3.5rem]'
							: 'grid-cols-[3.5rem_1fr_auto_auto_auto_auto_3.5rem]'}
					>
						<Table.Thead>
							<Table.Tr class="text-xs font-semibold uppercase">
								<!-- Chevron (small screens) -->
								<Table.Th class={smallTable ? 'visible' : 'hidden'}></Table.Th>

								<!-- Checkbox-->
								<Table.Th>
									<Checkbox
										disabled={courses.length === 0}
										indeterminate={isIndeterminate}
										checked={isChecked}
										onclick={onCheckboxClicked}
									/>
								</Table.Th>

								<!-- Course -->
								<Table.Th class="justify-start">Course</Table.Th>

								<!-- Available (large screens) -->
								<Table.Th class={smallTable ? 'hidden' : 'visible'}>Available</Table.Th>

								<!-- Card (large screens) -->
								<Table.Th class={smallTable ? 'hidden' : 'visible'}>Card</Table.Th>

								<!-- Added (large screens) -->
								<Table.Th class={smallTable ? 'hidden' : 'visible'}>Added</Table.Th>

								<!-- Updated (large screens) -->
								<Table.Th class={smallTable ? 'hidden' : 'visible'}>Updated</Table.Th>

								<!-- Row action menu -->
								<Table.Th></Table.Th>
							</Table.Tr>
						</Table.Thead>

						<Table.Tbody>
							{#if courses.length === 0}
								<Table.Tr>
									<Table.Td class="col-span-full flex-col gap-3 py-5 text-center ">
										<div>No courses</div>

										{#if filterAppliedValue}
											<div class="text-foreground-alt-3">Try adjusting your filters</div>
										{/if}
									</Table.Td>
								</Table.Tr>
							{/if}

							{#each courses as course (course.id)}
								<Table.Tr class="group">
									<!-- Chevron (small screens) -->
									<Table.Td
										class={cn(
											'group-hover:bg-background-alt-1 relative',
											smallTable ? 'visible' : 'hidden'
										)}
									>
										<Button
											class="text-foreground-alt-2 hover:text-foreground h-auto w-auto rounded bg-transparent p-1 enabled:hover:bg-transparent"
											title={expandedCourses[course.id] ? 'Collapse details' : 'Expand details'}
											aria-expanded={!!expandedCourses[course.id]}
											aria-controls={`expanded-row-${course.id}`}
											onclick={() => toggleRowExpansion(course.id)}
										>
											<RightChevronIcon
												class={cn(
													'size-4 stroke-2 transition-transform duration-200',
													expandedCourses[course.id] ? 'rotate-90' : ''
												)}
											/>
											<span class="sr-only">Details</span>
										</Button>
									</Table.Td>

									<!-- Checkbox -->
									<Table.Td class="group-hover:bg-background-alt-1 relative">
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

									<!-- Course -->
									<Table.Td class="group-hover:bg-background-alt-1 relative justify-start px-4">
										{#if scanMonitor.scans[course.id] !== undefined}
											<div
												class={cn(
													'absolute top-0 right-0 inline-block h-full w-1 opacity-60',
													scanMonitor.scans[course.id] === 'processing'
														? 'bg-background-primary'
														: 'bg-background-alt-4'
												)}
											></div>
										{/if}

										<span>{course.title}</span>
									</Table.Td>

									<!-- Available (large screens) -->
									<Table.Td
										class={cn(
											'group-hover:bg-background-alt-1 px-4',
											smallTable ? 'hidden' : 'visible'
										)}
									>
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

									<!-- Card (large screens) -->
									<Table.Td
										class={cn(
											'group-hover:bg-background-alt-1 px-4',
											smallTable ? 'hidden' : 'visible'
										)}
									>
										<div class="flex w-full place-content-center">
											{#if course.hasCard}
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

									<!-- Added (large screens) -->
									<Table.Td
										class={cn(
											'group-hover:bg-background-alt-1 px-4 whitespace-nowrap',
											smallTable ? 'hidden' : 'visible'
										)}
									>
										<NiceDate date={course.createdAt} />
									</Table.Td>

									<!-- Updated (large screens) -->
									<Table.Td
										class={cn(
											'group-hover:bg-background-alt-1 px-4 whitespace-nowrap',
											smallTable ? 'hidden' : 'visible'
										)}
									>
										<NiceDate date={course.updatedAt} />
									</Table.Td>

									<!-- Row action menu -->
									<Table.Td class="group-hover:bg-background-alt-1">
										<RowActionMenu
											{course}
											onDelete={async () => {
												await onRowDelete(1);
												if (selectedCourses[course.id] !== undefined) {
													delete selectedCourses[course.id];
												}
											}}
										/>
									</Table.Td>
								</Table.Tr>

								{#if smallTable && expandedCourses[course.id]}
									<Table.Tr>
										<Table.Td
											inTransition={slide}
											inTransitionParams={{ duration: 200 }}
											outTransition={slide}
											outTransitionParams={{ duration: 150 }}
											class="bg-background-alt-2/30 col-span-full justify-start pr-4 pl-14"
										>
											<div class="flex flex-col gap-2 py-3 text-sm">
												<div class="grid grid-cols-[8rem_1fr]">
													<span class="text-foreground-alt-3 font-medium">STATUS</span>
													<span
														class={course.available
															? 'text-background-success'
															: 'text-foreground-error'}
														>{course.available ? 'available' : 'unavailable'}</span
													>
												</div>

												<div class="grid grid-cols-[8rem_1fr]">
													<span class="text-foreground-alt-3 font-medium">HAS CARD</span>
													<span
														class={course.hasCard
															? 'text-background-success'
															: 'text-foreground-error'}>{course.hasCard ? 'yes' : 'no'}</span
													>
												</div>

												<div class="grid grid-cols-[8rem_1fr]">
													<span class="text-foreground-alt-3 font-medium">ADDED</span>
													<span class="text-foreground-alt-1">
														<NiceDate date={course.createdAt} />
													</span>
												</div>

												<div class="grid grid-cols-[8rem_1fr]">
													<span class="text-foreground-alt-3 font-medium">UPDATED</span>
													<span class="text-foreground-alt-1">
														<NiceDate date={course.updatedAt} />
													</span>
												</div>
											</div>
										</Table.Td>
									</Table.Tr>
								{/if}
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
