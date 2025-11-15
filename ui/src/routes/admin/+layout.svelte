<script lang="ts">
	import { page } from '$app/state';
	import { subscribeToScans, type ScanUpdateEvent } from '$lib/api/scan-api';
	import type { ScanModel } from '$lib/models/scan-model';
	import {
		BurgerMenuIcon,
		CourseIcon,
		LogsIcon,
		ScanIcon,
		TagIcon,
		UserIcon
	} from '$lib/components/icons';
	import { Badge, Button } from '$lib/components/ui';
	import { cn, remCalc } from '$lib/utils';
	import { Dialog } from 'bits-ui';
	import { innerWidth } from 'svelte/reactivity/window';
	import theme from 'tailwindcss/defaultTheme';

	let { children } = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let menuPopupMode = $state(false);
	let dialogOpen = $state(false);
	let scanCount = $state(0);

	let windowWidth = $derived(remCalc(innerWidth.current ?? 0));

	let sseClose: (() => void) | null = null;
	let seenScanIds = new Set<string>();

	const menu = [
		{
			label: 'Courses',
			href: '/admin/courses',
			matcher: '/admin/courses/',
			icon: CourseIcon
		},
		{
			label: 'Scans',
			href: '/admin/scans',
			matcher: '/admin/scans/',
			icon: ScanIcon
		},
		{
			label: 'Tags',
			href: '/admin/tags',
			matcher: '/admin/tags/',
			icon: TagIcon
		},
		{
			label: 'Users',
			href: '/admin/users',
			matcher: '/admin/users/',
			icon: UserIcon
		},
		{
			label: 'Logs',
			href: '/admin/logs',
			matcher: '/admin/logs/',
			icon: LogsIcon
		}
	];

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		// Subscribe to scan updates
		sseClose = subscribeToScans({
			onUpdate: (event: ScanUpdateEvent) => {
				if (event.type === 'all_scans') {
					// Initial load - set count and populate seen scan IDs
					const scans = event.data as ScanModel[];
					scanCount = scans.length;
					seenScanIds = new Set(scans.map((scan) => scan.id));
				} else if (event.type === 'scan_update') {
					const scan = event.data as ScanModel;
					// Only increment if this is a new scan we haven't seen before
					if (!seenScanIds.has(scan.id)) {
						seenScanIds.add(scan.id);
						scanCount++;
					}
					// If we've already seen it, it's just an update - no count change needed
				} else if (event.type === 'scan_deleted') {
					const deletedId = (event.data as { id: string }).id;
					// Only decrement if we were tracking this scan
					if (seenScanIds.has(deletedId)) {
						seenScanIds.delete(deletedId);
						scanCount = Math.max(0, scanCount - 1);
					}
				}
			},
			onError: () => {
				// On error, reset count - will be repopulated on reconnect
				scanCount = 0;
				seenScanIds.clear();
			}
		});

		// Cleanup on unmount
		return () => {
			if (sseClose) {
				sseClose();
				sseClose = null;
			}
		};
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the menu popup mode based on the screen size
	$effect(() => {
		menuPopupMode = windowWidth >= +theme.screens.lg.replace('rem', '') ? false : true;
	});
</script>

{#snippet menuContents(mobile: boolean)}
	{#each menu as item}
		<Button
			href={item.href}
			variant="ghost"
			class={cn(
				'text-foreground-alt-2 hover:text-foreground hover:bg-background-alt-1 relative h-auto justify-start gap-3 px-2.5 leading-6',
				page.url.pathname.startsWith(item.matcher) &&
					'bg-background-alt-1 after:bg-background-primary after:absolute after:right-0 after:top-0 after:h-full after:w-1',
				mobile ? 'py-6 text-base' : 'py-3'
			)}
			onclick={() => {
				if (mobile && menuPopupMode) {
					dialogOpen = false;
				}
			}}
			aria-current={page.url.pathname === item.matcher}
		>
			<item.icon class="size-6 stroke-[1.5]" />
			<span>{item.label}</span>
			{#if item.label === 'Scans' && scanCount > 0}
				<Badge class="bg-background-alt-4 text-foreground ml-auto mr-2.5 text-xs">
					{scanCount}
				</Badge>
			{/if}
		</Button>
	{/each}
{/snippet}

<div
	class={cn(
		'grid grid-rows-1 gap-6 pt-[calc(var(--header-height)+1)]',
		menuPopupMode ? 'grid-cols-1' : 'grid-cols-[var(--settings-menu-width)_1fr]'
	)}
>
	{#if menuPopupMode}
		<Dialog.Root bind:open={dialogOpen}>
			<Dialog.Portal>
				<Dialog.Overlay
					class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-40 bg-black/30"
				/>

				<Dialog.Content
					class="border-foreground-alt-4 bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left fixed left-0 top-0 z-50 h-full w-[--settings-menu-width] border-r pl-4 pt-4"
				>
					<nav class="flex h-full w-full flex-col gap-3 overflow-y-auto overflow-x-hidden pb-8">
						{@render menuContents(true)}
					</nav>
				</Dialog.Content>
			</Dialog.Portal>
		</Dialog.Root>
	{:else}
		<div class="relative row-span-full">
			<div class="absolute inset-0">
				<nav
					class="container-pl border-foreground-alt-5 sticky left-0 top-[calc(var(--header-height)+1px)] flex h-[calc(100dvh-(var(--header-height)+1px))] w-[--settings-menu-width] flex-col gap-4 border-r py-8"
				>
					{@render menuContents(false)}
				</nav>
			</div>
		</div>
	{/if}

	<!-- Popup trigger -->
	<div
		class={cn('border-background-alt-3 flex h-12 border-b', menuPopupMode ? 'visible' : 'hidden')}
	>
		<div class="container-pl flex h-full items-center">
			<Button
				variant="ghost"
				class="text-foreground-alt-2 hover:text-foreground h-auto hover:bg-transparent"
				onclick={() => {
					dialogOpen = !dialogOpen;
				}}
			>
				<BurgerMenuIcon class="size-6 stroke-[1.5]" />
				<span>Menu</span>
			</Button>
		</div>
	</div>

	<main class={cn('container-px flex w-full pb-8', !menuPopupMode && 'pt-8')}>
		{@render children()}
	</main>
</div>
