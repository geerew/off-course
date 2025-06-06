<script lang="ts">
	import { cn } from '$lib/utils';
	import { Dialog, type WithoutChild } from 'bits-ui';
	import type { Snippet } from 'svelte';

	type Props = WithoutChild<Dialog.ContentProps> & {
		class?: string;
		children: Snippet;
	};

	let { class: containerClass, children, ...restProps }: Props = $props();
</script>

<Dialog.Content
	class={cn(
		'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 data-[state=open]:slide-in-from-top-5 bg-background-alt-1 data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:slide-out-to-top-5 fixed top-20 left-1/2 z-50 w-[calc(100%-4rem)] max-w-[calc(100vw-4rem)] -translate-x-1/2 overflow-hidden rounded-lg data-[state=closed]:duration-200 data-[state=open]:duration-200',
		containerClass
	)}
	{...restProps}
>
	{@render children()}
</Dialog.Content>
