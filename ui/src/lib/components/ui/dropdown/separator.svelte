<script lang="ts">
	import { cn } from '$lib/utils';
	import { DropdownMenu, type DropdownMenuSeparatorProps, type WithoutChild } from 'bits-ui';

	type Props = WithoutChild<DropdownMenuSeparatorProps> & {
		ref?: HTMLButtonElement | null;
		class?: string;
	};

	let { ref = $bindable(null), class: containerClass, ...restProps }: Props = $props();
</script>

<DropdownMenu.Separator
	bind:ref
	class={cn('bg-background-alt-3 h-px w-full', containerClass)}
	{...restProps}
/>
