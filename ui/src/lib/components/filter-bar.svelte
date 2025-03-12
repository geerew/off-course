<script lang="ts">
	import { cn } from '$lib/utils';
	import { XIcon } from './icons';
	import { Button, Input } from './ui';

	type Props = {
		value: string;
		onUpdate?: () => void;
	};

	let { value = $bindable(), onUpdate }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let filterEnabled = $state(false);
</script>

<div class="relative flex flex-1">
	<Input
		bind:value
		placeholder="Filter"
		class="border-background-alt-5 peer focus:border-foreground-alt-2 h-10 border bg-transparent focus:bg-transparent"
		onkeyup={async (e) => {
			if (e.key === 'Enter') {
				filterEnabled = value ? true : false;
				onUpdate?.();
			}
		}}
	/>

	<Button
		class={cn(
			'bg-background-alt-4 text-foreground-alt-2 enabled:hover:text-foreground-alt-1 enabled:hover:bg-background-alt-6 absolute top-1/2 right-3 bottom-0 size-6 -translate-y-1/2 p-0',
			!value && !filterEnabled && 'cursor-default opacity-0'
		)}
		onclick={() => {
			filterEnabled = false;
			value = '';
			onUpdate?.();
		}}
	>
		<XIcon class="  size-4 stroke-2" />
	</Button>
</div>
