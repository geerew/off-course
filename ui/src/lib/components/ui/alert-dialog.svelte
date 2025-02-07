<script lang="ts">
	import { cn } from '$lib/utils';
	import { AlertDialog, type WithoutChild } from 'bits-ui';
	import type { Snippet } from 'svelte';

	type Props = AlertDialog.RootProps & {
		trigger: Snippet;
		triggerClass?: string;
		contentProps?: Omit<WithoutChild<AlertDialog.ContentProps>, 'class'>;
		contentClass?: string;
		description: Snippet;
		action: Snippet;
	};

	let {
		open = $bindable(false),
		children,
		trigger,
		triggerClass,
		contentProps,
		contentClass,
		description,
		action,
		...restProps
	}: Props = $props();
</script>

<AlertDialog.Root bind:open {...restProps}>
	<AlertDialog.Trigger
		class={cn(
			'bg-background-error hover:bg-background-error-alt-1 text-foreground-alt-1 hover:text-foreground w-36 cursor-pointer rounded-md py-2 duration-200 select-none',
			triggerClass
		)}
	>
		{@render trigger()}
	</AlertDialog.Trigger>

	<AlertDialog.Portal>
		<AlertDialog.Overlay
			class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/60"
		/>

		<AlertDialog.Content
			class={cn(
				'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 data-[state=open]:slide-in-from-top-5 data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:slide-out-to-top-5 fixed top-20 left-1/2 z-50 w-full max-w-lg min-w-[20rem] -translate-x-1/2 overflow-hidden px-10 data-[state=closed]:duration-200 data-[state=open]:duration-200',
				contentClass
			)}
			{...contentProps}
		>
			<div class="bg-background-alt-1 overflow-hidden rounded-lg">
				<div class="flex flex-col gap-2.5 p-5">
					<div class="flex items-center justify-center">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke-width="1.5"
							stroke="currentColor"
							class="text-foreground-error size-14"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
							/>
						</svg>
					</div>

					{@render description()}
				</div>

				<div
					class="bg-background-alt-2 border-background-alt-3 flex w-full items-center justify-end gap-2 border-t px-5 py-2.5"
				>
					<AlertDialog.Cancel
						type="button"
						class="border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground w-24 cursor-pointer rounded-md border py-2 duration-200 select-none"
					>
						Cancel
					</AlertDialog.Cancel>

					{@render action()}
				</div>
			</div>
		</AlertDialog.Content>
	</AlertDialog.Portal>
</AlertDialog.Root>
