<script lang="ts">
	import { cn } from '$lib/utils';
	import { Separator } from 'bits-ui';
	import { WarningIcon } from './icons';

	type Props = {
		class?: string;
		contentClass?: string;
		message: string;
	};

	let { class: containerClass, contentClass, message }: Props = $props();
</script>

<div class={cn('flex w-full justify-center pt-20', containerClass)}>
	<div
		class={cn(
			'border-background-alt-5 min-w-sm flex max-w-lg flex-col place-items-center gap-4 rounded-lg border p-5',
			contentClass
		)}
	>
		<WarningIcon class="text-foreground-error size-10" />

		<p class="text-foreground-alt-1 text-center text-lg">{message}</p>

		<Separator.Root class="bg-background-alt-5 h-px w-60 shrink-0" />

		<div class="text-foreground-alt-3 flex justify-center gap-2.5">Please try again</div>
	</div>
</div>
