<script module>
</script>

<script lang="ts">
	import { GetUsers } from '$lib/api/user-api';
	import { auth } from '$lib/auth.svelte';
	import { Pagination } from '$lib/components';
	import { AddUserDialog } from '$lib/components/dialogs';
	import { WarningIcon } from '$lib/components/icons';
	import RowActionMenu from '$lib/components/pages/admin/users/row-action-menu.svelte';
	import SortMenu from '$lib/components/pages/admin/users/sort-menu.svelte';
	import TableActionMenu from '$lib/components/pages/admin/users/table-action-menu.svelte';
	import Spinner from '$lib/components/spinner.svelte';
	import { Checkbox, Filter } from '$lib/components/ui';
	import * as Table from '$lib/components/ui/table';
	import type { UserModel, UsersModel } from '$lib/models/user-model';
	import type { SortColumns, SortDirection } from '$lib/types/sort';
	import { capitalizeFirstLetter } from '$lib/utils';
	import { tick } from 'svelte';
	import { toast } from 'svelte-sonner';

	let users: UsersModel = $state([]);

	let filterValue = $state('');

	let selectedUsers: Record<string, UserModel> = $state({});
	let selectedUsersCount = $derived(Object.keys(selectedUsers).length);

	let sortColumns = [
		{ label: 'Username', column: 'users.username', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Name', column: 'users.display_name', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Role', column: 'users.role', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Created At', column: 'users.created_at', asc: 'Newest', desc: 'Oldest' }
	] as const satisfies SortColumns;
	let selectedSortColumn = $state<(typeof sortColumns)[number]['column']>('users.username');
	let selectedSortDirection = $state<SortDirection>('asc');

	let paginationPage = $state(1);
	let paginationPerPage = $state(10);
	let paginationTotal = $state(0);
	let paginationTotalMinusSelf = $derived(paginationTotal - 1);

	let isIndeterminate = $derived(
		selectedUsersCount > 0 && selectedUsersCount < paginationTotalMinusSelf
	);
	let isChecked = $derived(
		selectedUsersCount !== 0 && selectedUsersCount === paginationTotalMinusSelf
	);

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
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function onRowUpdate() {
		loadPromise = fetchUsers();
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function onRowDelete() {
		// If the current page is greater than the new total, set it to the last
		// page
		if (paginationPage > Math.ceil(paginationTotalMinusSelf / paginationPerPage)) {
			paginationPage = Math.ceil(paginationTotalMinusSelf / paginationPerPage);
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

	function toastCount() {
		if (selectedUsersCount === 0) {
			toast.success('No users selected');
		} else {
			toast.success(`${selectedUsersCount} user${selectedUsersCount > 1 ? 's' : ''} selected`);
		}
	}
</script>

<div class="flex w-full place-content-center">
	<div class="flex w-full max-w-4xl min-w-2xl flex-col gap-6 pt-1">
		<div class="flex flex-row items-center justify-between">
			<AddUserDialog
				successFn={() => {
					loadPromise = fetchUsers();
				}}
			/>
		</div>

		<div class="flex flex-row gap-3">
			<div class="flex flex-1 flex-row">
				<Filter
					bind:value={filterValue}
					onUpdate={async () => {
						await tick();
						loadPromise = fetchUsers();
					}}
				/>
			</div>

			<div class="flex h-10 items-center gap-3 rounded-lg">
				<TableActionMenu
					bind:users={selectedUsers}
					onUpdate={() => {
						selectedUsers = {};
						onRowUpdate();
					}}
					onDelete={() => {
						selectedUsers = {};
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
						loadPromise = fetchUsers();
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
								<Table.Th>Username</Table.Th>
								<Table.Th>Name</Table.Th>
								<Table.Th class="min-w-[1%]">Role</Table.Th>
								<Table.Th class="min-w-[1%]" />
							</Table.Tr>
						</Table.Thead>

						<Table.Tbody>
							{#if users.length === 0}
								<Table.Tr>
									<Table.Td class="text-center" colspan={9999}>No users found</Table.Td>
								</Table.Tr>
							{/if}

							{#each users as user}
								<Table.Tr class="hover:bg-background-alt-1 items-center duration-200">
									<Table.Td>
										{#if user.id != auth.user?.id}
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

									<Table.Td>
										{#if user.id === auth.user?.id}
											<div class="flex items-center gap-2">
												<span>{user.username}</span>
												<div class="bg-background-primary mt-px size-2 rounded-full"></div>
											</div>
										{:else}
											{user.username}
										{/if}
									</Table.Td>

									<Table.Td>{user.displayName}</Table.Td>

									<Table.Td>{capitalizeFirstLetter(user.role)}</Table.Td>

									<Table.Td class="flex items-center justify-center">
										{#if user.id !== auth.user?.id}
											<RowActionMenu
												{user}
												onUpdate={onRowUpdate}
												onDelete={async () => {
													await onRowDelete();
													if (selectedUsers[user.id] !== undefined) {
														delete selectedUsers[user.id];
													}
												}}
											/>
										{/if}
									</Table.Td>
								</Table.Tr>
							{/each}
						</Table.Tbody>
					</Table.Root>

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
