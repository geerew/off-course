<script lang="ts">
	import { cn } from '$lib/utils.js';
	import { Popover, type PopoverContentProps, type WithoutChild } from 'bits-ui';
	import type { Snippet } from 'svelte';

	type Props = WithoutChild<PopoverContentProps> & {
		ref?: HTMLDivElement | null;
		class?: string;
		portalProps?: Popover.PortalProps;
		children: Snippet;
	};

	let {
		ref = $bindable(null),
		class: containerClass,
		portalProps,
		children,
		...restProps
	}: Props = $props();
</script>

<Popover.Portal {...portalProps}>
	<Popover.Content
		bind:ref
		align="end"
		sideOffset={2}
		class={cn(
			'bg-background border-background-alt-5 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2 flex w-36 flex-col gap-1 rounded-md border p-1.5 text-sm outline-none select-none data-[side=bottom]:translate-y-1 data-[side=top]:-translate-y-1',
			containerClass
		)}
		{...restProps}
	>
		{@render children?.()}
	</Popover.Content>
</Popover.Portal>
