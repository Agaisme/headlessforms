<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/stores/toast';
	import { auth } from '$lib/stores/auth';

	let formId = '';
	let form: any = null;
	let loading = true;
	let saving = false;
	let error: string | null = null;

	// Form fields
	let name = '';
	let redirectUrl = '';
	let notifyEmails = '';
	let status: 'active' | 'inactive' = 'active';
	let accessMode: 'public' | 'with_key' | 'private' = 'public';
	let submissionKey = '';
	let webhookUrl = '';
	let webhookSecret = '';

	$: formId = $page.params.id;

	onMount(async () => {
		await loadForm();
	});

	async function loadForm() {
		loading = true;
		error = null;
		try {
			const token = auth.getToken();
			const res = await fetch(`/api/v1/forms/${formId}`, {
				headers: token ? { 'Authorization': `Bearer ${token}` } : {}
			});
			const json = await res.json();
			
			if (json.status === 'success') {
				form = json.data;
				// Populate form fields
				name = form.name || '';
				redirectUrl = form.redirect_url || '';
				notifyEmails = (form.notify_emails || []).join(', ');
				status = form.status || 'active';
				accessMode = form.access_mode || 'public';
				submissionKey = form.submission_key || '';
				webhookUrl = form.webhook_url || '';
				webhookSecret = form.webhook_secret || '';
			} else {
				error = json.message || 'Failed to load form';
			}
		} catch (e) {
			error = 'Failed to load form';
		} finally {
			loading = false;
		}
	}

	async function handleSave() {
		saving = true;
		try {
			const token = auth.getToken();
			const res = await fetch(`/api/v1/forms/${formId}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					...(token ? { 'Authorization': `Bearer ${token}` } : {})
				},
				body: JSON.stringify({
					name,
					redirect_url: redirectUrl,
					notify_emails: notifyEmails.split(',').map(e => e.trim()).filter(e => e),
					status,
					access_mode: accessMode,
					submission_key: submissionKey,
					webhook_url: webhookUrl,
					webhook_secret: webhookSecret
				})
			});

			const json = await res.json();
			if (json.status === 'success') {
				toast.success('Form updated successfully');
				await loadForm();
			} else {
				toast.error(json.message || 'Failed to update form');
			}
		} catch (e) {
			toast.error('Failed to update form');
		} finally {
			saving = false;
		}
	}

	async function toggleStatus() {
		status = status === 'active' ? 'inactive' : 'active';
		await handleSave();
	}

	async function handleDelete() {
		if (!confirm('Are you sure you want to delete this form? This action cannot be undone.')) {
			return;
		}

		try {
			const token = auth.getToken();
			const res = await fetch(`/api/v1/forms/${formId}`, {
				method: 'DELETE',
				headers: token ? { 'Authorization': `Bearer ${token}` } : {}
			});

			const json = await res.json();
			if (json.status === 'success') {
				toast.success('Form deleted successfully');
				goto('/');
			} else {
				toast.error(json.message || 'Failed to delete form');
			}
		} catch (e) {
			toast.error('Failed to delete form');
		}
	}

	function generateSubmissionKey() {
		submissionKey = crypto.randomUUID().replace(/-/g, '').substring(0, 24);
	}

	function getEndpointUrl() {
		if (typeof window !== 'undefined') {
			return `${window.location.origin}/api/v1/submissions/${formId}`;
		}
		return `/api/v1/submissions/${formId}`;
	}

	function copyToClipboard(text: string) {
		navigator.clipboard.writeText(text);
		toast.success('Copied to clipboard');
	}
</script>

<div class="max-w-3xl mx-auto space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<Button variant="ghost" size="sm" href="/forms/{formId}" class="gap-2 mb-2">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
					<path fill-rule="evenodd" d="M9.707 16.707a1 1 0 01-1.414 0l-6-6a1 1 0 010-1.414l6-6a1 1 0 011.414 1.414L5.414 9H17a1 1 0 110 2H5.414l4.293 4.293a1 1 0 010 1.414z" clip-rule="evenodd" />
				</svg>
				Back to Form
			</Button>
			<h1 class="text-2xl font-bold">Edit Form</h1>
		</div>
	</div>

	{#if loading}
		<div class="space-y-4">
			<Skeleton class="h-32 w-full" />
			<Skeleton class="h-48 w-full" />
		</div>
	{:else if error}
		<Card>
			<CardContent class="p-6">
				<div class="text-destructive">{error}</div>
			</CardContent>
		</Card>
	{:else}
		<!-- Basic Settings -->
		<Card>
			<CardHeader class="pb-3">
				<div class="flex items-center justify-between">
					<h2 class="text-lg font-semibold">Basic Settings</h2>
					<button
						onclick={toggleStatus}
						class="flex items-center gap-2 px-3 py-1.5 rounded-full text-sm font-medium transition-colors
						       {status === 'active' 
						         ? 'bg-emerald-500/10 text-emerald-500 hover:bg-emerald-500/20' 
						         : 'bg-zinc-500/10 text-zinc-400 hover:bg-zinc-500/20'}"
					>
						<span class="w-2 h-2 rounded-full {status === 'active' ? 'bg-emerald-500' : 'bg-zinc-500'}"></span>
						{status === 'active' ? 'Active' : 'Inactive'}
					</button>
				</div>
			</CardHeader>
			<CardContent class="space-y-4">
				<div>
					<label class="text-sm font-medium mb-1.5 block">Form Name</label>
					<Input bind:value={name} placeholder="Contact Form" />
				</div>

				<div>
					<label class="text-sm font-medium mb-1.5 block">
						Endpoint Form Submission URL
						<Badge variant="outline" class="ml-2">POST</Badge>
					</label>
					<div class="flex gap-2">
						<code class="flex-1 bg-muted px-3 py-2 rounded text-sm font-mono truncate">{getEndpointUrl()}</code>
						<Button variant="ghost" size="icon" onclick={() => copyToClipboard(getEndpointUrl())} title="Copy URL">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
								<path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
								<path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
							</svg>
						</Button>
					</div>
				</div>

				<div>
					<label class="text-sm font-medium mb-1.5 block">Redirect URL (after submission)</label>
					<Input bind:value={redirectUrl} placeholder="https://yoursite.com/thank-you" />
				</div>

				<div>
					<label class="text-sm font-medium mb-1.5 block">Notification Emails (comma separated)</label>
					<Input bind:value={notifyEmails} placeholder="admin@yoursite.com, team@yoursite.com" />
				</div>
			</CardContent>
		</Card>

		<!-- Access Control -->
		<Card>
			<CardHeader class="pb-3">
				<h2 class="text-lg font-semibold">Access Control</h2>
				<p class="text-sm text-muted-foreground">Control who can submit to this form</p>
			</CardHeader>
			<CardContent class="space-y-4">
				<div class="grid grid-cols-3 gap-3">
					<button
						onclick={() => accessMode = 'public'}
						class="p-4 rounded-lg border text-left transition-all
						       {accessMode === 'public' 
						         ? 'border-primary bg-primary/5 ring-1 ring-primary' 
						         : 'border-border hover:border-primary/50'}"
					>
						<div class="font-medium mb-1">Public</div>
						<p class="text-xs text-muted-foreground">Anyone can submit</p>
					</button>
					
					<button
						onclick={() => accessMode = 'with_key'}
						class="p-4 rounded-lg border text-left transition-all
						       {accessMode === 'with_key' 
						         ? 'border-primary bg-primary/5 ring-1 ring-primary' 
						         : 'border-border hover:border-primary/50'}"
					>
						<div class="font-medium mb-1">With Key</div>
						<p class="text-xs text-muted-foreground">Requires secret key</p>
					</button>
					
					<button
						onclick={() => accessMode = 'private'}
						class="p-4 rounded-lg border text-left transition-all
						       {accessMode === 'private' 
						         ? 'border-primary bg-primary/5 ring-1 ring-primary' 
						         : 'border-border hover:border-primary/50'}"
					>
						<div class="font-medium mb-1">Private</div>
						<p class="text-xs text-muted-foreground">Authenticated users only</p>
					</button>
				</div>

				{#if accessMode === 'with_key'}
					<div class="p-4 bg-muted/50 rounded-lg space-y-3">
						<div class="flex items-center justify-between">
							<label class="text-sm font-medium">Submission Key</label>
							<Button variant="ghost" size="sm" onclick={generateSubmissionKey}>Generate New</Button>
						</div>
						<div class="flex gap-2">
							<Input bind:value={submissionKey} placeholder="Enter or generate a secret key" class="font-mono" />
							<Button variant="ghost" size="icon" onclick={() => copyToClipboard(submissionKey)} title="Copy Key">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
									<path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
									<path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
								</svg>
							</Button>
						</div>
						<p class="text-xs text-muted-foreground">
							Add this as a hidden field in your form: <code class="bg-muted px-1 rounded">_submission_key</code>
						</p>
					</div>
				{/if}
			</CardContent>
		</Card>

		<!-- Webhook Settings -->
		<Card>
			<CardHeader class="pb-3">
				<h2 class="text-lg font-semibold">Webhook</h2>
				<p class="text-sm text-muted-foreground">Send submissions to an external URL</p>
			</CardHeader>
			<CardContent class="space-y-4">
				<div>
					<label class="text-sm font-medium mb-1.5 block">Webhook URL</label>
					<Input bind:value={webhookUrl} placeholder="https://your-server.com/webhook" />
				</div>
				<div>
					<label class="text-sm font-medium mb-1.5 block">Webhook Secret (for HMAC signature)</label>
					<Input bind:value={webhookSecret} type="password" placeholder="Optional: shared secret" />
				</div>
			</CardContent>
		</Card>

		<!-- Actions -->
		<div class="flex items-center justify-between pt-4">
			<Button variant="destructive" onclick={handleDelete} class="gap-2">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
					<path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
				</svg>
				Delete Form
			</Button>
			
			<Button onclick={handleSave} disabled={saving} class="gap-2">
				{#if saving}
					<svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
						<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
					</svg>
					Saving...
				{:else}
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
					</svg>
					Save Changes
				{/if}
			</Button>
		</div>
	{/if}
</div>
