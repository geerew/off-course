<script lang="ts">
	import { TickIcon } from '$lib/components/icons';
	import { cn } from '$lib/utils';
	import { DropdownMenu, type WithoutChild } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		children: childrenProp,
		...restProps
	}: WithoutChild<DropdownMenu.CheckboxItemProps> = $props();
</script>

<DropdownMenu.CheckboxItem
	bind:ref
	data-slot="dropdown-menu-checkbox-item"
	class={cn(
		'text-foreground-alt-1 hover:bg-background-alt-2 hover:text-foreground inline-flex w-full cursor-pointer select-none items-center gap-2.5 rounded-md px-1 py-1 text-sm duration-200',
		className
	)}
	{...restProps}
>
	{#snippet children({ checked, indeterminate })}
		<span class="inline-flex size-3.5 shrink-0 items-center justify-center">
			{#if checked}
				<TickIcon class="size-3.5 stroke-2" />
			{:else}
				<span class="size-3.5"></span>
			{/if}
		</span>

		<div class="wrap-break-word min-w-0">
			{@render childrenProp?.({ checked, indeterminate })}
		</div>
	{/snippet}
</DropdownMenu.CheckboxItem>
