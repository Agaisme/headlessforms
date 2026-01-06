<script>
	import Modal from './Modal.svelte';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { auth } from '$lib/stores/auth';
	import { toast } from '$lib/stores/toast';

	export let isOpen = false;
	export let form = null; // null for create, object for edit
	export let onsuccess = null; // Svelte 5 callback prop
	
	let name = '';
	let redirectUrl = '';
	let notifyEmails = '';
	let webhookUrl = '';
	let webhookSecret = '';
	let accessMode = 'public';
	let submissionKey = '';
	let showAdvanced = false;
	let loading = false;
	let error = '';

	$: isEditing = form !== null;
	$: modalTitle = isEditing ? 'Edit Form' : 'Create New Form';

	// Reset form when modal opens
	$: if (isOpen) {
		if (form) {
			name = form.name || '';
			redirectUrl = form.redirect_url || '';
			notifyEmails = (form.notify_emails || []).join(', ');
			webhookUrl = form.webhook_url || '';
			webhookSecret = form.webhook_secret || '';
			accessMode = form.access_mode || 'public';
			submissionKey = form.submission_key || '';
		} else {
			name = '';
			redirectUrl = '';
			notifyEmails = '';
			webhookUrl = '';
			webhookSecret = '';
			accessMode = 'public';
			submissionKey = '';
			showAdvanced = false;
		}
		error = '';
	}

	function generateSecret() {
		const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
		webhookSecret = Array.from({ length: 32 }, () => chars[Math.floor(Math.random() * chars.length)]).join('');
	}

	function generateSubmissionKey() {
		const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
		submissionKey = Array.from({ length: 24 }, () => chars[Math.floor(Math.random() * chars.length)]).join('');
	}

	async function handleSubmit() {
		if (!name.trim()) {
			error = 'Form name is required';
			return;
		}

		// Require submission key for with_key mode
		if (accessMode === 'with_key' && !submissionKey) {
			error = 'Submission key is required for "With Key" access mode';
			return;
		}

		loading = true;
		error = '';

		try {
			const token = auth.getToken();
			const emails = notifyEmails
				.split(',')
				.map(e => e.trim())
				.filter(e => e);

			const body = {
				name: name.trim(),
				redirect_url: redirectUrl.trim() || null,
				notify_emails: emails,
				webhook_url: webhookUrl.trim() || null,
				webhook_secret: webhookSecret.trim() || null,
				access_mode: accessMode,
				submission_key: accessMode === 'with_key' ? submissionKey.trim() : ''
			};

			const url = isEditing ? `/api/v1/forms/${form.public_id}` : '/api/v1/forms';
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
				throw new Error(data.message || 'Failed to save form');
			}

			const data = await res.json();
			toast.success(isEditing ? 'Form updated successfully' : 'Form created successfully');
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

<Modal bind:isOpen title={modalTitle} size="lg" onclose={handleClose}>
	<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-5">
		{#if error}
			<div class="bg-destructive/10 text-destructive p-4 rounded-lg flex items-center gap-3 text-sm">
				<svg xmlns="http://www.w3.org/2000/svg" class="shrink-0 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				<span>{error}</span>
			</div>
		{/if}

		<!-- Form Name -->
		<div class="space-y-1.5">
			<label for="form-name" class="text-sm font-medium flex items-center justify-between">
				<span>Form Name</span>
				<span class="text-destructive text-xs font-normal">Required</span>
			</label>
			<Input
				id="form-name"
				type="text"
				bind:value={name}
				placeholder="e.g. Contact Form, Newsletter Signup"
				required
			/>
			<p class="text-xs text-muted-foreground">A descriptive name for your form</p>
		</div>

		<!-- Access Mode -->
		<div class="space-y-1.5">
			<label class="text-sm font-medium">Access Mode</label>
			<div class="grid grid-cols-3 gap-2">
				<button
					type="button"
					class="p-3 rounded-lg border text-left transition-all {accessMode === 'public' ? 'border-primary bg-primary/5 ring-1 ring-primary' : 'border-border hover:bg-muted/50'}"
					onclick={() => accessMode = 'public'}
				>
					<span class="font-medium text-sm block">Public</span>
					<span class="text-xs text-muted-foreground">Anyone can submit</span>
				</button>
				<button
					type="button"
					class="p-3 rounded-lg border text-left transition-all {accessMode === 'with_key' ? 'border-primary bg-primary/5 ring-1 ring-primary' : 'border-border hover:bg-muted/50'}"
					onclick={() => accessMode = 'with_key'}
				>
					<span class="font-medium text-sm block">With Key</span>
					<span class="text-xs text-muted-foreground">Requires hidden field</span>
				</button>
				<button
					type="button"
					class="p-3 rounded-lg border text-left transition-all {accessMode === 'private' ? 'border-primary bg-primary/5 ring-1 ring-primary' : 'border-border hover:bg-muted/50'}"
					onclick={() => accessMode = 'private'}
				>
					<span class="font-medium text-sm block">Private</span>
					<span class="text-xs text-muted-foreground">Auth required</span>
				</button>
			</div>
		</div>

		<!-- Submission Key (only for with_key mode) -->
		{#if accessMode === 'with_key'}
			<div class="space-y-1.5 p-4 rounded-lg bg-muted/30 border border-border">
				<label for="submission-key" class="text-sm font-medium flex items-center justify-between">
					<span>Submission Key</span>
					<button type="button" class="text-primary text-xs hover:underline" onclick={generateSubmissionKey}>
						Generate
					</button>
				</label>
				<Input
					id="submission-key"
					type="text"
					bind:value={submissionKey}
					placeholder="Your secret submission key"
				/>
				<p class="text-xs text-muted-foreground">
					Add this as a hidden field: <code class="bg-muted px-1 rounded">&lt;input type="hidden" name="_key" value="{submissionKey || 'YOUR_KEY'}"&gt;</code>
				</p>
			</div>
		{/if}

		<!-- Redirect URL -->
		<div class="space-y-1.5">
			<label for="redirect-url" class="text-sm font-medium flex items-center justify-between">
				<span>Redirect URL</span>
				<span class="text-muted-foreground text-xs font-normal">Optional</span>
			</label>
			<Input
				id="redirect-url"
				type="url"
				bind:value={redirectUrl}
				placeholder="https://example.com/thank-you"
			/>
			<p class="text-xs text-muted-foreground">Where to redirect users after form submission</p>
		</div>

		<!-- Notification Emails -->
		<div class="space-y-1.5">
			<label for="notify-emails" class="text-sm font-medium flex items-center justify-between">
				<span>Notification Emails</span>
				<span class="text-muted-foreground text-xs font-normal">Optional</span>
			</label>
			<Input
				id="notify-emails"
				type="text"
				bind:value={notifyEmails}
				placeholder="admin@example.com, team@example.com"
			/>
			<p class="text-xs text-muted-foreground">Comma-separated list of emails to notify on new submissions</p>
		</div>

		<!-- Advanced/Webhook Section -->
		<div class="border border-border rounded-lg overflow-hidden">
			<button
				type="button"
				class="w-full px-4 py-3 flex items-center justify-between text-sm font-medium hover:bg-muted/50 transition-colors"
				onclick={() => showAdvanced = !showAdvanced}
			>
				<span class="flex items-center gap-2">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-muted-foreground" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M12.316 3.051a1 1 0 01.633 1.265l-4 12a1 1 0 11-1.898-.632l4-12a1 1 0 011.265-.633zM5.707 6.293a1 1 0 010 1.414L3.414 10l2.293 2.293a1 1 0 11-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0zm8.586 0a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 11-1.414-1.414L16.586 10l-2.293-2.293a1 1 0 010-1.414z" clip-rule="evenodd" />
					</svg>
					Webhook Integration
				</span>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-muted-foreground transition-transform {showAdvanced ? 'rotate-180' : ''}" viewBox="0 0 20 20" fill="currentColor">
					<path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
				</svg>
			</button>
			
			{#if showAdvanced}
				<div class="px-4 pb-4 space-y-4 border-t border-border">
					<div class="pt-4 space-y-1.5">
						<label for="webhook-url" class="text-sm font-medium flex items-center justify-between">
							<span>Webhook URL</span>
							<span class="text-muted-foreground text-xs font-normal">Optional</span>
						</label>
						<Input
							id="webhook-url"
							type="url"
							bind:value={webhookUrl}
							placeholder="https://example.com/webhook"
						/>
						<p class="text-xs text-muted-foreground">We'll POST submission data to this URL</p>
					</div>

					<div class="space-y-1.5">
						<label for="webhook-secret" class="text-sm font-medium flex items-center justify-between">
							<span>Webhook Secret</span>
							<button type="button" class="text-primary text-xs hover:underline" onclick={generateSecret}>
								Generate
							</button>
						</label>
						<Input
							id="webhook-secret"
							type="text"
							bind:value={webhookSecret}
							placeholder="Optional signing secret"
						/>
						<p class="text-xs text-muted-foreground">Used to sign webhook payloads (HMAC-SHA256)</p>
					</div>
				</div>
			{/if}
		</div>
	</form>

	<svelte:fragment slot="footer">
		<Button variant="ghost" onclick={handleClose} disabled={loading}>
			Cancel
		</Button>
		<Button onclick={handleSubmit} loading={loading} disabled={!name.trim()}>
			{isEditing ? 'Save Changes' : 'Create Form'}
		</Button>
	</svelte:fragment>
</Modal>
