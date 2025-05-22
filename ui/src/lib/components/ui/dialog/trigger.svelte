<script lang="ts">
	import { cn } from '$lib/utils';
	import { Dialog, type WithoutChild } from 'bits-ui';
	import type { Snippet } from 'svelte';

	type Props = WithoutChild<Dialog.TriggerProps> & {
		class?: string;
		children: Snippet;
	};

	let { class: containerClass, children, ...restProps }: Props = $props();
</script>

<Dialog.Trigger
	class={cn(
		'bg-background-alt-4 disabled:bg-background-alt-3 hover:bg-background-alt-5 disabled:text-foreground-alt-3 disabled:hover:bg-background-alt-3 text-foreground-alt-1 hover:text-foreground w-38 cursor-pointer rounded-md py-2 duration-200 select-none disabled:cursor-auto',
		containerClass
	)}
	{...restProps}
>
	{@render children()}
</Dialog.Trigger>
