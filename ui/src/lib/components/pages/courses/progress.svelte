<script module lang="ts">
	export let progressStates = [
		{ label: 'Not Started', value: 'not started' },
		{ label: 'Started', value: 'started' },
		{ label: 'Completed', value: 'completed' }
	] as const;

	export type ProgressState = (typeof progressStates)[number]['value'];
</script>

<script lang="ts">
	import { LoaderCircleIcon, RightChevronIcon, TickIcon } from '$lib/components/icons';
	import { Button, Dropdown } from '$lib/components/ui';
	import { cn } from '$lib/utils';
	import { Accordion, Checkbox, Label, useId } from 'bits-ui';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		type: 'dropdown' | 'accordion';
		value?: string;
		selected?: Array<ProgressState>;
		disabled?: boolean;
		onApply: () => void;
	};

	let {
		type,
		value = $bindable(''),
		selected = $bindable([]),
		disabled = false,
		onApply
	}: Props = $props();
</script>

{#if type === 'dropdown'}
	<div class="flex h-10 items-center gap-3 rounded-lg">
		<Dropdown.Root>
			<Dropdown.Trigger
				class={cn(
					'relative w-36 [&[data-state=open]>svg]:rotate-90 ',
					value && 'border-b-background-primary-alt-1'
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
						<span class="text-background-primary-alt-1 text-sm font-semibold">Progress</span>
						<Button
							variant="ghost"
							class={cn(
								'text-foreground-alt-3 hover:text-foreground-alt-2 p-0 text-sm hover:bg-transparent',
								selected.length === 0 && 'invisible'
							)}
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
{:else}
	<Accordion.Item value="progress" class="bg-background-alt-1 overflow-hidden rounded-lg">
		<Accordion.Header>
			<Accordion.Trigger
				class={cn(
					'data-[state=open]:border-b-foreground-alt-4 group flex w-full flex-1 select-none items-center justify-between border-b border-transparent px-2.5 py-5 font-medium transition-transform hover:cursor-pointer',
					value &&
						'data-[state=open]:border-b-background-primary-alt-1 data-[state=closed]:border-b-background-primary-alt-1 data-[state=closed]:border-b-2'
				)}
			>
				<div class="flex items-center gap-1.5">
					<LoaderCircleIcon class="size-6 stroke-2" />
					<span class="w-full text-left">Progress</span>
				</div>

				<div class="flex flex-row items-center gap-3">
					<Button
						variant="ghost"
						class={cn(
							'text-foreground-alt-3 hover:text-foreground-alt-2 h-auto p-0 text-sm hover:bg-transparent',
							selected.length === 0 && 'invisible'
						)}
						onclick={(e: MouseEvent) => {
							e.preventDefault();
							e.stopPropagation();
							selected = [];
							value = '';
							onApply();
						}}
					>
						clear
					</Button>

					<RightChevronIcon
						class="size-4.5 stroke-2 transition-transform duration-100 group-data-[state=open]:rotate-90"
					/>
				</div>
			</Accordion.Trigger>
		</Accordion.Header>

		<Accordion.Content
			class="data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down max-h-72 overflow-hidden overflow-y-scroll px-2.5 py-3 text-sm tracking-[-0.01em]"
		>
			<Checkbox.Group
				class="flex flex-col"
				bind:value={selected}
				name="tags"
				onValueChange={() => {
					onApply();
				}}
			>
				{#each progressStates as progress}
					{@const id = useId()}
					<div
						class="hover:bg-background-alt-3 flex flex-row items-center overflow-hidden rounded-md hover:cursor-pointer"
					>
						<Checkbox.Root
							{id}
							aria-labelledby="{id}-label"
							class="inline-flex size-3.5 h-full shrink-0 items-center justify-center py-1.5 hover:cursor-pointer"
							name={progress.value}
							value={progress.value}
						>
							{#snippet children({ checked })}
								<div class="inline-flex pl-2.5">
									{#if checked}
										<TickIcon class="size-3.5 stroke-2" />
									{:else}
										<span class="size-3.5"></span>
									{/if}
								</div>
							{/snippet}
						</Checkbox.Root>

						<Label.Root
							id="{id}-label"
							for={id}
							class="inline-flex w-full select-none py-1 pl-3.5 pr-1.5 text-sm hover:cursor-pointer"
						>
							{progress.label}
						</Label.Root>
					</div>
				{/each}
			</Checkbox.Group>
		</Accordion.Content>
	</Accordion.Item>
{/if}
