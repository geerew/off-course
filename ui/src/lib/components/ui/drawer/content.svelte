<script lang="ts">
	import { cn } from '$lib/utils.js';
	import type { WithoutChild } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { Drawer as DrawerPrimitive } from 'vaul-svelte';
	import DrawerOverlay from './overlay.svelte';

	type Props = WithoutChild<DrawerPrimitive.ContentProps> & {
		ref?: HTMLDivElement | null;
		class?: string;
		handleClass?: string;
		portalProps?: DrawerPrimitive.PortalProps;
		children: Snippet;
	};

	let {
		ref = $bindable(null),
		class: containerClass,
		handleClass = '',
		portalProps,
		children,
		...restProps
	}: Props = $props();
</script>

<DrawerPrimitive.Portal {...portalProps}>
	<DrawerOverlay />
	<DrawerPrimitive.Content
		bind:ref
		data-slot="drawer-content"
		class={cn(
			'group/drawer-content bg-background-alt-1 border-background-alt-3 fixed z-50 flex flex-col',
			'data-[vaul-drawer-direction=top]:inset-x-0 data-[vaul-drawer-direction=top]:top-0 data-[vaul-drawer-direction=top]:mb-24 data-[vaul-drawer-direction=top]:max-h-[80vh] data-[vaul-drawer-direction=top]:rounded-b-lg data-[vaul-drawer-direction=top]:border-b',
			'data-[vaul-drawer-direction=bottom]:inset-x-0 data-[vaul-drawer-direction=bottom]:bottom-0 data-[vaul-drawer-direction=bottom]:mt-24 data-[vaul-drawer-direction=bottom]:max-h-[80vh] data-[vaul-drawer-direction=bottom]:rounded-t-lg data-[vaul-drawer-direction=bottom]:border-t',
			'data-[vaul-drawer-direction=right]:inset-y-0 data-[vaul-drawer-direction=right]:right-0 data-[vaul-drawer-direction=right]:w-3/4 data-[vaul-drawer-direction=right]:border-l data-[vaul-drawer-direction=right]:sm:max-w-sm',
			'data-[vaul-drawer-direction=left]:inset-y-0 data-[vaul-drawer-direction=left]:left-0 data-[vaul-drawer-direction=left]:w-3/4 data-[vaul-drawer-direction=left]:border-r data-[vaul-drawer-direction=left]:sm:max-w-sm',
			containerClass
		)}
		{...restProps}
	>
		<div
			class={cn(
				'bg-background-alt-2 mx-auto mt-3 mb-2 hidden h-2 w-[100px] shrink-0 rounded-full group-data-[vaul-drawer-direction=bottom]/drawer-content:block',
				handleClass
			)}
		></div>
		{@render children?.()}
	</DrawerPrimitive.Content>
</DrawerPrimitive.Portal>
