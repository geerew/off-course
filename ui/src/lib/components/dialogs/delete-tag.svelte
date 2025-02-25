<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { DeleteTag } from '$lib/api/tag-api';
	import { Spinner } from '$lib/components';
	import { AlertDialog, Button } from '$lib/components/ui';
	import type { TagModel, TagsModel } from '$lib/models/tag-model';
	import { type Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';

	type Props = {
		open?: boolean;
		value: TagModel | TagsModel;
		trigger?: Snippet;
		triggerClass?: string;
		successFn?: () => void;
	};

	let { open = $bindable(false), value, trigger, triggerClass, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let isPosting = $state(false);
	const isArray = Array.isArray(value);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	$effect(() => {
		if (open) {
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doDelete(): Promise<void> {
		isPosting = true;

		try {
			if (isArray) {
				await Promise.all(Object.values(value).map((t) => DeleteTag(t.id)));
				toast.success('Selected tags deleted');
			} else {
				await DeleteTag(value.id);
			}

			successFn?.();
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
		open = false;
	}
</script>

<AlertDialog
	bind:open
	contentProps={{
		interactOutsideBehavior: 'close'
	}}
	{trigger}
	{triggerClass}
>
	{#snippet description()}
		<div class="text-foreground-alt-1 flex flex-col gap-2 text-center">
			{#if isArray && Object.values(value).length > 1}
				<span class="text-lg">Are you sure you want to delete these tags?</span>
			{:else}
				<span class="text-lg">Are you sure you want to delete this tag?</span>
			{/if}
		</div>
	{/snippet}

	{#snippet action()}
		<Button
			disabled={isPosting}
			onclick={doDelete}
			class="bg-background-error disabled:bg-background-error/80 enabled:hover:bg-background-error-alt-1 text-foreground-alt-1 enabled:hover:text-foreground w-24"
		>
			{#if !isPosting}
				Delete
			{:else}
				<Spinner class="bg-foreground-alt-1 size-2" />
			{/if}
		</Button>
	{/snippet}
</AlertDialog>
