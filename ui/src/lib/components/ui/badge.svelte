<script lang="ts">
	import { cn } from '$lib/utils';
	import type { WithElementRef } from 'bits-ui';
	import type { HTMLAttributes } from 'svelte/elements';

	let {
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLSpanElement>> = $props();
</script>

<span
	class={cn(
		'focus:ring-ring bg-background-alt-4 text-foreground-alt-1 inline-flex items-center rounded-full px-2.5 py-0.5 text-xs transition-colors focus:ring-2 focus:ring-offset-2 focus:outline-none',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</span>
