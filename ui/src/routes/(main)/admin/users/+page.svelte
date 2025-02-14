<script lang="ts">
	import { GetUsers } from '$lib/api/users';
	import { auth } from '$lib/auth.svelte';
	import { Pagination } from '$lib/components';
	import { PlusIcon, WarningIcon } from '$lib/components/icons';
	import ActionMenu from '$lib/components/pages/admin/users/action-menu.svelte';
	import Spinner from '$lib/components/spinner.svelte';
	import * as Table from '$lib/components/table';
	import { Button, Checkbox } from '$lib/components/ui';
	import type { UsersModel } from '$lib/models/user';
	import { capitalizeFirstLetter } from '$lib/utils';

	let users: UsersModel = $state([]);

	let paginationPage = $state(1);
	let paginationPerPage = $state(10);
	let paginationTotal = $state(0);

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
			users = data.items as UsersModel;
		} catch (error) {
			throw error;
		}
	}
</script>

<div class="flex w-full place-content-center">
	<div class="flex w-full max-w-4xl min-w-2xl flex-col gap-6 pt-1">
		<div>
			<Button
				href="/admin/users/add"
				class="bg-background-alt-4 hover:bg-background-alt-5 text-foreground-alt-1 inline-flex h-10 w-auto flex-row items-center gap-2 rounded-md px-5 hover:cursor-pointer"
				aria-label="Toggle password visibility"
			>
				<PlusIcon class="size-5 stroke-[1.5]" />
				Add User
			</Button>
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
								<Table.Th class="w-[1%]"><Checkbox /></Table.Th>
								<Table.Th>Username</Table.Th>
								<Table.Th>Name</Table.Th>
								<Table.Th>Role</Table.Th>
								<Table.Th class="w-[1%]" />
							</Table.Tr>
						</Table.Thead>
						<Table.Tbody>
							{#each users as user}
								<Table.Tr class="hover:bg-background-alt-1 items-center duration-200">
									<Table.Td>
										{#if user.id != auth.user?.id}
											<Checkbox />
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
											<ActionMenu
												{user}
												onUpdate={() => {
													loadPromise = fetchUsers();
												}}
												onDelete={() => {
													// If the current page is greater than the new total, set it to the last
													// page
													const newTotal = paginationTotal - 1;
													if (paginationPage > Math.ceil(newTotal / paginationPerPage)) {
														paginationPage = Math.ceil(newTotal / paginationPerPage);
													}
													loadPromise = fetchUsers();
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
