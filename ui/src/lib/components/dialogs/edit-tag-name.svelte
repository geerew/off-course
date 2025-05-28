<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { UpdateTag } from '$lib/api/tag-api';
	import { Spinner } from '$lib/components';
	import { Button, Dialog, Input } from '$lib/components/ui';
	import type { TagModel } from '$lib/models/tag-model';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';

	type Props = {
		open?: boolean;
		value: TagModel;
		trigger?: Snippet;
		successFn?: () => void;
	};

	let { open = $bindable(false), value = $bindable(), trigger, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let inputEl = $state<HTMLInputElement>();
	let newValue = $state<string>('');
	let isPosting = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			newValue = '';
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doUpdate(e: Event) {
		e.preventDefault();
		isPosting = true;

		try {
			await UpdateTag(value.id, { tag: newValue });
			value.tag = newValue;
			open = false;
			successFn?.();
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
	}
</script>

<Dialog.Root bind:open {trigger}>
	<Dialog.Content
		class="max-w-xs"
		interactOutsideBehavior="close"
		onOpenAutoFocus={(e) => {
			e.preventDefault();
			inputEl?.focus();
		}}
		onCloseAutoFocus={(e) => {
			e.preventDefault();
		}}
	>
		<form onsubmit={doUpdate}>
			<main class="flex flex-col gap-2.5 p-5">
				<div>Tag:</div>
				<Input
					bind:ref={inputEl}
					bind:value={newValue}
					name="tag name"
					type="text"
					placeholder={value.tag}
				/>
			</main>

			<Dialog.Footer>
				<Dialog.CloseButton>Close</Dialog.CloseButton>

				<Button
					type="submit"
					variant="default"
					class="w-24"
					disabled={newValue === '' || isPosting}
				>
					{#if isPosting}
						<Spinner class="bg-background-alt-4  size-2" />
					{:else}
						Update
					{/if}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
