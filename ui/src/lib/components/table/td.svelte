<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';
	import type { HTMLTdAttributes } from 'svelte/elements';

	type Props = HTMLTdAttributes & {
		class?: string;
		children?: Snippet;
	};

	let { class: containerClass = '', children, ...restProps }: Props = $props();
</script>

<td class={cn('h-14 px-4 text-start text-sm', containerClass)} {...restProps}>
	{@render children?.()}
</td>
