<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	interface Props {
		class?: string;
		children?: Snippet;
	}

	let { class: containerClass = '', children }: Props = $props();
</script>

<th class={cn('px-2.5 py-4 text-start font-semibold', containerClass)}>
	{@render children?.()}
</th>
