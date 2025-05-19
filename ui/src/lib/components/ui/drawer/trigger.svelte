<script lang="ts">
	import { cn } from '$lib/utils';
	import type { WithoutChild } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { Drawer as DrawerPrimitive } from 'vaul-svelte';

	type Props = WithoutChild<DrawerPrimitive.TriggerProps> & {
		ref?: HTMLButtonElement | null;
		class?: string;
		children: Snippet;
	};

	let { ref = $bindable(null), class: containerClass, children, ...restProps }: Props = $props();
</script>

<DrawerPrimitive.Trigger
	bind:ref
	data-slot="drawer-trigger"
	class={cn(
		'bg-background-alt-4 hover:bg-background-alt-5 text-foreground-alt-1 hover:text-foreground w-38 cursor-pointer rounded-md py-2 duration-200 select-none',
		containerClass
	)}
	{...restProps}
>
	{@render children()}
</DrawerPrimitive.Trigger>
