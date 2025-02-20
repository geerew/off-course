<script lang="ts">
	import { GetUsers } from '$lib/api/user-api';
	import { auth } from '$lib/auth.svelte';
	import { Pagination } from '$lib/components';
	import { WarningIcon } from '$lib/components/icons';
	import AddUserDialog from '$lib/components/pages/admin/users/add-user-dialog.svelte';
	import RowActionMenu from '$lib/components/pages/admin/users/row-action-menu.svelte';
	import TableActionMenu from '$lib/components/pages/admin/users/table-action-menu.svelte';
	import Spinner from '$lib/components/spinner.svelte';
	import * as Table from '$lib/components/table';
	import { Checkbox } from '$lib/components/ui';
	import type { UserModel, UsersModel } from '$lib/models/user-model';
	import { capitalizeFirstLetter } from '$lib/utils';
	import { toast } from 'svelte-sonner';

	let users: UsersModel = $state([]);

	// Object of selected users and the count of selected users
	let selectedUsers: Record<string, UserModel> = $state({});
	let selectedUserCount = $derived(Object.keys(selectedUsers).length);

	let paginationPage = $state(1);
	let paginationPerPage = $state(10);
	let paginationTotal = $state(0);
	let paginationTotalMinusSelf = $derived(paginationTotal - 1);

	// Whether the main checkbox is indeterminate/checked
	let isIndeterminate = $derived(
		selectedUserCount > 0 && selectedUserCount < paginationTotalMinusSelf
	);
	let isChecked = $derived(
		selectedUserCount !== 0 && selectedUserCount === paginationTotalMinusSelf
	);

	let loadPromise = $state(fetchUsers());

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function fetchUsers(): Promise<void> {
		try {
			const data = await GetUsers({
				orderBy: 'username',
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
		if (selectedUserCount === 0) {
			toast.success('No users selected');
		} else {
			toast.success(`${selectedUserCount} user${selectedUserCount > 1 ? 's' : ''} selected`);
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

			<div class="flex h-10 items-center gap-3 rounded-lg">
				<TableActionMenu
					users={selectedUsers}
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
