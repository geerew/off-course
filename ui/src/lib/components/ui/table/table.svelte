<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';
	import type { HTMLTableAttributes } from 'svelte/elements';

	type Props = HTMLTableAttributes & {
		class?: string;
		children?: Snippet;
	};

	let { class: containerClass = '', children, ...restProps }: Props = $props();
</script>

<div role="table" class={cn('grid w-full border-collapse', containerClass)}>
	{@render children?.()}
</div>
