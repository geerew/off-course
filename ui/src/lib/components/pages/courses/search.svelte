<script lang="ts">
	import { SearchIcon, XIcon } from '$lib/components/icons';
	import { Button, Input } from '$lib/components/ui';
	import { cn } from '$lib/utils';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		value?: string;
		disabled?: boolean;
		onApply: () => void;
	};

	let { value = $bindable(''), disabled = false, onApply }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Reference to the search input element
	let searchCoursesEl = $state<HTMLInputElement>();

	// Internal state to track the last applied value
	let appliedValue = $state('');
</script>

<div class="group relative w-96">
	<Button
		variant="ghost"
		class="text-foreground-alt-3 group-focus-within:text-foreground-alt-1 absolute top-1/2 left-2.5 -translate-y-1/2 transform cursor-text rounded-full p-0 hover:bg-transparent"
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
			'placeholder:text-foreground-alt-3 focus:bg-alt-3 h-10 border-b-2 ps-10 pe-5 text-sm',
			appliedValue && 'border-b-background-primary-alt-1'
		)}
		{disabled}
		onkeydown={(e: KeyboardEvent) => {
			if (e.key === 'Enter') {
				e.preventDefault();

				// Do nothing when the value hasn't changed
				if (value == appliedValue) return;

				appliedValue = value;
				onApply();
			}
		}}
	/>

	{#if value}
		<Button
			variant="ghost"
			class="hover:bg-background-alt-2 text-foreground-alt-2 hover:text-foreground absolute top-1/2 right-1 h-auto -translate-y-1/2 transform rounded-md p-1"
			onclick={() => {
				value = '';
				if (appliedValue) onApply();
				appliedValue = '';
			}}
		>
			<XIcon class="size-5" />
		</Button>
	{/if}
</div>
