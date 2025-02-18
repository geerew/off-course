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

<table class={cn('w-full border-collapse', containerClass)} {...restProps}>
	{@render children?.()}
</table>
