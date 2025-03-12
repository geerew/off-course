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

<thead
	class={cn(
		'border-background-alt-3 text-foreground-alt-2 border-b text-xs uppercase',
		containerClass
	)}
	{...restProps}
>
	{@render children?.()}
</thead>
