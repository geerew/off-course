<script lang="ts">
	import { page } from '$app/state';
	import { CourseIcon, TagIcon, UserIcon } from '$lib/components/icons';
	import { cn } from '$lib/utils';
	import { Button } from 'bits-ui';

	const menu = [
		{
			label: 'Users',
			href: '/admin/users',
			matcher: '/admin/users/',
			icon: UserIcon
		},
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
		}
	];

	let { children } = $props();
</script>

<div
	class="grid grid-cols-[var(--settings-menu-width)_1fr] grid-rows-1 gap-6 pt-[calc(var(--height-header)+1))]"
>
	<div class="relative row-span-full">
		<div class="absolute inset-0">
			<nav
				class="container-pl border-foreground-alt-4 sticky top-[calc(var(--height-header)+1px)] left-0 flex h-[calc(100dvh-(var(--height-header)+1px))] w-[--settings-menu-width] flex-col gap-4 border-r py-8"
			>
				{#each menu as item}
					<Button.Root
						href={item.href}
						class={cn(
							'text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-1 relative flex flex-row gap-3 px-2.5 py-3 leading-6 font-semibold duration-200',
							page.url.pathname.startsWith(item.matcher) &&
								'bg-background-alt-1 after:bg-background-primary after:absolute after:top-0 after:right-0 after:h-full after:w-1'
						)}
						aria-current={page.url.pathname === item.matcher}
					>
						<item.icon class="size-6 stroke-[1.5]" />
						<span>{item.label}</span>
					</Button.Root>
				{/each}
			</nav>
		</div>
	</div>

	<main class="container-pr flex w-full py-8">
		{@render children()}
	</main>
</div>
