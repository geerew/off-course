<script lang="ts">
	import type { AttachmentsModel } from '$lib/models/attachment-model';
	import { DropdownMenu } from 'bits-ui';
	import { Button, Dropdown } from '.';
	import { DownloadIcon, RightChevronIcon } from '../icons';

	type Props = {
		attachments: AttachmentsModel;
		courseId: string;
		assetId: string;
	};

	let { attachments, courseId, assetId }: Props = $props();
</script>

<Dropdown
	triggerClass="group text-foreground-alt-3 data-[state=open]:text-foreground-alt-1 hover:text-foreground-alt-1 h-auto rounded-lg border-none"
	contentClass="text-foreground-alt-3 flex max-h-[10rem] w-auto max-w-xs overflow-y-scroll px-1.5 py-1"
	portalProps={{ disabled: false }}
>
	{#snippet trigger()}
		<div class="flex flex-row items-center gap-1.5">
			{attachments.length + ' attachment' + (attachments.length > 1 ? 's' : '')}

			<RightChevronIcon class="size-3 stroke-2 duration-200 group-data-[state=open]:rotate-90" />
		</div>
	{/snippet}

	{#snippet content()}
		{#each attachments as attachment, index}
			{@const lastAttachment = attachments.length - 1 == index}

			<DropdownMenu.Item>
				<Button
					href={`/api/courses/${courseId}/assets/${assetId}/attachments/${attachment.id}/serve`}
					download
					class="hover:bg-background-alt-3 hover:text-foreground text-foreground-alt-1 bg-background flex h-auto cursor-pointer flex-row items-center justify-between rounded-md p-1 text-xs duration-200"
				>
					<div class="flex flex-row items-center gap-3">
						<span class="shrink-0">{index + 1}.</span>
						<span>{attachment.title}</span>
					</div>

					<DownloadIcon class="size-4 shrink-0" />
				</Button>
			</DropdownMenu.Item>

			{#if !lastAttachment}
				<DropdownMenu.Separator class="bg-background-alt-1 flex h-px w-full shrink-0" />
			{/if}
		{/each}
	{/snippet}
</Dropdown>
