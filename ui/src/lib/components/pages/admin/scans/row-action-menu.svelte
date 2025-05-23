<script lang="ts">
	import { DeleteScanDialog } from '$lib/components/dialogs';
	import { DeleteIcon, DotsIcon } from '$lib/components/icons';
	import { Dropdown } from '$lib/components/ui';
	import type { ScanModel } from '$lib/models/scan-model';

	type Props = {
		scan: ScanModel;
		onDelete: () => void;
	};

	let { scan, onDelete }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let deleteDialogOpen = $state(false);
</script>

<Dropdown.Root>
	<Dropdown.Trigger
		class="hover:bg-background-alt-3 data-[state=open]:bg-background-alt-3 w-auto rounded-lg border-none"
	>
		<DotsIcon class="size-5 stroke-[1.5]" />
	</Dropdown.Trigger>

	<Dropdown.Content class="w-38">
		<Dropdown.CautionItem
			onclick={() => {
				deleteDialogOpen = true;
			}}
		>
			<DeleteIcon class="size-4 stroke-[1.5]" />
			<span>Delete Scan</span>
		</Dropdown.CautionItem>
	</Dropdown.Content>
</Dropdown.Root>

<DeleteScanDialog bind:open={deleteDialogOpen} value={scan} successFn={onDelete} />
