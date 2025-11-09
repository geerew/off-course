<!-- TODO have a columns dropdown to hide show columns -->
<!-- TODO store selection state in localstorage -->
<script lang="ts">
	import { GetScans } from '$lib/api/scan-api';
	import { Pagination, SortMenu } from '$lib/components';
	import { RightChevronIcon, WarningIcon } from '$lib/components/icons';
	import RowActionMenu from '$lib/components/pages/admin/scans/row-action-menu.svelte';
	import TableActionMenu from '$lib/components/pages/admin/scans/table-action-menu.svelte';
	import Spinner from '$lib/components/spinner.svelte';
	import { Badge, Button, Checkbox } from '$lib/components/ui';
	import * as Table from '$lib/components/ui/table';
	import type { ScanModel, ScansModel } from '$lib/models/scan-model';
	import { scanMonitor } from '$lib/scans.svelte';
	import type { SortColumns, SortDirection } from '$lib/types/sort';
	import { cn, remCalc } from '$lib/utils';
	import { ElementSize, PersistedState } from 'runed';
	import { tick } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { slide } from 'svelte/transition';
	import theme from 'tailwindcss/defaultTheme';

	let scans: ScansModel = $state([]);

	let filterValue = $state('');
	let filterAppliedValue = $state('');
	let filterOptions = {};

	let expandedScans: Record<string, boolean> = $state({});

	let selectedScans: Record<string, ScanModel> = $state({});
	let selectedScansCount = $derived(Object.keys(selectedScans).length);

	let sortColumns = [
		{ label: 'Course Path', column: 'courses.path', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Status', column: 'scans.status', asc: 'Ascending', desc: 'Descending' }
	] as const satisfies SortColumns;

	type PersistedState = {
		sort: {
			column: (typeof sortColumns)[number]['column'];
			direction: SortDirection;
		};
	};

	const persistedState = new PersistedState<PersistedState>('admin_scans', {
		sort: { column: 'scans.status', direction: 'desc' }
	});

	let selectedSortColumn = $state(persistedState.current.sort.column);
	let selectedSortDirection = $state(persistedState.current.sort.direction);

	let paginationPage = $state(1);
	let paginationPerPage = $state(10);
	let paginationTotal = $state(0);

	let isIndeterminate = $derived(selectedScansCount > 0 && selectedScansCount < paginationTotal);
	let isChecked = $derived(selectedScansCount !== 0 && selectedScansCount === paginationTotal);

	let mainEl = $state() as HTMLElement;
	const mainSize = new ElementSize(() => mainEl);
	let smallTable = $state(false);

	let loadPromise = $state(fetchScans());

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch scans
	async function fetchScans(): Promise<void> {
		try {
			scanMonitor.clearAll();

			const sort = `sort:"${selectedSortColumn} ${selectedSortDirection}"`;
			const q = filterValue ? `${filterValue} ${sort}` : sort;

			const data = await GetScans({
				q,
				page: paginationPage,
				perPage: paginationPerPage
			});
			paginationTotal = data.totalItems;
			scans = data.items;
			expandedScans = {};

			scanMonitor.trackScansArray(scans);
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

		loadPromise = fetchScans();
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function onCheckboxClicked(e: MouseEvent) {
		e.preventDefault();

		const allScansSelectedOnPage = scans.every((s) => {
			return selectedScans[s.id] !== undefined;
		});

		if (allScansSelectedOnPage) {
			scans.forEach((s) => {
				delete selectedScans[s.id];
			});
		} else {
			scans.forEach((s) => {
				selectedScans[s.id] = s;
			});
		}

		toastCount();
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function toggleRowExpansion(userId: string) {
		expandedScans[userId] = !expandedScans[userId];
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function toastCount() {
		if (scans.length === 0) return;

		if (selectedScansCount === 0) {
			toast.success('No scans selected');
		} else {
			toast.success(`${selectedScansCount} row${selectedScansCount > 1 ? 's' : ''} selected`);
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	// Flip between table and card mode based on screen size
	$effect(() => {
		smallTable = remCalc(mainSize.width) <= +theme.columns['4xl'].replace('rem', '') ? true : false;
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Stop the scan monitor when the component is destroyed
	$effect(() => {
		return () => scanMonitor.clearAll();
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update the page when the scans array is empty and the current page is greater than 1.
	// This is to prevent the user from being stuck on an empty page
	$effect(() => {
		if (scans.length === 0 && paginationPage > 1) {
			paginationPage -= 1;
			loadPromise = fetchScans();
		}
	});
</script>

<div class="flex w-full place-content-center" bind:this={mainEl}>
	<div class="flex w-full max-w-7xl flex-col gap-6 pt-1">
		<div class="flex flex-col gap-3 md:flex-row">
			<div class="flex flex-1 flex-row">
				<!-- <FilterBar
					bind:value={filterValue}
					disabled={!filterAppliedValue && courses.length === 0}
					{filterOptions}
					onApply={async () => {
						if (filterValue !== filterAppliedValue) {
							filterAppliedValue = filterValue;
							paginationPage = 1;
							loadPromise = fetchScans();
						}
					}}
				/> -->
			</div>

			<div class="flex flex-row justify-end gap-3">
				<div class="flex h-10 items-center gap-3 rounded-lg">
					<TableActionMenu
						bind:scans={selectedScans}
						onDelete={() => {
							Object.values(selectedScans).forEach((scan) => {
								scanMonitor.untrackScan(scan.courseId);
							});

							const numDeleted = Object.keys(selectedScans).length;
							selectedScans = {};
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

							loadPromise = fetchScans();
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
							: 'grid-cols-[3.5rem_1fr_auto_auto_3.5rem]'}
					>
						<Table.Thead>
							<Table.Tr class="text-xs font-semibold uppercase">
								<!-- Chevron (small screens) -->
								<Table.Th class={smallTable ? 'visible' : 'hidden'}></Table.Th>

								<!-- Checkbox-->
								<Table.Th>
									<Checkbox
										disabled={scans.length === 0}
										indeterminate={isIndeterminate}
										checked={isChecked}
										onclick={onCheckboxClicked}
									/>
								</Table.Th>

								<!-- Course path -->
								<Table.Th class="justify-start">Course Path</Table.Th>

								<!-- Status (large screens) -->
								<Table.Th class={smallTable ? 'hidden' : 'visible'}>Status</Table.Th>

								<!-- Message (large screens) -->
								<Table.Th class={smallTable ? 'hidden' : 'visible'}>Progress</Table.Th>

								<!-- Row action menu -->
								<Table.Th></Table.Th>
							</Table.Tr>
						</Table.Thead>

						<Table.Tbody>
							{#if scans.length === 0}
								<Table.Tr>
									<Table.Td class="col-span-full flex-col gap-3 py-5 text-center ">
										<div>No scans</div>

										{#if filterAppliedValue}
											<div class="text-foreground-alt-3">Try adjusting your filters</div>
										{/if}
									</Table.Td>
								</Table.Tr>
							{/if}

							{#each scans as scan (scan.id)}
								<Table.Tr class="group">
									<!-- Chevron (small screens) -->
									<Table.Td
										class={cn(
											'group-hover:bg-background-alt-1 relative',
											smallTable ? 'visible' : 'hidden'
										)}
									>
										<div
											class={cn(
												'absolute left-1 top-1/2 inline-block h-[70%] w-1 -translate-y-1/2 opacity-60',
												scan.status === 'processing'
													? 'bg-background-primary'
													: 'bg-background-alt-4',
												smallTable ? 'visible' : 'hidden'
											)}
										></div>

										<Button
											variant="ghost"
											class="text-foreground-alt-2 hover:text-foreground h-auto p-1 enabled:hover:bg-transparent"
											title={expandedScans[scan.id] ? 'Collapse details' : 'Expand details'}
											aria-expanded={!!expandedScans[scan.id]}
											aria-controls={`expanded-row-${scan.id}`}
											onclick={() => toggleRowExpansion(scan.id)}
										>
											<RightChevronIcon
												class={cn(
													'size-4 stroke-2 transition-transform duration-200',
													expandedScans[scan.id] ? 'rotate-90' : ''
												)}
											/>
											<span class="sr-only">Details</span>
										</Button>
									</Table.Td>

									<!-- Checkbox -->
									<Table.Td class="group-hover:bg-background-alt-1 relative">
										<div
											class={cn(
												'absolute left-1 top-1/2 inline-block h-[70%] w-1 -translate-y-1/2 opacity-60',
												scan.status === 'processing'
													? 'bg-background-primary'
													: 'bg-background-alt-4',
												smallTable ? 'hidden' : 'visible'
											)}
										></div>

										<Checkbox
											checked={selectedScans[scan.id] !== undefined}
											onCheckedChange={(checked) => {
												if (checked) {
													selectedScans[scan.id] = scan;
												} else {
													delete selectedScans[scan.id];
												}

												toastCount();
											}}
										/>
									</Table.Td>

									<!-- Course path -->
									<Table.Td class="group-hover:bg-background-alt-1 relative justify-start px-4">
										<span>{scan.coursePath}</span>
									</Table.Td>

									<!-- Status (large screens) -->
									<Table.Td
										class={cn(
											'group-hover:bg-background-alt-1 px-4',
											smallTable ? 'hidden' : 'visible'
										)}
									>
										<div class="flex w-full place-content-center">
											<Badge
												class={scan.status === 'processing'
													? 'bg-background-success text-foreground'
													: 'bg-background-alt-4 text-foreground-alt-1'}
											>
												{scan.status}
											</Badge>
										</div>
									</Table.Td>

									<!-- Message (large screens) -->
									<Table.Td
										class={cn(
											'group-hover:bg-background-alt-1 px-4',
											smallTable ? 'hidden' : 'visible'
										)}
									>
										<div class="flex w-full justify-start">
											{#if scan.message}
												<span class="text-foreground-alt-1">{scan.message}</span>
											{:else}
												<span class="text-foreground-alt-3">—</span>
											{/if}
										</div>
									</Table.Td>

									<!-- Row action menu -->
									<Table.Td class="group-hover:bg-background-alt-1">
										<RowActionMenu
											{scan}
											onDelete={async () => {
												scanMonitor.untrackScan(scan.courseId);
												await onRowDelete(1);
												if (selectedScans[scan.id] !== undefined) {
													delete selectedScans[scan.id];
												}
											}}
										/>
									</Table.Td>
								</Table.Tr>

								{#if smallTable && expandedScans[scan.id]}
									<Table.Tr>
										<Table.Td
											inTransition={slide}
											inTransitionParams={{ duration: 200 }}
											outTransition={slide}
											outTransitionParams={{ duration: 150 }}
											class="bg-background-alt-2/30 col-span-full justify-start pl-14 pr-4"
										>
											<div class="flex flex-col gap-2 py-3 text-sm">
												<div class="grid grid-cols-[8rem_1fr]">
													<span class="text-foreground-alt-3 font-medium">STATUS</span>
													<span class="text-foreground-alt-1">
														<Badge
															class={scan.status === 'processing'
																? 'bg-background-success text-foreground'
																: 'bg-background-alt-4 text-foreground-alt-1'}
														>
															{scan.status}
														</Badge>
													</span>
												</div>

												{#if scan.message}
													<div class="grid grid-cols-[8rem_1fr]">
														<span class="text-foreground-alt-3 font-medium">PROGRESS</span>
														<span class="text-foreground-alt-1">
															{scan.message}
														</span>
													</div>
												{/if}
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

					{#if scans.length !== 0}
						<Pagination
							count={paginationTotal}
							bind:perPage={paginationPerPage}
							bind:page={paginationPage}
							onPageChange={() => fetchScans()}
							onPerPageChange={() => fetchScans()}
						/>
					{/if}
				</div>
			{:catch error}
				<div class="flex w-full flex-col items-center gap-2 pt-10">
					<WarningIcon class="text-foreground-error size-10" />
					<span class="text-lg">Failed to fetch scans: {error.message}</span>
				</div>
			{/await}
		</div>
	</div>
</div>
