<!-- TODO Show number of ongoing/completed courses (backend work too) -->
<!-- TODO have a columns dropdown to hide show columns -->
<!-- TODO store selection state in localstorage -->
<script lang="ts">
	import { GetUsers } from '$lib/api/user-api';
	import { auth } from '$lib/auth.svelte';
	import { FilterBar, Pagination, SortMenu, Spinner } from '$lib/components';
	import { AddUserDialog } from '$lib/components/dialogs';
	import { RightChevronIcon, WarningIcon } from '$lib/components/icons';
	import RowActionMenu from '$lib/components/pages/admin/users/row-action-menu.svelte';
	import TableActionMenu from '$lib/components/pages/admin/users/table-action-menu.svelte';
	import { Button, Checkbox } from '$lib/components/ui';
	import * as Table from '$lib/components/ui/table';
	import type { UserModel, UsersModel } from '$lib/models/user-model';
	import type { SortColumns, SortDirection } from '$lib/types/sort';
	import { capitalizeFirstLetter, cn, remCalc } from '$lib/utils';
	import { ElementSize } from 'runed';
	import { tick } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { slide } from 'svelte/transition';
	import theme from 'tailwindcss/defaultTheme';

	let users: UsersModel = $state([]);

	let filterValue = $state('');
	let filterAppliedValue = $state('');

	let expandedUsers: Record<string, boolean> = $state({});

	let selectedUsers: Record<string, UserModel> = $state({});
	let selectedUsersCount = $derived(Object.keys(selectedUsers).length);

	let sortColumns = [
		{ label: 'Username', column: 'users.username', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Name', column: 'users.display_name', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Role', column: 'users.role', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Created At', column: 'users.created_at', asc: 'Oldest', desc: 'Newest' }
	] as const satisfies SortColumns;
	let selectedSortColumn = $state<(typeof sortColumns)[number]['column']>('users.username');
	let selectedSortDirection = $state<SortDirection>('asc');

	let paginationPage = $state(1);
	let paginationPerPage = $state(10);
	let paginationTotal = $state(0);

	let usersOnPageExcludingSelf = $derived(users.filter((u) => u.id !== auth.user?.id));
	let countUsersOnPageExcludingSelf = $derived(usersOnPageExcludingSelf.length);
	let selectedUsersOnThisPageCount = $derived(
		usersOnPageExcludingSelf.filter((u) => selectedUsers[u.id]).length
	);

	let isIndeterminate = $derived(
		selectedUsersOnThisPageCount > 0 && selectedUsersOnThisPageCount < countUsersOnPageExcludingSelf
	);
	let isChecked = $derived(
		countUsersOnPageExcludingSelf > 0 &&
			selectedUsersOnThisPageCount === countUsersOnPageExcludingSelf
	);

	let mainEl = $state() as HTMLElement;
	const mainSize = new ElementSize(() => mainEl);
	let smallTable = $state(false);

	let loadPromise = $state(fetchUsers());

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function fetchUsers(): Promise<void> {
		try {
			const sort = `sort:"${selectedSortColumn} ${selectedSortDirection}"`;
			const q = filterValue ? `${filterValue} ${sort}` : sort;

			const data = await GetUsers({
				q,
				page: paginationPage,
				perPage: paginationPerPage
			});
			paginationTotal = data.totalItems;
			users = data.items;
			expandedUsers = {};
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function onRowUpdate() {
		loadPromise = fetchUsers();
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

		loadPromise = fetchUsers();
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function onCheckboxClicked(e: MouseEvent) {
		e.preventDefault();

		const allUsersSelectedOnPage = users.every((u) => {
			if (u.id === auth.user?.id) {
				return true;
			}
			return selectedUsers[u.id] !== undefined;
		});

		if (allUsersSelectedOnPage) {
			users.forEach((u) => {
				if (u.id !== auth.user?.id) {
					delete selectedUsers[u.id];
				}
			});
		} else {
			users.forEach((u) => {
				if (u.id !== auth.user?.id) {
					selectedUsers[u.id] = u;
				}
			});
		}

		toastCount();
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function toggleRowExpansion(userId: string) {
		expandedUsers[userId] = !expandedUsers[userId];
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function toastCount() {
		if (selectedUsersCount === 0) {
			toast.success('No users selected');
		} else {
			toast.success(`${selectedUsersCount} user${selectedUsersCount > 1 ? 's' : ''} selected`);
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Flip between table and card mode based on screen size
	$effect(() => {
		smallTable = remCalc(mainSize.width) <= +theme.columns['2xl'].replace('rem', '') ? true : false;
	});
</script>

<div class="flex w-full place-content-center" bind:this={mainEl}>
	<div class="flex w-full max-w-4xl flex-col gap-6 pt-1">
		<div class="flex flex-row items-center justify-between">
			<AddUserDialog
				successFn={() => {
					loadPromise = fetchUsers();
				}}
			/>
		</div>

		<div class="flex flex-col gap-3 md:flex-row">
			<div class="flex flex-1 flex-row">
				<FilterBar
					bind:value={filterValue}
					disabled={!filterAppliedValue && users.length === 0}
					onApply={async () => {
						if (filterValue !== filterAppliedValue) {
							filterAppliedValue = filterValue;
							paginationPage = 1;
							loadPromise = fetchUsers();
						}
					}}
				/>
			</div>

			<div class="flex flex-row justify-end gap-3">
				<div class="flex h-10 items-center gap-3 rounded-lg">
					<TableActionMenu
						bind:users={selectedUsers}
						onUpdate={() => {
							selectedUsers = {};
							onRowUpdate();
						}}
						onDelete={() => {
							const numDeleted = Object.keys(selectedUsers).length;
							selectedUsers = {};
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
							loadPromise = fetchUsers();
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
						<!-- Header -->
						<Table.Thead>
							<Table.Tr class="text-xs font-semibold uppercase">
								<!-- Chevron (small screens) -->
								<Table.Th class={smallTable ? 'visible' : 'hidden'}></Table.Th>

								<!-- Checkbox -->
								<Table.Th>
									<Checkbox
										indeterminate={isIndeterminate}
										checked={isChecked}
										onclick={onCheckboxClicked}
									/>
								</Table.Th>

								<!-- Username -->
								<Table.Th class="justify-start">Username</Table.Th>

								<!-- Display name (large screens) -->
								<Table.Th class={smallTable ? 'hidden' : 'visible'}>Name</Table.Th>

								<!-- Role (large screens) -->
								<Table.Th class={smallTable ? 'hidden' : 'visible'}>Role</Table.Th>

								<!-- Row action menu -->
								<Table.Th></Table.Th>
							</Table.Tr>
						</Table.Thead>

						<Table.Tbody>
							{#each users as user (user.id)}
								<Table.Tr class="group">
									<Table.Td
										class={cn('group-hover:bg-background-alt-1', smallTable ? 'visible' : 'hidden')}
									>
										<Button
											class="text-foreground-alt-2 hover:text-foreground h-auto w-auto rounded bg-transparent p-1 enabled:hover:bg-transparent"
											title={expandedUsers[user.id] ? 'Collapse details' : 'Expand details'}
											aria-expanded={!!expandedUsers[user.id]}
											aria-controls={`expanded-row-${user.id}`}
											onclick={() => toggleRowExpansion(user.id)}
										>
											<RightChevronIcon
												class={cn(
													'size-4 stroke-2 transition-transform duration-200',
													expandedUsers[user.id] ? 'rotate-90' : ''
												)}
											/>
											<span class="sr-only">Details</span>
										</Button>
									</Table.Td>

									<!-- Checkbox -->
									<Table.Td class="group-hover:bg-background-alt-1">
										{#if user.id !== auth.user?.id}
											<Checkbox
												checked={selectedUsers[user.id] !== undefined}
												onCheckedChange={(checked) => {
													if (checked) {
														selectedUsers[user.id] = user;
													} else {
														delete selectedUsers[user.id];
													}

													toastCount();
												}}
											/>
										{/if}
									</Table.Td>

									<!-- Username -->
									<Table.Td class="group-hover:bg-background-alt-1 justify-start px-4">
										{#if user.id === auth.user?.id}
											<div class="flex items-center gap-2">
												<span>{user.username}</span>
												<div class="bg-background-primary mt-0.5 size-2 rounded-full"></div>
											</div>
										{:else}
											{user.username}
										{/if}
									</Table.Td>

									<!-- Display name (auto) -->
									<Table.Td
										class={cn(
											'group-hover:bg-background-alt-1 px-4',
											smallTable ? 'hidden' : 'visible'
										)}
									>
										{user.displayName}
									</Table.Td>

									<!-- Role (large screens) -->
									<Table.Td
										class={cn(
											'group-hover:bg-background-alt-1 px-4',
											smallTable ? 'hidden' : 'visible'
										)}
									>
										{capitalizeFirstLetter(user.role)}
									</Table.Td>

									<!-- Row action menu -->
									<Table.Td class="group-hover:bg-background-alt-1">
										{#if user.id !== auth.user?.id}
											<RowActionMenu
												{user}
												onUpdate={onRowUpdate}
												onDelete={async () => {
													await onRowDelete(1);
													if (selectedUsers[user.id] !== undefined) {
														delete selectedUsers[user.id];
													}
												}}
											/>
										{/if}
									</Table.Td>
								</Table.Tr>

								{#if smallTable && expandedUsers[user.id]}
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
													<span class="text-foreground-alt-3 font-medium">DISPLAY NAME</span>
													<span class="text-foreground-alt-1">{user.displayName}</span>
												</div>

												<div class="grid grid-cols-[8rem_1fr]">
													<span class="text-foreground-alt-3 font-medium">ROLE</span>
													<span class="text-foreground-alt-1">{user.role}</span>
												</div>
											</div>
										</Table.Td>
									</Table.Tr>
								{/if}
							{/each}
						</Table.Tbody>
					</Table.Root>

					<div class="flex flex-row items-center gap-3 text-sm">
						<span class="text-foreground-alt-2">Current User</span>
						<div class="bg-background-primary mt-px size-4 rounded-md"></div>
					</div>

					<Pagination
						count={paginationTotal}
						bind:perPage={paginationPerPage}
						bind:page={paginationPage}
						onPageChange={fetchUsers}
						onPerPageChange={fetchUsers}
					/>
				</div>
			{:catch error}
				<div class="flex w-full flex-col items-center gap-2 pt-10">
					<WarningIcon class="text-foreground-error size-10" />
					<span class="text-lg">Failed to fetch users: {error.message}</span>
				</div>
			{/await}
		</div>
	</div>
</div>
