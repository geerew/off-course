<script lang="ts">
	import type { AttachmentsModel } from '$lib/models/attachment-model';
	import { Button, Dropdown } from '.';
	import { DownloadIcon, RightChevronIcon } from '../icons';

	type Props = {
		attachments: AttachmentsModel;
		courseId: string;
		lessonId: string;
	};

	let { attachments, courseId, lessonId }: Props = $props();
</script>

<Dropdown.Root>
	<Dropdown.Trigger
		class="text-foreground-alt-3 data-[state=open]:text-foreground-alt-1 hover:text-foreground-alt-1 group h-auto w-auto rounded-lg border-none p-0"
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
		class="text-foreground-alt-3 z-10 flex max-h-40 w-auto min-w-44 max-w-60 overflow-y-scroll px-1 py-2"
		align="start"
		portalProps={{ disabled: true }}
	>
		{#each attachments as attachment, index}
			{@const lastAttachment = attachments.length - 1 == index}

			<Dropdown.Item>
				<Button
					href={`/api/courses/${courseId}/lessons/${lessonId}/attachments/${attachment.id}/serve`}
					variant="ghost"
					class="hover:text-foreground h-auto w-full justify-between gap-5 px-1.5 text-xs hover:bg-transparent"
					download
				>
					<span class="grid w-full grid-cols-[auto_1fr_auto] items-start gap-1.5">
						<span class="shrink-0">{index + 1}.</span>
						<span class="wrap-break-word min-w-0 whitespace-normal text-left">
							{attachment.title}
						</span>

						<DownloadIcon class="size-4 shrink-0 self-start justify-self-end" />
					</span>
				</Button>
			</Dropdown.Item>

			{#if !lastAttachment}
				<Dropdown.Separator />
			{/if}
		{/each}
	</Dropdown.Content>
</Dropdown.Root>
