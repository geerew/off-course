<script lang="ts">
	import { DeleteTagDialog, EditTagNameDialog } from '$lib/components/dialogs';
	import { DeleteIcon, DotsIcon, EditIcon } from '$lib/components/icons';
	import { Dropdown } from '$lib/components/ui';
	import type { TagModel } from '$lib/models/tag-model';

	type Props = {
		tag: TagModel;
		onDelete: () => void;
	};

	let { tag = $bindable(), onDelete }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let editDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);
</script>

<Dropdown.Root>
	<Dropdown.Trigger
		class="hover:bg-background-alt-3 data-[state=open]:bg-background-alt-3 w-auto rounded-lg border-none"
	>
		<DotsIcon class="size-5 stroke-[1.5]" />
	</Dropdown.Trigger>

	<Dropdown.Content class="w-38">
		<Dropdown.Item
			onclick={() => {
				editDialogOpen = true;
			}}
		>
			<EditIcon class="size-4 stroke-[1.5]" />
			<span>Rename</span>
		</Dropdown.Item>

		<Dropdown.Separator />

		<Dropdown.CautionItem
			onclick={() => {
				deleteDialogOpen = true;
			}}
		>
			<DeleteIcon class="size-4 stroke-[1.5]" />
			<span>Delete tag</span>
		</Dropdown.CautionItem>
	</Dropdown.Content>
</Dropdown.Root>

<EditTagNameDialog bind:open={editDialogOpen} bind:value={tag} />
<DeleteTagDialog bind:open={deleteDialogOpen} value={tag} successFn={onDelete} />
