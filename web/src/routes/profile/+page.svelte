<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { auth } from '$lib/stores/auth';
	import { toast } from '$lib/stores/toast';

	let loading = false;
	let savingProfile = false;
	let savingPassword = false;

	// Profile form
	let profileName = '';
	let profileEmail = '';

	// Password form
	let currentPassword = '';
	let newPassword = '';
	let confirmPassword = '';

	onMount(() => {
		if ($auth.user) {
			profileName = $auth.user.name || '';
			profileEmail = $auth.user.email || '';
		}
	});

	// Keep profile form in sync with auth store
	$: if ($auth.user) {
		profileName = $auth.user.name || '';
		profileEmail = $auth.user.email || '';
	}

	async function handleUpdateProfile() {
		if (!profileName.trim() && !profileEmail.trim()) {
			toast.error('Please enter at least one field to update');
			return;
		}

		savingProfile = true;
		try {
			const token = auth.getToken();
			const res = await fetch('/api/v1/auth/profile', {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					...(token ? { 'Authorization': `Bearer ${token}` } : {})
				},
				body: JSON.stringify({
					name: profileName,
					email: profileEmail
				})
			});

			const json = await res.json();

			if (json.status === 'success') {
				toast.success('Profile updated successfully');
				// Update local storage with new user data
				if (typeof window !== 'undefined') {
					localStorage.setItem('auth_user', JSON.stringify(json.data));
				}
			} else {
				toast.error(json.message || 'Failed to update profile');
			}
		} catch (e) {
			toast.error('Failed to update profile');
		} finally {
			savingProfile = false;
		}
	}

	async function handleUpdatePassword() {
		if (!currentPassword || !newPassword || !confirmPassword) {
			toast.error('Please fill in all password fields');
			return;
		}

		if (newPassword !== confirmPassword) {
			toast.error('New passwords do not match');
			return;
		}

		if (newPassword.length < 8) {
			toast.error('New password must be at least 8 characters');
			return;
		}

		savingPassword = true;
		try {
			const token = auth.getToken();
			const res = await fetch('/api/v1/auth/password', {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					...(token ? { 'Authorization': `Bearer ${token}` } : {})
				},
				body: JSON.stringify({
					current_password: currentPassword,
					new_password: newPassword
				})
			});

			const json = await res.json();

			if (json.status === 'success') {
				toast.success('Password updated successfully');
				// Clear form
				currentPassword = '';
				newPassword = '';
				confirmPassword = '';
			} else {
				toast.error(json.message || 'Failed to update password');
			}
		} catch (e) {
			toast.error('Failed to update password');
		} finally {
			savingPassword = false;
		}
	}

	function getRoleBadgeVariant(role: string) {
		if (role === 'super_admin') return 'default';
		if (role === 'admin') return 'secondary';
		return 'outline';
	}

	function formatRole(role: string) {
		if (role === 'super_admin') return 'Super Admin';
		if (role === 'admin') return 'Admin';
		return 'User';
	}
</script>

<div class="max-w-2xl mx-auto space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold">Profile Settings</h1>
			<p class="text-muted-foreground">Manage your account settings</p>
		</div>
	</div>

	<!-- Profile Card -->
	<Card>
		<CardContent class="p-6">
			<div class="flex items-center gap-4 mb-6">
				<div class="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center text-primary text-2xl font-bold">
					{$auth.user?.name?.[0]?.toUpperCase() || $auth.user?.email?.[0]?.toUpperCase() || '?'}
				</div>
				<div>
					<h2 class="text-lg font-semibold">{$auth.user?.name || 'No name'}</h2>
					<p class="text-muted-foreground">{$auth.user?.email}</p>
					<Badge variant={getRoleBadgeVariant($auth.user?.role || 'user')} class="mt-1">
						{formatRole($auth.user?.role || 'user')}
					</Badge>
				</div>
			</div>

			<form on:submit|preventDefault={handleUpdateProfile} class="space-y-4">
				<h3 class="font-medium border-b pb-2">Update Profile</h3>
				
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div class="form-control">
						<label class="label" for="profile-name">
							<span class="label-text font-medium">Display Name</span>
						</label>
						<Input
							id="profile-name"
							type="text"
							bind:value={profileName}
							placeholder="Your name"
						/>
					</div>

					<div class="form-control">
						<label class="label" for="profile-email">
							<span class="label-text font-medium">Email Address</span>
						</label>
						<Input
							id="profile-email"
							type="email"
							bind:value={profileEmail}
							placeholder="your@email.com"
						/>
					</div>
				</div>

				<div class="flex justify-end">
					<Button type="submit" disabled={savingProfile} class="gap-2">
						{#if savingProfile}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						Save Profile
					</Button>
				</div>
			</form>
		</CardContent>
	</Card>

	<!-- Password Card -->
	<Card>
		<CardContent class="p-6">
			<form on:submit|preventDefault={handleUpdatePassword} class="space-y-4">
				<h3 class="font-medium border-b pb-2">Change Password</h3>
				
				<div class="space-y-4">
					<div class="form-control">
						<label class="label" for="current-password">
							<span class="label-text font-medium">Current Password</span>
						</label>
						<Input
							id="current-password"
							type="password"
							bind:value={currentPassword}
							placeholder="••••••••"
						/>
					</div>

					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div class="form-control">
							<label class="label" for="new-password">
								<span class="label-text font-medium">New Password</span>
							</label>
							<Input
								id="new-password"
								type="password"
								bind:value={newPassword}
								placeholder="••••••••"
							/>
							<p class="text-xs text-muted-foreground mt-1">Minimum 8 characters</p>
						</div>

						<div class="form-control">
							<label class="label" for="confirm-password">
								<span class="label-text font-medium">Confirm New Password</span>
							</label>
							<Input
								id="confirm-password"
								type="password"
								bind:value={confirmPassword}
								placeholder="••••••••"
							/>
						</div>
					</div>
				</div>

				<div class="flex justify-end">
					<Button type="submit" disabled={savingPassword} variant="outline" class="gap-2">
						{#if savingPassword}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						Update Password
					</Button>
				</div>
			</form>
		</CardContent>
	</Card>

	<!-- Account Info -->
	<Card>
		<CardContent class="p-6">
			<h3 class="font-medium border-b pb-2 mb-4">Account Information</h3>
			<div class="grid grid-cols-2 gap-4 text-sm">
				<div>
					<p class="text-muted-foreground">User ID</p>
					<p class="font-mono text-xs">{$auth.user?.id}</p>
				</div>
				<div>
					<p class="text-muted-foreground">Role</p>
					<p class="font-medium">{formatRole($auth.user?.role || 'user')}</p>
				</div>
			</div>
		</CardContent>
	</Card>
</div>
