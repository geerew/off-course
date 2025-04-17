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
		'bg-background-primary-alt-1 text-foreground-alt-5 group enabled:hover:bg-background-primary flex h-10 w-full cursor-pointer flex-row place-content-center items-center gap-2 rounded-md duration-200 select-none disabled:cursor-not-allowed disabled:opacity-50',
		containerClass
	)}
	{...restProps}
>
	{@render children?.()}
</Button.Root>
