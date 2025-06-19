<script lang="ts">
	import type { AttachmentsModel } from '$lib/models/attachment-model';
	import { Button, Dropdown } from '.';
	import { DownloadIcon, RightChevronIcon } from '../icons';

	type Props = {
		attachments: AttachmentsModel;
		courseId: string;
		assetId: string;
	};

	let { attachments, courseId, assetId }: Props = $props();
</script>

<Dropdown.Root>
	<Dropdown.Trigger
		class="group text-foreground-alt-3 data-[state=open]:text-foreground-alt-1 hover:text-foreground-alt-1 h-auto w-auto rounded-lg border-none p-0"
		onclick={(e) => {
			e.stopPropagation();
		}}
	>
		<div class="flex flex-row items-center gap-1.5">
			{attachments.length + ' attachment' + (attachments.length > 1 ? 's' : '')}

			<RightChevronIcon class="size-3 stroke-2 duration-200 group-data-[state=open]:rotate-90" />
		</div>
	</Dropdown.Trigger>

	<Dropdown.Content
		class="text-foreground-alt-3 z-10 flex max-h-[10rem] w-auto max-w-xs overflow-y-scroll px-1.5 py-1"
		align="start"
	>
		{#each attachments as attachment, index}
			{@const lastAttachment = attachments.length - 1 == index}

			<Dropdown.Item>
				<Button
					href={`/api/courses/${courseId}/assets/${assetId}/attachments/${attachment.id}/serve`}
					variant="ghost"
					class="hover:text-foreground h-auto justify-between gap-5 text-xs hover:bg-transparent"
					download
				>
					<div class="flex flex-row items-center gap-3">
						<span class="shrink-0">{index + 1}.</span>
						<span>{attachment.title}</span>
					</div>

					<DownloadIcon class="size-4 shrink-0" />
				</Button>
			</Dropdown.Item>

			{#if !lastAttachment}
				<Dropdown.Separator />
			{/if}
		{/each}
	</Dropdown.Content>
</Dropdown.Root>
