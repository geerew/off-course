<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';
	import type { HTMLThAttributes } from 'svelte/elements';

	type Props = HTMLThAttributes & {
		class?: string;
		children?: Snippet;
	};

	let { class: containerClass = '', children, ...restProps }: Props = $props();
</script>

<th class={cn('px-4 py-4 text-start font-semibold', containerClass)} {...restProps}>
	{@render children?.()}
</th>
