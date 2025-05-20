<script lang="ts">
	import { page } from '$app/state';
	import { BurgerMenuIcon, CourseIcon, TagIcon, UserIcon } from '$lib/components/icons';
	import { Button } from '$lib/components/ui';
	import { cn, remCalc } from '$lib/utils';
	import { Dialog } from 'bits-ui';
	import { innerWidth } from 'svelte/reactivity/window';
	import theme from 'tailwindcss/defaultTheme';

	let { children } = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let menuPopupMode = $state(false);
	let dialogOpen = $state(false);

	let windowWidth = $derived(remCalc(innerWidth.current ?? 0));

	const menu = [
		{
			label: 'Courses',
			href: '/admin/courses',
			matcher: '/admin/courses/',
			icon: CourseIcon
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
		}
	];

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
			class={cn(
				'bg-background text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-1 relative flex h-auto flex-row justify-start gap-3 px-2.5 leading-6 font-semibold duration-200',
				page.url.pathname.startsWith(item.matcher) &&
					'bg-background-alt-1 after:bg-background-primary after:absolute after:top-0 after:right-0 after:h-full after:w-1',
				mobile ? ' py-6' : ' py-3'
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
		</Button>
	{/each}
{/snippet}

<div
	class={cn(
		'grid grid-rows-1 gap-6 pt-[calc(var(--header-height)+1))]',
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
					class="border-foreground-alt-4 bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left fixed top-0 left-0 z-50 h-full w-[var(--settings-menu-width)] border-r pt-4 pl-4"
				>
					<nav class="flex h-full w-full flex-col gap-3 overflow-x-hidden overflow-y-auto pb-8">
						{@render menuContents(true)}
					</nav>
				</Dialog.Content>
			</Dialog.Portal>
		</Dialog.Root>
	{:else}
		<div class="relative row-span-full">
			<div class="absolute inset-0">
				<nav
					class="container-pl border-foreground-alt-5 sticky top-[calc(var(--header-height)+1px)] left-0 flex h-[calc(100dvh-(var(--header-height)+1px))] w-[var(--settings-menu-width)] flex-col gap-4 border-r py-8"
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
				class="bg-background text-foreground-alt-2 hover:text-foreground-alt-1 flex h-auto items-start justify-start gap-1.5 text-start duration-200 enabled:hover:bg-transparent"
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
