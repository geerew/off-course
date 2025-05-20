<script lang="ts">
	import { cn } from '$lib/utils';
	import type { WithoutChild } from 'bits-ui';
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

	type Props = WithoutChild<Dialog.CloseProps> & {
		ref?: HTMLButtonElement | null;
		class?: string;
		children: Snippet;
	};

	let { ref = $bindable(null), class: containerClass, children, ...restProps }: Props = $props();
</script>

<Dialog.Close
	bind:ref
	data-slot="dialog-close"
	type="button"
	class={cn(
		'border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground w-24 cursor-pointer rounded-md border py-2 duration-200 select-none',
		containerClass
	)}
	{...restProps}
>
	{@render children()}
</Dialog.Close>
