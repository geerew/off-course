<script lang="ts">
	import { cn } from '$lib/utils';
	import { Tooltip, type WithoutChildren } from 'bits-ui';
	import { type Snippet } from 'svelte';
	import { fly } from 'svelte/transition';

	type Props = WithoutChildren<
		Tooltip.RootProps & {
			trigger: Snippet;
			triggerProps?: Omit<WithoutChildren<Tooltip.TriggerProps>, 'class'>;
			triggerClass?: string;
			content: Snippet;
			contentProps?: Omit<WithoutChildren<Tooltip.ContentProps>, 'class'>;
			contentClass?: string;
		}
	>;

	let {
		open = $bindable(false),
		trigger,
		triggerProps,
		triggerClass,
		content,
		contentProps,
		contentClass,
		...restProps
	}: Props = $props();
</script>

<Tooltip.Root bind:open {...restProps}>
	<Tooltip.Trigger class={triggerClass} {...triggerProps}>
		{@render trigger()}
	</Tooltip.Trigger>

	<Tooltip.Portal>
		<Tooltip.Content
			class={cn(
				'bg-background border-background-alt-5 flex flex-col gap-1 rounded-md border px-1.5 py-1 outline-none select-none',
				contentClass
			)}
			{...contentProps}
			forceMount
		>
			{#snippet child({ wrapperProps, props, open })}
				{#if open}
					<div {...wrapperProps}>
						<div {...props} transition:fly={{ duration: 150 }}>
							{@render content()}
						</div>
					</div>
				{/if}
			{/snippet}
		</Tooltip.Content>
	</Tooltip.Portal>
</Tooltip.Root>
