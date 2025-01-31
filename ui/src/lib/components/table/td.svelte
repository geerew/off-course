<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	interface Props {
		class?: string;
		children?: Snippet;
	}

	let { class: containerClass = '', children }: Props = $props();
</script>

<td class={cn('h-14 px-2.5 text-start text-sm', containerClass)}>
	{@render children?.()}
</td>
