<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	type Props = {
		class?: string;
		children: Snippet;
	};
	let { class: containerClass, children }: Props = $props();
</script>

<header
	class={cn(
		'border-background-alt-3 bg-background-alt-2 flex h-16 items-center justify-between border-b px-3 text-base font-medium',
		containerClass
	)}
>
	{@render children()}
</header>
