<script lang="ts">
	import { cn } from '$lib/utils';
	import { DropdownMenu, type DropdownMenuTriggerProps, type WithoutChild } from 'bits-ui';
	import type { Snippet } from 'svelte';

	type Props = WithoutChild<DropdownMenuTriggerProps> & {
		ref?: HTMLButtonElement | null;
		class?: string;
		children: Snippet;
	};

	let { ref = $bindable(null), class: containerClass, children, ...restProps }: Props = $props();
</script>

<DropdownMenu.Trigger
	bind:ref
	class={cn(
		'border-background-alt-4 data-[state=open]:border-foreground-alt-3 hover:border-foreground-alt-3 disabled:text-foreground-alt-3 disabled:hover:border-background-alt-4 inline-flex h-10 w-full items-center justify-between rounded-md border px-2.5 text-sm duration-200 select-none hover:cursor-pointer disabled:cursor-not-allowed',
		containerClass
	)}
	{...restProps}
>
	{@render children()}
</DropdownMenu.Trigger>
