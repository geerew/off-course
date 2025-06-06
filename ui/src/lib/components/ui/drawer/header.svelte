<script lang="ts">
	import { cn } from '$lib/utils';
	import type { WithElementRef } from 'bits-ui';
	import type { HTMLAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLDivElement>> = $props();
</script>

<div
	bind:this={ref}
	data-slot="drawer-header"
	class={cn(
		'border-background-alt-3 bg-background-alt-2 flex h-16 shrink-0 items-center justify-between border-b px-3 text-base font-medium',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</div>
