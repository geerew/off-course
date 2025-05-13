<script lang="ts">
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import { cn } from '$lib/utils';
	import { Button, DropdownMenu, Separator } from 'bits-ui';
	import { Logo } from '.';
	import { LockIcon, LogoutIcon, RightChevronIcon, UserIcon } from './icons';
	import { Dropdown } from './ui';

	const menu = [
		{
			label: 'Courses',
			href: '/courses',
			matcher: '/courses/'
		}
	];

	function logout() {
		auth.logout();
	}
</script>

<header class="border-background-alt-3 bg-background fixed top-0 z-1 w-full border-b">
	<div class="container-px h-header flex items-center justify-between py-6">
		<!-- Logo -->
		<div class="flex flex-1">
			<a href="/" class="-m-1.5 p-1.5">
				<Logo size="small" />
			</a>
		</div>

		<!-- Menu -->
		<nav class="flex gap-x-12">
			{#each menu as item}
				<Button.Root
					href={item.href}
					class={cn(
						'text-foreground-alt-1 hover:text-foreground relative rounded-lg px-2.5 py-1.5 leading-6 font-semibold duration-200',
						page.url.pathname === item.matcher &&
							'after:bg-background-primary after:absolute after:-bottom-0.5 after:left-0 after:h-0.5 after:w-full'
					)}
					aria-current={page.url.pathname === item.matcher}
				>
					{item.label}
				</Button.Root>
			{/each}
		</nav>

		{#if auth.user !== null}
			<div class="flex flex-1 justify-end">
				<Dropdown
					triggerClass="bg-background-primary-alt-1 hover:bg-background-primary data-[state=open]:bg-background-primary text-foreground-alt-5 size-10 items-center justify-center rounded-full border-none font-semibold"
					contentClass="w-42 p-1"
				>
					{#snippet trigger()}
						{auth.userLetter}
					{/snippet}

					{#snippet content()}
						<div class="flex flex-col select-none">
							<!-- Name -->
							<div class="flex flex-row items-center gap-3 p-1.5">
								<span
									class="bg-background-primary text-foreground-alt-5 relative flex size-10 shrink-0 items-center justify-center rounded-full font-semibold"
								>
									{auth.userLetter}
								</span>
								<span class="text-base font-semibold tracking-wide">
									{auth.user?.displayName}
								</span>
							</div>

							<Separator.Root class="bg-background-alt-3 mb-2 h-px w-full shrink-0" />

							<div class="flex flex-col gap-2">
								<!-- Profile link -->
								<DropdownMenu.Item>
									<Button.Root
										href="/profile"
										class="hover:bg-background-alt-3 hover:text-foreground flex cursor-pointer flex-row items-center justify-between rounded-lg p-1.5 duration-200"
									>
										<div class="flex flex-row items-center gap-3">
											<UserIcon class="size-5 stroke-[1.5]" />
											<span>Profile</span>
										</div>

										<RightChevronIcon class="size-4" />
									</Button.Root>
								</DropdownMenu.Item>

								<!-- Admin link -->
								{#if auth.user?.role === 'admin'}
									<DropdownMenu.Item>
										<Button.Root
											href="/admin"
											class="hover:bg-background-alt-3 hover:text-foreground flex cursor-pointer flex-row items-center justify-between rounded-lg p-1.5 duration-200"
										>
											<div class="flex flex-row items-center gap-3">
												<LockIcon class="size-5 stroke-[1.5]" />
												<span>Admin</span>
											</div>

											<RightChevronIcon class="size-4" />
										</Button.Root>
									</DropdownMenu.Item>
								{/if}

								<!-- Logout link-->
								<DropdownMenu.Item>
									<Button.Root
										onclick={logout}
										class="hover:bg-background-error hover:text-foreground flex w-full cursor-pointer flex-row items-center justify-between rounded-lg p-1.5 duration-200"
									>
										<div class="flex flex-row items-center gap-3">
											<LogoutIcon class="size-5 stroke-[1.5]" />
											<span>Logout</span>
										</div>

										<RightChevronIcon class="size-4" />
									</Button.Root>
								</DropdownMenu.Item>
							</div>
						</div>
					{/snippet}
				</Dropdown>
			</div>
		{/if}
	</div>
</header>
