<script lang="ts">
	import { GetUsers } from '$lib/api/users';
	import { DotsIcon, PlusIcon, WarningIcon } from '$lib/components/icons';
	import Spinner from '$lib/components/spinner.svelte';
	import * as Table from '$lib/components/table';
	import type { UsersModel } from '$lib/models/user';
	import { capitalizeFirstLetter } from '$lib/utils';
	import { Button } from 'bits-ui';

	let users: UsersModel = $state([]);
	let promise = $state(fetchUsers());

	async function fetchUsers(): Promise<void> {
		try {
			const data = await GetUsers();
			users = data.items as UsersModel;
		} catch (error) {
			throw error;
		}
	}
</script>

<div class="flex w-full place-content-center">
	<div class="flex w-full max-w-4xl min-w-2xl flex-col gap-6 pt-1">
		<div>
			<Button.Root
				href="/admin/users/create"
				class="bg-background-alt-4 hover:bg-background-alt-5 inline-flex h-10 flex-row items-center gap-2 rounded-md px-5 hover:cursor-pointer"
				aria-label="Toggle password visibility"
			>
				<PlusIcon class="size-5 stroke-[1.5]" />
				Add User
			</Button.Root>
		</div>

		<div class="flex w-full place-content-center">
			{#await promise}
				<div class="flex justify-center pt-10">
					<Spinner class="bg-foreground-alt-2 size-4" />
				</div>
			{:then _}
				<Table.Root class="">
					<Table.Thead>
						<Table.Tr>
							<Table.Th>Username</Table.Th>
							<Table.Th>Name</Table.Th>
							<Table.Th>Role</Table.Th>
							<Table.Th class="w-[1%]" />
						</Table.Tr>
					</Table.Thead>
					<Table.Tbody>
						{#each users as user}
							<Table.Tr class="hover:bg-background-alt-1 duration-200">
								<Table.Td>{user.username}</Table.Td>
								<Table.Td>{user.displayName}</Table.Td>
								<Table.Td>{capitalizeFirstLetter(user.role)}</Table.Td>
								<Table.Td class="flex items-center justify-center">
									<div
										class="hover:bg-background-alt-3 flex h-8 w-8 place-items-center items-center justify-center rounded-lg hover:cursor-pointer"
									>
										<DotsIcon class="size-4" />
									</div>
								</Table.Td>
							</Table.Tr>
						{/each}
					</Table.Tbody>
				</Table.Root>
			{:catch error}
				<div class="flex w-full flex-col items-center gap-2 pt-10">
					<WarningIcon class="text-foreground-error size-10" />
					<span class="text-lg">Failed to fetch users: {error.message}</span>
				</div>
			{/await}
		</div>
	</div>
</div>
