<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';

	let email = '';
	let password = '';
	let loading = false;
	let error: string | null = null;

	async function handleSubmit() {
		loading = true;
		error = null;

		const result = await auth.login(email, password);
		
		if (result.success) {
			goto('/');
		} else {
			error = result.message || 'Invalid credentials';
		}
		
		loading = false;
	}
</script>

<svelte:head>
	<title>Login - HeadlessForms</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-background p-4">
	<div class="w-full max-w-md">
		<!-- Logo/Brand -->
		<div class="text-center mb-8">
			<div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-violet-500 to-purple-600 shadow-lg mb-4">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-white" viewBox="0 0 20 20" fill="currentColor">
					<path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
					<path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clip-rule="evenodd" />
				</svg>
			</div>
			<h1 class="text-2xl font-bold">HeadlessForms</h1>
			<p class="text-muted-foreground">Sign in to your account</p>
		</div>

		<Card class="shadow-xl">
			<CardHeader class="text-center pb-2">
				<CardTitle class="text-xl">Welcome back</CardTitle>
				<CardDescription>Enter your credentials to continue</CardDescription>
			</CardHeader>
			<CardContent class="pt-4">
				{#if error}
					<div class="bg-destructive/10 text-destructive p-3 rounded-lg mb-4 text-sm flex items-center gap-2">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 shrink-0" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
						</svg>
						{error}
					</div>
				{/if}

				<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-4">
					<div class="space-y-2">
						<label for="email" class="text-sm font-medium">Email</label>
						<Input
							id="email"
							type="email"
							bind:value={email}
							placeholder="you@example.com"
							required
							autocomplete="email"
						/>
					</div>

					<div class="space-y-2">
						<label for="password" class="text-sm font-medium">Password</label>
						<Input
							id="password"
							type="password"
							bind:value={password}
							placeholder="••••••••"
							required
							autocomplete="current-password"
						/>
					</div>

					<Button type="submit" class="w-full" disabled={loading || !email || !password}>
						{#if loading}
							<svg class="animate-spin -ml-1 mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
						{/if}
						Sign In
					</Button>
				</form>

				<div class="mt-4 text-center">
					<a href="/forgot-password" class="text-sm text-muted-foreground hover:text-primary">
						Forgot your password?
					</a>
				</div>

				<div class="mt-6 text-center text-sm text-muted-foreground">
					Don't have an account?
					<a href="/register" class="text-primary hover:underline font-medium">Create one</a>
				</div>
			</CardContent>
		</Card>
	</div>
</div>
