<script lang="ts">
	import type { HTMLAttributes } from "svelte/elements";
	import type { Snippet } from "svelte";
	import { cn } from "$lib/utils.js";

	type Variant = "default" | "secondary" | "destructive" | "outline" | "success" | "warning";

	interface Props extends HTMLAttributes<HTMLDivElement> {
		variant?: Variant;
		class?: string;
		children?: Snippet;
	}

	let { variant = "default", class: className, children, ...restProps }: Props = $props();

	const variants: Record<Variant, string> = {
		default: "border-transparent bg-primary text-primary-foreground shadow",
		secondary: "border-transparent bg-secondary text-secondary-foreground",
		destructive: "border-transparent bg-destructive text-destructive-foreground shadow",
		outline: "border border-input bg-background text-foreground",
		success: "border-transparent bg-green-500 text-white shadow",
		warning: "border-transparent bg-yellow-500 text-white shadow"
	};
</script>

<div
	class={cn(
		"inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
		variants[variant],
		className
	)}
	{...restProps}
>
	{#if children}{@render children()}{/if}
</div>
