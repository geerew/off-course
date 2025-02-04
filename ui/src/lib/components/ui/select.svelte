<script lang="ts">
	import { cn } from '$lib/utils';
	import { Select, type WithoutChildren } from 'bits-ui';
	import { TickIcon } from '../icons';
	import RightChevron from '../icons/right-chevron.svelte';

	type Props = Omit<
		WithoutChildren<
			Select.RootProps & {
				placeholder?: string;
				items: { value: string; label: string; disabled?: boolean }[];
				triggerClass?: string;
				contentProps?: Omit<WithoutChildren<Select.ContentProps>, 'class'>;
				contentClass?: string;
				itemClass?: string;
				value?: string;
				onValueChange?: (value: string) => void;
				type: string;
			}
		>,
		'type'
	>;

	let {
		value = $bindable(''),
		placeholder,
		items,
		triggerClass,
		contentProps,
		contentClass,
		itemClass,
		...restProps
	}: Props = $props();

	const selectedLabel = $derived(items.find((item) => item.value === value)?.label);
</script>

<Select.Root type="single" bind:value {...restProps}>
	<Select.Trigger
		class={cn(
			'border-background-alt-4 data-[state=open]:border-foreground-alt-2 hover:border-foreground-alt-2 data-placeholder:text-foreground-alt-2 inline-flex h-11.5 items-center justify-between rounded-md border px-2.5 text-sm duration-200 select-none hover:cursor-pointer [&[data-state=open]>svg]:rotate-90',
			triggerClass
		)}
	>
		<span>
			{selectedLabel ? selectedLabel : placeholder}
		</span>
		<RightChevron class="stroke-foreground-alt-2 size-4.5 duration-200" />
	</Select.Trigger>
	<Select.Portal>
		<Select.Content
			class={cn(
				'bg-background border-foreground-alt-2 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 w-[var(--bits-select-anchor-width)] min-w-[var(--bits-select-anchor-width)] rounded-md border py-3 outline-none select-none data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1',
				contentClass
			)}
			{...contentProps}
		>
			<Select.Viewport>
				{#each items as { value, label, disabled } (value)}
					<Select.Item
						{value}
						{label}
						{disabled}
						class={cn(
							'data-[highlighted]:bg-background-alt-2 flex h-10 w-full cursor-pointer items-center justify-between px-2.5 text-sm duration-75 outline-none select-none data-[disabled]:opacity-50',
							itemClass
						)}
					>
						{#snippet children({ selected })}
							{label}
							{#if selected}
								<TickIcon class="size-4.5" />
							{/if}
						{/snippet}
					</Select.Item>
				{/each}
			</Select.Viewport>
		</Select.Content>
	</Select.Portal>
</Select.Root>
