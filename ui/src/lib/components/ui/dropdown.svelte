<script lang="ts">
	import { cn } from '$lib/utils';
	import { DropdownMenu, type WithoutChildren } from 'bits-ui';
	import type { Snippet } from 'svelte';

	type Props = WithoutChildren<
		DropdownMenu.RootProps & {
			open?: boolean;
			trigger?: Snippet;
			triggerProps?: Omit<WithoutChildren<DropdownMenu.TriggerProps>, 'class'>;
			triggerClass?: string;
			content: Snippet;
			contentProps?: Omit<WithoutChildren<DropdownMenu.ContentProps>, 'class'>;
			contentClass?: string;
		}
	>;

	let {
		open = $bindable(false),
		trigger,
		triggerProps,
		triggerClass,
		content,
		contentProps,
		contentClass,
		...restProps
	}: Props = $props();
</script>

<DropdownMenu.Root bind:open {...restProps}>
	{#if trigger}
		<DropdownMenu.Trigger
			class={cn(
				'border-background-alt-4 data-[state=open]:border-foreground-alt-2 hover:border-foreground-alt-2 disabled:text-foreground-alt-2 disabled:hover:border-background-alt-4 inline-flex h-10 items-center justify-between rounded-md border px-2.5 text-sm duration-200 select-none hover:cursor-pointer disabled:cursor-not-allowed',
				triggerClass
			)}
			{...triggerProps}
		>
			{@render trigger()}
		</DropdownMenu.Trigger>
	{/if}

	<DropdownMenu.Content
		align="end"
		sideOffset={2}
		class={cn(
			'bg-background border-background-alt-5 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2 flex w-36 flex-col gap-1 rounded-md border outline-none select-none data-[side=bottom]:translate-y-1 data-[side=top]:-translate-y-1',
			contentClass
		)}
		{...contentProps}
	>
		{@render content()}
	</DropdownMenu.Content>
</DropdownMenu.Root>
