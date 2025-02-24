<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLInputAttributes } from 'svelte/elements';

	type Props = HTMLInputAttributes & {
		value: string;
		ref?: HTMLInputElement;
		class?: string;
	};

	let {
		value = $bindable(''),
		ref = $bindable(),
		class: containerClass = '',
		...restProps
	}: Props = $props();
</script>

<input
	bind:value
	bind:this={ref}
	class={cn(
		'bg-background-alt-3 placeholder:text-foreground-alt-2 focus:bg-background-alt-4 h-11.5 w-full rounded-md border border-transparent px-2.5 ring-0 duration-250 ease-in-out placeholder:text-sm placeholder:tracking-wide focus:outline-none',
		containerClass
	)}
	{...restProps}
/>
