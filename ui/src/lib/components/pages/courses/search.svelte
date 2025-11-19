<script lang="ts">
	import { tick } from 'svelte';
	import { Debounced } from 'runed';
	import { SearchIcon, XIcon } from '$lib/components/icons';
	import { Button, Input } from '$lib/components/ui';
	import { cn } from '$lib/utils';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		value?: string;
		disabled?: boolean;
		onApply: () => void;
	};

	let { value = $bindable(''), disabled = false, onApply }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Reference to the search input element
	let searchCoursesEl = $state<HTMLInputElement>();

	// Internal state to track the last applied value
	let appliedValue = $state('');

	// Debounced value for automatic filter application
	const valueDebounced = new Debounced(() => value, 250);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Automatically apply filter when debounced value changes
	$effect(() => {
		const debouncedValue = valueDebounced.current;

		// Skip if value hasn't actually changed
		if (debouncedValue === appliedValue) return;

		// Update applied value and trigger filter
		appliedValue = debouncedValue;
		onApply();
	});
</script>

<div class="group relative flex flex-1 sm:w-96 sm:flex-none">
	<Button
		variant="ghost"
		class="text-foreground-alt-3 group-focus-within:text-foreground-alt-1 absolute left-2.5 top-1/2 -translate-y-1/2 transform cursor-text rounded-full p-0 hover:bg-transparent"
		{disabled}
		onclick={() => {
			if (!searchCoursesEl) return;
			searchCoursesEl.focus();
		}}
	>
		<SearchIcon class="size-5" />
	</Button>

	<Input
		bind:ref={searchCoursesEl}
		bind:value
		placeholder="Search courses..."
		class={cn(
			'placeholder:text-foreground-alt-3 focus:bg-alt-3 h-10 border-b pe-5 ps-10 text-sm',
			appliedValue && 'border-b-background-primary-alt-1'
		)}
		{disabled}
	/>

	{#if value}
		<Button
			variant="ghost"
			class="hover:bg-background-alt-2 text-foreground-alt-2 hover:text-foreground absolute right-1 top-1/2 h-auto -translate-y-1/2 transform rounded-md p-1"
			onclick={async () => {
				value = '';
				appliedValue = '';
				// Wait for reactive updates to complete before applying filter
				await tick();

				// Always call onApply when clearing to ensure filter updates
				onApply();
			}}
		>
			<XIcon class="size-5" />
		</Button>
	{/if}
</div>
