<script lang="ts">
	import { LoaderCircleIcon, RightChevronIcon } from '$lib/components/icons';
	import { Button, Dropdown } from '$lib/components/ui';
	import { cn } from '$lib/utils';

	let progressStates = [
		{ label: 'Not Started', value: 'not started' },
		{ label: 'Started', value: 'started' },
		{ label: 'Completed', value: 'completed' }
	] as const;

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		value?: string;
		defaultProgress?: Array<(typeof progressStates)[number]['value']>;
		disabled?: boolean;
		onApply: () => void;
	};

	let { value = $bindable(''), defaultProgress = [], disabled = false, onApply }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let selected = $state<string[]>(defaultProgress);
</script>

<div class="flex h-10 items-center gap-3 rounded-lg">
	<Dropdown.Root>
		<Dropdown.Trigger
			class={cn(
				'w-36 [&[data-state=open]>svg]:rotate-90 ',
				value &&
					'data-[state=open]:border-b-background-primary-alt-1 hover:border-b-background-primary-alt-1 border-b-background-primary-alt-1 border-b-2'
			)}
			{disabled}
		>
			<div class="flex items-center gap-1.5">
				<LoaderCircleIcon class="size-4 stroke-2" />

				<span>Progress</span>
			</div>
			<RightChevronIcon class="stroke-foreground-alt-3 size-4.5 duration-200" />
		</Dropdown.Trigger>

		<Dropdown.Content class="w-50" align="start">
			<div class="flex flex-col gap-1">
				<div class="flex flex-row items-center justify-between px-1.5">
					<span class="text-background-primary-alt-1 text-base font-semibold">Progress</span>
					<Button
						variant="ghost"
						class="text-foreground-alt-3 hover:text-foreground-alt-2 p-0 text-sm hover:bg-transparent"
						onclick={() => {
							selected = [];
							value = '';
							onApply();
						}}
					>
						clear
					</Button>
				</div>

				<Dropdown.CheckboxGroup
					bind:value={selected}
					onValueChange={() => {
						value = selected.map((v) => `progress:"${v}"`).join(' OR ');
						onApply();
					}}
				>
					{#each progressStates as state}
						<Dropdown.CheckboxItem value={state.value} closeOnSelect={false}>
							{state.label}
						</Dropdown.CheckboxItem>
					{/each}
				</Dropdown.CheckboxGroup>
			</div>
		</Dropdown.Content>
	</Dropdown.Root>
</div>
