<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	type Props = {
		class?: string;
		children: Snippet;
	};
	let { class: containerClass, children }: Props = $props();
</script>

<footer
	class={cn(
		'bg-background-alt-2 border-background-alt-3 flex h-16 w-full shrink-0 items-center justify-end gap-2 border-t px-5 py-2.5',
		containerClass
	)}
>
	{@render children()}
</footer>
