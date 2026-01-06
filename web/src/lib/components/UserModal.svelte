<script>
	import Modal from './Modal.svelte';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { auth } from '$lib/stores/auth';
	import { toast } from '$lib/stores/toast';

	export let isOpen = false;
	export let user = null; // null for create, object for edit
	export let onsuccess = null; // Svelte 5 callback prop
	
	let email = '';
	let name = '';
	let password = '';
	let role = 'user';
	let loading = false;
	let error = '';

	$: isEditing = user !== null;
	$: modalTitle = isEditing ? 'Edit User' : 'Create New User';

	// Reset form when modal opens
	$: if (isOpen) {
		if (user) {
			email = user.email || '';
			name = user.name || '';
			role = user.role || 'user';
			password = ''; // Don't expose password
		} else {
			email = '';
			name = '';
			password = '';
			role = 'user';
		}
		error = '';
	}

	async function handleSubmit() {
		if (!email.trim()) {
			error = 'Email is required';
			return;
		}
		if (!isEditing && !password) {
			error = 'Password is required for new users';
			return;
		}
		if (!isEditing && password.length < 8) {
			error = 'Password must be at least 8 characters';
			return;
		}

		loading = true;
		error = '';

		try {
			const token = auth.getToken();

			const body = {
				email: email.trim(),
				name: name.trim(),
				role
			};

			if (!isEditing || password) {
				body.password = password;
			}

			const url = isEditing ? `/api/v1/users/${user.id}` : '/api/v1/users';
			const method = isEditing ? 'PUT' : 'POST';

			const res = await fetch(url, {
				method,
				headers: {
					'Content-Type': 'application/json',
					...(token ? { 'Authorization': `Bearer ${token}` } : {})
				},
				body: JSON.stringify(body)
			});

			if (!res.ok) {
				const data = await res.json();
				throw new Error(data.message || 'Failed to save user');
			}

			const data = await res.json();
			toast.success(isEditing ? 'User updated successfully' : 'User created successfully');
			if (onsuccess) {
				onsuccess(data.data);
			}
			isOpen = false;
		} catch (e) {
			error = e.message || 'An error occurred';
		} finally {
			loading = false;
		}
	}

	function handleClose() {
		isOpen = false;
	}
</script>

<Modal bind:isOpen title={modalTitle} size="md" onclose={handleClose}>
	<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-5">
		{#if error}
			<div class="bg-destructive/10 text-destructive p-4 rounded-lg flex items-center gap-3 text-sm">
				<svg xmlns="http://www.w3.org/2000/svg" class="shrink-0 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				<span>{error}</span>
			</div>
		{/if}

		<!-- Email -->
		<div class="space-y-1.5">
			<label for="user-email" class="text-sm font-medium flex items-center justify-between">
				<span>Email Address</span>
				<span class="text-destructive text-xs font-normal">Required</span>
			</label>
			<Input
				id="user-email"
				type="email"
				bind:value={email}
				placeholder="user@example.com"
				required
				disabled={isEditing}
			/>
			{#if isEditing}
				<p class="text-xs text-muted-foreground">Email cannot be changed after creation</p>
			{/if}
		</div>

		<!-- Name -->
		<div class="space-y-1.5">
			<label for="user-name" class="text-sm font-medium flex items-center justify-between">
				<span>Full Name</span>
				<span class="text-muted-foreground text-xs font-normal">Optional</span>
			</label>
			<Input
				id="user-name"
				type="text"
				bind:value={name}
				placeholder="John Doe"
			/>
		</div>

		<!-- Password -->
		<div class="space-y-1.5">
			<label for="user-password" class="text-sm font-medium flex items-center justify-between">
				<span>Password</span>
				{#if isEditing}
					<span class="text-muted-foreground text-xs font-normal">Leave blank to keep current</span>
				{:else}
					<span class="text-destructive text-xs font-normal">Required (min 8 chars)</span>
				{/if}
			</label>
			<Input
				id="user-password"
				type="password"
				bind:value={password}
				placeholder="••••••••"
			/>
		</div>

		<!-- Role -->
		<div class="space-y-1.5">
			<label for="user-role" class="text-sm font-medium">Role</label>
			<select
				id="user-role"
				bind:value={role}
				class="w-full h-10 px-3 py-2 rounded-md border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
			>
				<option value="user">User — Manage own forms only</option>
				<option value="admin">Admin — Manage all forms & users</option>
				<option value="super_admin">Super Admin — Full system access</option>
			</select>
			<p class="text-xs text-muted-foreground">
				{#if role === 'user'}
					Can create and manage their own forms
				{:else if role === 'admin'}
					Can manage all forms and users (except super admins)
				{:else}
					Full access including system settings
				{/if}
			</p>
		</div>
	</form>

	<svelte:fragment slot="footer">
		<Button variant="ghost" onclick={handleClose} disabled={loading}>
			Cancel
		</Button>
		<Button onclick={handleSubmit} loading={loading} disabled={!email.trim()}>
			{isEditing ? 'Save Changes' : 'Create User'}
		</Button>
	</svelte:fragment>
</Modal>
