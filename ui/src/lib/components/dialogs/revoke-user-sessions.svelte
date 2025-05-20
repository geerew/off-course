<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { RevokeUserSessions } from '$lib/api/user-api';
	import { Spinner } from '$lib/components';
	import { Button, Dialog } from '$lib/components/ui';
	import type { UserModel, UsersModel } from '$lib/models/user-model';
	import { type Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';

	type Props = {
		open?: boolean;
		value: UserModel | UsersModel;
		trigger?: Snippet;
		triggerClass?: string;
		successFn?: () => void;
	};

	let { open = $bindable(false), value, trigger, triggerClass, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let isPosting = $state(false);
	const multipleUsers = Array.isArray(value);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doRevoke(): Promise<void> {
		isPosting = true;

		try {
			if (multipleUsers) {
				await Promise.all(Object.values(value).map((u) => RevokeUserSessions(u.id)));
				toast.success('Selected users deleted');
			} else {
				await RevokeUserSessions(value.id);
			}

			successFn?.();
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
		open = false;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content interactOutsideBehavior="close" class="w-lg">
		<div class="bg-background-alt-1 overflow-hidden rounded-lg">
			<Dialog.Alert>
				<div class="text-foreground-alt-1 flex flex-col gap-2 text-center">
					{#if multipleUsers && Object.values(value).length > 1}
						<span class="text-lg">
							Are you sure you want to continue revoking all sessions for these users?
						</span>
					{:else}
						<span class="text-lg">
							Are you sure you want to continue revoking all sessions for this user?
						</span>
					{/if}
				</div>
			</Dialog.Alert>

			<Dialog.Footer>
				<Dialog.CloseButton>Close</Dialog.CloseButton>
				<Button
					disabled={isPosting}
					onclick={doRevoke}
					class="bg-background-error enabled:hover:bg-background-error-alt-1 text-foreground-alt-1 enabled:hover:text-foreground w-24"
				>
					{#if isPosting}
						<Spinner class="bg-foreground-alt-1 size-2" />
					{:else}
						Delete
					{/if}
				</Button>
			</Dialog.Footer>
		</div>
	</Dialog.Content>
</Dialog.Root>
