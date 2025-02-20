<script lang="ts">
	import { cn } from '$lib/utils';
	import { Dialog, type WithoutChild } from 'bits-ui';
	import type { Snippet } from 'svelte';

	type Props = Dialog.RootProps & {
		trigger?: Snippet;
		triggerClass?: string;
		content: Snippet;
		contentProps?: Omit<WithoutChild<Dialog.ContentProps>, 'class'>;
		contentClass?: string;
		action?: Snippet;
	};

	let {
		open = $bindable(false),
		trigger,
		triggerClass,
		content,
		contentProps,
		contentClass,
		action,
		...restProps
	}: Props = $props();
</script>

<Dialog.Root bind:open {...restProps}>
	{#if trigger}
		<Dialog.Trigger
			class={cn(
				'bg-background-alt-4 hover:bg-background-alt-5 text-foreground-alt-1 hover:text-foreground w-38 cursor-pointer rounded-md py-2 duration-200 select-none',
				triggerClass
			)}
		>
			{@render trigger()}
		</Dialog.Trigger>
	{/if}

	<Dialog.Portal>
		<Dialog.Overlay
			class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/60"
		/>

		<Dialog.Content
			class={cn(
				'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 data-[state=open]:slide-in-from-top-5 bg-background-alt-1 data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:slide-out-to-top-5 fixed top-20 left-1/2 z-50 w-[calc(100%-4rem)] max-w-[calc(100vw-4rem)] -translate-x-1/2 overflow-hidden rounded-lg data-[state=closed]:duration-200 data-[state=open]:duration-200',
				contentClass
			)}
			{...contentProps}
		>
			{@render content()}

			<footer
				class="bg-background-alt-2 border-background-alt-3 flex h-16 w-full shrink-0 items-center justify-end gap-2 border-t px-5 py-2.5"
			>
				<Dialog.Close
					type="button"
					class="border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground w-24 cursor-pointer rounded-md border py-2 duration-200 select-none"
				>
					Close
				</Dialog.Close>

				{@render action?.()}
			</footer>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
