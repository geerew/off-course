<script lang="ts">
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import { cn } from '$lib/utils';
	import { Logo } from '.';
	import { LockIcon, LogoutIcon, RightChevronIcon, UserIcon } from './icons';
	import { Button, Dropdown } from './ui';

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

<header class="border-background-alt-3 bg-background fixed top-0 z-50 w-full border-b">
	<div class="container-px flex h-(--header-height) items-center justify-between py-6">
		<!-- Logo -->
		<div class="flex flex-1">
			<a href="/" class="-m-1.5 p-1.5">
				<Logo size="small" />
			</a>
		</div>

		<!-- Menu -->
		<nav class="flex gap-x-12">
			{#each menu as item}
				<Button
					href={item.href}
					variant="ghost"
					class={cn(
						'hover:text-foreground relative px-2.5 py-1.5 text-base leading-6 hover:bg-transparent',
						page.url.pathname === item.matcher &&
							'after:bg-background-primary text-foreground after:absolute after:-bottom-0.5 after:left-0 after:h-0.5 after:w-full'
					)}
					aria-current={page.url.pathname === item.matcher}
				>
					{item.label}
				</Button>
			{/each}
		</nav>

		{#if auth.user !== null}
			<div class="flex flex-1 justify-end">
				<Dropdown.Root>
					<Dropdown.Trigger
						class="bg-background-primary-alt-1 hover:bg-background-primary data-[state=open]:bg-background-primary text-foreground-alt-6 size-10 justify-center rounded-full border-none font-semibold"
					>
						{auth.userLetter}
					</Dropdown.Trigger>

					<Dropdown.Content class="w-42" portalProps={{ disabled: true }}>
						<div class="flex flex-col select-none">
							<!-- Name -->
							<div class="flex flex-row items-center gap-3 p-1.5 pb-2.5">
								<span
									class="bg-background-primary-alt-1 text-foreground-alt-6 relative flex size-10 shrink-0 items-center justify-center rounded-full font-semibold"
								>
									{auth.userLetter}
								</span>

								<span class="text-base font-semibold tracking-wide">
									{auth.user?.displayName}
								</span>
							</div>

							<Dropdown.Separator />

							<div class="flex flex-col gap-2 pt-2">
								<!-- Profile link -->
								<Dropdown.Item>
									<Button
										href="/profile"
										variant="ghost"
										class="hover:text-foreground h-auto w-full justify-between p-1 hover:bg-transparent"
									>
										<div class="flex flex-row items-center gap-2.5">
											<UserIcon class="size-5 stroke-[1.5]" />
											<span>Profile</span>
										</div>

										<RightChevronIcon class="size-4" />
									</Button>
								</Dropdown.Item>

								<!-- Admin link -->
								{#if auth.user?.role === 'admin'}
									<Dropdown.Item>
										<Button
											href="/admin"
											variant="ghost"
											class="hover:text-foreground h-auto w-full justify-between p-1 hover:bg-transparent"
										>
											<div class="flex flex-row items-center gap-2.5">
												<LockIcon class="size-5 stroke-[1.5]" />
												<span>Admin</span>
											</div>

											<RightChevronIcon class="size-4" />
										</Button>
									</Dropdown.Item>
								{/if}

								<!-- Logout link-->
								<Dropdown.CautionItem>
									<Button
										onclick={logout}
										variant="ghost"
										class="hover:text-foreground h-auto w-full justify-between p-1 hover:bg-transparent"
									>
										<div class="flex flex-row items-center gap-2.5">
											<LogoutIcon class="size-5 stroke-[1.5]" />
											<span>Logout</span>
										</div>

										<RightChevronIcon class="size-4" />
									</Button>
								</Dropdown.CautionItem>
							</div>
						</div>
					</Dropdown.Content>
				</Dropdown.Root>
			</div>
		{/if}
	</div>
</header>
