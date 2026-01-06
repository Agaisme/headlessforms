<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { toast } from '$lib/stores/toast';

	let password = '';
	let confirmPassword = '';
	let loading = false;
	let success = false;

	$: token = $page.url.searchParams.get('token') || '';

	async function handleSubmit() {
		if (!token) {
			toast.error('Invalid reset link');
			return;
		}

		if (password.length < 8) {
			toast.error('Password must be at least 8 characters');
			return;
		}

		if (password !== confirmPassword) {
			toast.error('Passwords do not match');
			return;
		}

		loading = true;
		try {
			const res = await fetch('/api/v1/auth/reset-password', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ token, new_password: password })
			});

			const json = await res.json();
			if (json.status === 'success') {
				success = true;
				toast.success('Password reset successfully');
			} else {
				toast.error(json.message || 'Failed to reset password');
			}
		} catch (e) {
			toast.error('Failed to reset password');
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
			<h1 class="text-2xl font-bold">Reset Password</h1>
			<p class="text-muted-foreground text-sm mt-1">Enter your new password</p>
		</CardHeader>
		<CardContent>
			{#if !token}
				<div class="text-center py-6">
					<div class="w-16 h-16 rounded-full bg-rose-500/10 flex items-center justify-center mx-auto mb-4">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-rose-500" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
						</svg>
					</div>
					<h2 class="text-lg font-semibold mb-2">Invalid Reset Link</h2>
					<p class="text-muted-foreground text-sm mb-6">
						This password reset link is invalid or has expired.
					</p>
					<Button href="/forgot-password">Request New Link</Button>
				</div>
			{:else if success}
				<div class="text-center py-6">
					<div class="w-16 h-16 rounded-full bg-emerald-500/10 flex items-center justify-center mx-auto mb-4">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-emerald-500" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
						</svg>
					</div>
					<h2 class="text-lg font-semibold mb-2">Password Reset!</h2>
					<p class="text-muted-foreground text-sm mb-6">
						Your password has been successfully changed.
					</p>
					<Button href="/login">Sign In</Button>
				</div>
			{:else}
				<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-4">
					<div>
						<label class="text-sm font-medium mb-1.5 block">New Password</label>
						<Input type="password" bind:value={password} placeholder="Min 8 characters" required />
					</div>
					<div>
						<label class="text-sm font-medium mb-1.5 block">Confirm Password</label>
						<Input type="password" bind:value={confirmPassword} placeholder="Repeat password" required />
					</div>
					<Button type="submit" class="w-full" disabled={loading}>
						{loading ? 'Resetting...' : 'Reset Password'}
					</Button>
				</form>
			{/if}
		</CardContent>
	</Card>
</div>
