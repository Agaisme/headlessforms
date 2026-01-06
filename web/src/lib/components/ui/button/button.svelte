<script lang="ts">
	import type { HTMLButtonAttributes } from "svelte/elements";
	import type { Snippet } from "svelte";
	import { cn } from "$lib/utils.js";

	type Variant = "default" | "destructive" | "outline" | "secondary" | "ghost" | "link" | "success";
	type Size = "default" | "sm" | "lg" | "icon";

	interface Props extends HTMLButtonAttributes {
		variant?: Variant;
		size?: Size;
		class?: string;
		href?: string;
		loading?: boolean;
		children?: Snippet;
	}

	let { 
		variant = "default", 
		size = "default", 
		class: className,
		href,
		loading = false,
		children,
		disabled,
		...restProps 
	}: Props = $props();

	const baseStyles = "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0";

	const variants: Record<Variant, string> = {
		default: "bg-primary text-primary-foreground shadow hover:bg-primary/90",
		destructive: "bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90",
		outline: "border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground",
		secondary: "bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80",
		ghost: "hover:bg-accent hover:text-accent-foreground",
		link: "text-primary underline-offset-4 hover:underline",
		success: "bg-green-600 text-white shadow hover:bg-green-700"
	};

	const sizes: Record<Size, string> = {
		default: "h-9 px-4 py-2",
		sm: "h-8 rounded-md px-3 text-xs",
		lg: "h-10 rounded-md px-8",
		icon: "h-9 w-9"
	};
</script>

{#snippet loadingSpinner()}
	<svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
		<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
		<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
	</svg>
{/snippet}

{#if href && !loading}
	<a
		{href}
		class={cn(baseStyles, variants[variant], sizes[size], className)}
		{...restProps}
	>
		{#if children}{@render children()}{/if}
	</a>
{:else}
	<button
		class={cn(baseStyles, variants[variant], sizes[size], className)}
		disabled={disabled || loading}
		{...restProps}
	>
		{#if loading}
			{@render loadingSpinner()}
		{/if}
		{#if children}{@render children()}{/if}
	</button>
{/if}
