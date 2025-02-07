<script lang="ts">
	import { cn } from '$lib/utils';
	import { Button } from 'bits-ui';

	type Props = Button.RootProps & {
		containerClass?: string;
	};

	let { class: containerClass = '', children, ...restProps }: Props = $props();
</script>

<Button.Root
	class={cn(
		'bg-background-primary text-foreground-alt-3 disabled:bg-background-primary-alt-1 group flex h-10 w-full cursor-pointer flex-row place-content-center items-center gap-2 rounded-md duration-200 select-none enabled:hover:brightness-105 disabled:cursor-not-allowed',
		containerClass
	)}
	{...restProps}
>
	{@render children?.()}
</Button.Root>
