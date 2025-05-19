<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';

	type Props = HTMLAttributes<HTMLElement> & {
		class?: string;
		children?: Snippet;
	};

	let { class: containerClass = '', children, ...restProps }: Props = $props();
</script>

<div
	role="rowgroup"
	class={cn('text-foreground-alt-3 contents text-xs uppercase', containerClass)}
	{...restProps}
>
	{@render children?.()}
</div>
