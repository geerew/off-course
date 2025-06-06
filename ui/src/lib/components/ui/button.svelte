<script lang="ts" module>
	import { type VariantProps, tv } from 'tailwind-variants';

	export const buttonVariants = tv({
		base: "inline-flex shrink-0 items-center cursor-pointer justify-center gap-2 select-none whitespace-nowrap rounded-md text-sm font-medium outline-none transition-all duration-200 disabled:pointer-events-none disabled:opacity-50 px-2 h-10 aria-disabled:pointer-events-none aria-disabled:opacity-50 [&_svg:not([class*='size-'])]:size-4 [&_svg]:pointer-events-none [&_svg]:shrink-0",
		variants: {
			variant: {
				default: 'bg-background-primary-alt-1 text-foreground-alt-6 hover:bg-background-primary',
				secondary:
					'bg-background-alt-4 hover:bg-background-alt-5 text-foreground-alt-1 hover:text-foreground',
				destructive:
					'bg-background-error hover:bg-background-error-alt-1 text-foreground-alt-1 hover:text-foreground',
				outline:
					'text-foreground-alt-1 border-background-alt-4 hover:bg-background-alt-4 hover:text-foreground border bg-transparent',
				ghost: 'text-foreground-alt-1 hover:bg-background bg-transparent'
			}
		},
		defaultVariants: {
			variant: 'default'
		}
	});

	export type ButtonVariant = VariantProps<typeof buttonVariants>['variant'];
</script>

<script lang="ts">
	import { cn } from '$lib/utils';
	import { Button } from 'bits-ui';

	type Props = Button.RootProps & {
		containerClass?: string;
		variant?: ButtonVariant;
	};

	let { class: className = '', variant = 'default', children, ...restProps }: Props = $props();
</script>

<Button.Root class={cn(buttonVariants({ variant }), className)} {...restProps}>
	{@render children?.()}
</Button.Root>
