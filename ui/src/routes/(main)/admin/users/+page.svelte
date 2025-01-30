<script lang="ts">
	import { Plus } from '$lib/components/icons';
	import { Button } from 'bits-ui';
</script>

<div class="flex flex-col gap-6 pt-1">
	<Button.Root
		type="button"
		class="bg-background-alt-4 hover:bg-background-alt-5 flex h-auto flex-row items-center gap-2 rounded-md px-5 py-2.5 hover:cursor-pointer"
		aria-label="Toggle password visibility"
	>
		<Plus class="size-5 stroke-[1.5]" />
		Add User</Button.Root
	>
	<div>Table</div>
</div>
