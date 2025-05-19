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
	role="columnheader"
	class={cn(
		'border-background-alt-3 flex h-12 items-center justify-center border-b',
		containerClass
	)}
	{...restProps}
>
	{@render children?.()}
</div>
