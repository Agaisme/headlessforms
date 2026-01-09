<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { toast } from '$lib/stores/toast';

	let email = '';
	let loading = false;
	let submitted = false;

	async function handleSubmit() {
		if (!email) {
			toast.error('Please enter your email address');
			return;
		}

		loading = true;
		try {
			const res = await fetch('/api/v1/auth/forgot-password', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email })
			});

			const json = await res.json();
			if (json.status === 'success') {
				submitted = true;
			} else {
				toast.error(json.message || 'Failed to send reset email');
			}
		} catch (e) {
			toast.error('Failed to send reset email');
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center bg-background p-4">
	<Card class="w-full max-w-md">
		<CardHeader class="text-center pb-4">
			<div class="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center mx-auto mb-4">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-primary" viewBox="0 0 20 20" fill="currentColor">
					<path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
				</svg>
			</div>
			<h1 class="text-2xl font-bold">Forgot Password</h1>
			<p class="text-muted-foreground text-sm mt-1">Enter your email to receive a reset link</p>
		</CardHeader>
		<CardContent>
			{#if submitted}
				<div class="text-center py-6">
					<div class="w-16 h-16 rounded-full bg-emerald-500/10 flex items-center justify-center mx-auto mb-4">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-emerald-500" viewBox="0 0 20 20" fill="currentColor">
							<path d="M2.003 5.884L10 9.882l7.997-3.998A2 2 0 0016 4H4a2 2 0 00-1.997 1.884z" />
							<path d="M18 8.118l-8 4-8-4V14a2 2 0 002 2h12a2 2 0 002-2V8.118z" />
						</svg>
					</div>
					<h2 class="text-lg font-semibold mb-2">Check your email</h2>
					<p class="text-muted-foreground text-sm mb-6">
						If an account exists for {email}, you'll receive a password reset link shortly.
					</p>
					<Button variant="outline" href="/login">Back to Login</Button>
				</div>
			{:else}
				<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-4">
				<div>
						<label for="email" class="text-sm font-medium mb-1.5 block">Email Address</label>
						<Input id="email" type="email" bind:value={email} placeholder="you@example.com" required />
					</div>
					<Button type="submit" class="w-full" disabled={loading}>
						{loading ? 'Sending...' : 'Send Reset Link'}
					</Button>
				</form>
				<div class="mt-4 text-center text-sm text-muted-foreground">
					Remember your password? <a href="/login" class="text-primary hover:underline">Sign in</a>
				</div>
			{/if}
		</CardContent>
	</Card>
</div>
