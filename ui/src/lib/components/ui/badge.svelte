<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLAttributes } from 'svelte/elements';

	type Props = HTMLAttributes<HTMLSpanElement> & {
		class?: string;
		children?: () => void;
	};

	let { class: className, children, ...restProps }: Props = $props();
</script>

<span
	class={cn(
		'focus:ring-ring bg-background-alt-4 text-foreground-alt-1 inline-flex w-auto flex-row items-center rounded-md px-2.5 py-0.5 text-xs whitespace-nowrap transition-colors focus:ring-2 focus:ring-offset-2 focus:outline-none',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</span>
