<script lang="ts">
	import { Checkbox as CheckboxPrimitive, type WithoutChildrenOrChild } from 'bits-ui';

	import { cn } from '$lib/utils.js';
	import { MinusIcon, TickIcon } from '../icons';

	let {
		ref = $bindable(null),
		checked = $bindable(false),
		indeterminate = $bindable(false),
		class: className,
		...restProps
	}: WithoutChildrenOrChild<CheckboxPrimitive.RootProps> = $props();
</script>

<CheckboxPrimitive.Root
	bind:ref
	class={cn(
		'border-foreground-alt-2 ring-offset-background focus-visible:ring-ring data-[state=checked]:bg-background-primary-alt-1 data-[state=checked]:border-background-primary-alt-1 data-[state=checked]:text-background data-[state=indeterminate]:bg-background-primary-alt-1 data-[state=indeterminate]:border-background-primary-alt-1 data-[state=indeterminate]:text-background peer box-content size-4 shrink-0 rounded-sm border hover:cursor-pointer focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 data-[disabled=true]:cursor-not-allowed data-[disabled=true]:opacity-50',
		className
	)}
	bind:checked
	bind:indeterminate
	{...restProps}
>
	{#snippet children({ checked, indeterminate })}
		<div class="flex size-4 items-center justify-center text-current">
			{#if indeterminate}
				<MinusIcon class="size-3.5 stroke-2" />
			{:else}
				<TickIcon class={cn('size-3.5 stroke-2', !checked && 'text-transparent')} />
			{/if}
		</div>
	{/snippet}
</CheckboxPrimitive.Root>
