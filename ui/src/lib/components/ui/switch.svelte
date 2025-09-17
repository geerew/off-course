<script lang="ts">
	import { cn } from '$lib/utils';
	import { Label, Switch as SwitchPrimitive, useId, type WithoutChildrenOrChild } from 'bits-ui';
	import type { ClassValue } from 'clsx';

	let {
		id = useId(),
		ref = $bindable(null),
		checked = $bindable(false),
		labelText,
		labelBefore = true,
		class: containerClass,
		thumbClass: thumbClass,
		...restProps
	}: WithoutChildrenOrChild<SwitchPrimitive.RootProps> & {
		labelText: string;
		labelBefore?: boolean;
		thumbClass?: ClassValue;
	} = $props();
</script>

{#if labelBefore}
	<Label.Root for={id}>{labelText}</Label.Root>
{/if}
<SwitchPrimitive.Root
	bind:checked
	bind:ref
	{id}
	class={cn(
		'focus-visible:ring-foreground focus-visible:ring-offset-background data-[state=checked]:bg-background-primary-alt-1 data-[state=unchecked]:bg-background-alt-6 data-[state=unchecked]:shadow-mini-inset peer inline-flex h-6 min-h-6 w-12 shrink-0 cursor-pointer items-center rounded-full px-[3px] transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-hidden disabled:cursor-not-allowed disabled:opacity-50',
		containerClass
	)}
	{...restProps}
>
	<SwitchPrimitive.Thumb
		class={cn(
			'bg-foreground pointer-events-none block size-5 shrink-0 rounded-full shadow transition-transform data-[state=checked]:translate-x-5.5 data-[state=unchecked]:translate-x-0',
			thumbClass
		)}
	/>
</SwitchPrimitive.Root>

{#if !labelBefore}
	<Label.Root for={id}>{labelText}</Label.Root>
{/if}
