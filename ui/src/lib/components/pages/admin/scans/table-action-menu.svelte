<script lang="ts">
	import { DeleteScanDialog } from '$lib/components/dialogs';
	import { ActionIcon, DeleteIcon, DeselectAllIcon, RightChevronIcon } from '$lib/components/icons';
	import { Dropdown } from '$lib/components/ui';
	import type { ScanModel } from '$lib/models/scan-model';

	type Props = {
		scans: Record<string, ScanModel>;
		onDelete: () => void;
	};

	let { scans = $bindable(), onDelete }: Props = $props();

	let deleteDialogOpen = $state(false);
</script>

<Dropdown.Root>
	<Dropdown.Trigger
		class="w-32 [&[data-state=open]>svg]:rotate-90"
		disabled={Object.keys(scans).length === 0}
	>
		<div class="flex items-center gap-1.5">
			<ActionIcon class="size-4 stroke-[1.5]" />
			<span>Actions</span>
		</div>
		<RightChevronIcon class="stroke-foreground-alt-3 size-4.5 duration-200" />
	</Dropdown.Trigger>

	<Dropdown.Content class="w-42">
		<Dropdown.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				scans = {};
			}}
		>
			<DeselectAllIcon class="size-4 stroke-[1.5]" />
			<span>Deselect all</span>
		</Dropdown.Item>

		<Dropdown.Separator />

		<Dropdown.CautionItem
			onclick={() => {
				deleteDialogOpen = true;
			}}
		>
			<DeleteIcon class="size-4 stroke-[1.5]" />
			<span>Delete</span>
		</Dropdown.CautionItem>
	</Dropdown.Content>
</Dropdown.Root>

<DeleteScanDialog bind:open={deleteDialogOpen} value={Object.values(scans)} successFn={onDelete} />
