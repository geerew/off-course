<script lang="ts">
	import { cn } from '$lib/utils';
	import type { WithoutChild } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { Drawer as DrawerPrimitive } from 'vaul-svelte';

	type Props = WithoutChild<DrawerPrimitive.CloseProps> & {
		ref?: HTMLButtonElement | null;
		class?: string;
		children: Snippet;
	};

	let { ref = $bindable(null), class: containerClass, children, ...restProps }: Props = $props();
</script>

<DrawerPrimitive.Close
	bind:ref
	data-slot="drawer-close"
	class={cn(
		'border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground w-24 cursor-pointer rounded-md border py-2 duration-200 select-none',
		containerClass
	)}
	{...restProps}
>
	{@render children()}
</DrawerPrimitive.Close>
