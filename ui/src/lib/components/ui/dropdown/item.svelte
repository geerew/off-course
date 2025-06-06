<script lang="ts">
	import { cn } from '$lib/utils';
	import { DropdownMenu, type DropdownMenuItemProps, type WithoutChild } from 'bits-ui';
	import type { Snippet } from 'svelte';

	type Props = WithoutChild<DropdownMenuItemProps> & {
		ref?: HTMLButtonElement | null;
		class?: string;
		children: Snippet;
	};

	let { ref = $bindable(null), class: containerClass, children, ...restProps }: Props = $props();
</script>

<DropdownMenu.Item
	bind:ref
	class={cn(
		'text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 data-disabled:text-foreground-alt-3 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none disabled:opacity-50 data-disabled:cursor-default data-disabled:hover:bg-transparent',
		containerClass
	)}
	{...restProps}
>
	{@render children()}
</DropdownMenu.Item>
