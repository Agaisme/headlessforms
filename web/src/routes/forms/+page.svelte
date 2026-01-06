<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { auth } from '$lib/stores/auth';
	import { toast } from '$lib/stores/toast';
	import FormModal from '$lib/components/FormModal.svelte';

	interface Form {
		public_id: string;
		name: string;
		status: string;
		submission_count: number;
		created_at: string;
		redirect_url?: string;
		notify_emails?: string[];
		webhook_url?: string;
		access_mode?: string;
	}

	interface Submission {
		id: string;
		data: any;
		meta: any;
		status: string;
		created_at: string;
	}

	let forms: Form[] = [];
	let loading = true;
	let selectedForm: Form | null = null;
	
	// Submissions state
	let submissions: Submission[] = [];
	let loadingSubmissions = false;
	let selectedSubmission: Submission | null = null;
	
	// Modal state
	let showFormModal = false;
	let editingForm: Form | null = null;

	// Mobile view state: 'forms' | 'inbox' | 'detail'
	let mobileView: 'forms' | 'inbox' | 'detail' = 'forms';

	onMount(async () => {
		await loadForms();
	});

	async function loadForms() {
		loading = true;
		try {
			const token = auth.getToken();
			const res = await fetch('/api/v1/forms?limit=100', {
				headers: token ? { 'Authorization': `Bearer ${token}` } : {}
			});
			const json = await res.json();
			if (json.status === 'success') {
				forms = json.data?.forms || [];
				if (forms.length > 0 && !selectedForm) {
					selectForm(forms[0]);
				}
			}
		} catch (e) {
			toast.error('Failed to load forms');
		} finally {
			loading = false;
		}
	}

	async function selectForm(form: Form) {
		selectedForm = form;
		selectedSubmission = null;
		mobileView = 'inbox';
		await loadSubmissions(form.public_id);
	}

	async function loadSubmissions(formId: string) {
		loadingSubmissions = true;
		submissions = [];
		try {
			const token = auth.getToken();
			const res = await fetch(`/api/v1/forms/${formId}/submissions?limit=50`, {
				headers: token ? { 'Authorization': `Bearer ${token}` } : {}
			});
			const json = await res.json();
			if (json.status === 'success') {
				submissions = json.data?.submissions || [];
				if (submissions.length > 0) {
					selectSubmission(submissions[0]);
				}
			}
		} catch (e) {
			console.error('Failed to load submissions');
		} finally {
			loadingSubmissions = false;
		}
	}

	async function selectSubmission(sub: Submission) {
		selectedSubmission = sub;
		mobileView = 'detail';
		if (sub.status === 'unread') {
			try {
				const token = auth.getToken();
				await fetch(`/api/v1/submissions/${sub.id}/read`, {
					method: 'PUT',
					headers: token ? { 'Authorization': `Bearer ${token}` } : {}
				});
				sub.status = 'read';
				submissions = [...submissions];
			} catch (e) {}
		}
	}

	function openCreateModal() {
		editingForm = null;
		showFormModal = true;
	}

	function openEditModal() {
		if (selectedForm) {
			editingForm = selectedForm;
			showFormModal = true;
		}
	}

	async function handleFormSuccess() {
		await loadForms();
	}

	async function handleDeleteSubmission(subId: string) {
		if (!confirm('Delete this submission?')) return;
		try {
			const token = auth.getToken();
			await fetch(`/api/v1/submissions/${subId}`, {
				method: 'DELETE',
				headers: token ? { 'Authorization': `Bearer ${token}` } : {}
			});
			submissions = submissions.filter(s => s.id !== subId);
			if (selectedSubmission?.id === subId) {
				selectedSubmission = submissions.length > 0 ? submissions[0] : null;
			}
			toast.success('Submission deleted');
		} catch (e) {
			toast.error('Failed to delete');
		}
	}

	function formatDate(dateStr: string) {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const days = Math.floor(diff / (1000 * 60 * 60 * 24));
		
		if (days === 0) return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
		if (days === 1) return 'Yesterday';
		if (days < 7) return date.toLocaleDateString([], { weekday: 'short' });
		return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
	}

	function getSubmissionPreview(sub: Submission) {
		const data = typeof sub.data === 'string' ? JSON.parse(sub.data) : sub.data;
		return data?.email || data?.name || data?.message?.substring(0, 50) || 'Submission';
	}

	function getSubmissionSubject(sub: Submission) {
		const data = typeof sub.data === 'string' ? JSON.parse(sub.data) : sub.data;
		return data?.subject || data?.message?.substring(0, 30) || 'No subject';
	}

	function goBackMobile() {
		if (mobileView === 'detail') mobileView = 'inbox';
		else if (mobileView === 'inbox') mobileView = 'forms';
	}
</script>

<!-- Mobile Header (visible only on mobile) -->
<div class="lg:hidden flex items-center gap-2 mb-4">
	{#if mobileView !== 'forms'}
		<Button variant="ghost" size="sm" onclick={goBackMobile} class="p-2">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M9.707 16.707a1 1 0 01-1.414 0l-6-6a1 1 0 010-1.414l6-6a1 1 0 011.414 1.414L5.414 9H17a1 1 0 110 2H5.414l4.293 4.293a1 1 0 010 1.414z" clip-rule="evenodd" />
			</svg>
		</Button>
	{/if}
	<h1 class="font-semibold text-lg">
		{#if mobileView === 'forms'}Forms
		{:else if mobileView === 'inbox'}{selectedForm?.name || 'Inbox'}
		{:else}Submission
		{/if}
	</h1>
	{#if mobileView === 'forms'}
		<Button size="sm" onclick={openCreateModal} class="ml-auto h-8 w-8 p-0">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" />
			</svg>
		</Button>
	{/if}
</div>

<!-- Desktop 3-Panel Layout / Mobile Single Panel -->
<div class="flex gap-4 h-[calc(100vh-8rem)] lg:h-[calc(100vh-8rem)]">
	
	<!-- Panel 1: Form List -->
	<div class="w-full lg:w-72 flex-shrink-0 flex flex-col bg-card border border-border rounded-xl overflow-hidden {mobileView !== 'forms' ? 'hidden lg:flex' : 'flex'}">
		<!-- Desktop Header -->
		<div class="hidden lg:flex p-4 border-b border-border items-center justify-between">
			<h2 class="font-semibold">Forms</h2>
			<Button size="sm" onclick={openCreateModal} class="h-8 w-8 p-0">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
					<path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" />
				</svg>
			</Button>
		</div>

		<!-- Form List -->
		<div class="flex-1 overflow-y-auto">
			{#if loading}
				<div class="p-4 space-y-3">
					{#each [1, 2, 3, 4, 5] as _}
						<div class="flex items-center gap-3 p-3 rounded-lg bg-muted/30 animate-pulse">
							<div class="w-10 h-10 rounded-lg bg-muted"></div>
							<div class="flex-1 space-y-2">
								<Skeleton class="h-4 w-3/4" />
								<Skeleton class="h-3 w-1/2" />
							</div>
						</div>
					{/each}
				</div>
			{:else if forms.length === 0}
				<div class="p-6 text-center text-muted-foreground">
					<div class="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
						</svg>
					</div>
					<p class="text-sm mb-4">No forms yet</p>
					<Button size="sm" onclick={openCreateModal}>Create Form</Button>
				</div>
			{:else}
				{#each forms as form}
					<button
						class="w-full px-4 py-3 text-left border-b border-border hover:bg-muted/50 transition-all duration-200 {selectedForm?.public_id === form.public_id ? 'bg-primary/5 border-l-2 border-l-primary' : ''}"
						onclick={() => selectForm(form)}
					>
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-primary" viewBox="0 0 20 20" fill="currentColor">
									<path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
									<path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clip-rule="evenodd" />
								</svg>
							</div>
							<div class="flex-1 min-w-0">
								<div class="flex items-center justify-between">
									<span class="font-medium text-sm truncate">{form.name}</span>
									{#if form.submission_count > 0}
										<Badge variant="secondary" class="text-xs ml-2">{form.submission_count}</Badge>
									{/if}
								</div>
								<p class="text-xs text-muted-foreground mt-0.5">
									{form.access_mode === 'private' ? '🔒' : form.access_mode === 'with_key' ? '🔑' : '🌐'} 
									{form.status === 'active' ? 'Active' : 'Inactive'}
								</p>
							</div>
							<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-muted-foreground lg:hidden" viewBox="0 0 20 20" fill="currentColor">
								<path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
							</svg>
						</div>
					</button>
				{/each}
			{/if}
		</div>
	</div>

	<!-- Panel 2: Submissions Inbox -->
	<div class="w-full lg:w-80 flex-shrink-0 flex flex-col bg-card border border-border rounded-xl overflow-hidden {mobileView !== 'inbox' ? 'hidden lg:flex' : 'flex'}">
		{#if selectedForm}
			<!-- Inbox Header -->
			<div class="p-4 border-b border-border">
				<div class="flex items-center justify-between">
					<div>
						<h3 class="font-semibold hidden lg:block">{selectedForm.name}</h3>
						<p class="text-xs text-muted-foreground">{submissions.length} submissions</p>
					</div>
					<div class="flex gap-1">
						<Button variant="ghost" size="sm" onclick={openEditModal} title="Edit form" class="h-8 w-8 p-0">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
								<path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
							</svg>
						</Button>
						<Button variant="ghost" size="sm" href="/forms/{selectedForm.public_id}" title="Full view" class="h-8 w-8 p-0">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
								<path d="M11 3a1 1 0 100 2h2.586l-6.293 6.293a1 1 0 101.414 1.414L15 6.414V9a1 1 0 102 0V4a1 1 0 00-1-1h-5z" />
								<path d="M5 5a2 2 0 00-2 2v8a2 2 0 002 2h8a2 2 0 002-2v-3a1 1 0 10-2 0v3H5V7h3a1 1 0 000-2H5z" />
							</svg>
						</Button>
					</div>
				</div>
			</div>

			<!-- Submissions List -->
			<div class="flex-1 overflow-y-auto">
				{#if loadingSubmissions}
					<div class="p-4 space-y-3">
						{#each [1, 2, 3, 4] as _}
							<div class="flex items-start gap-3 p-3 rounded-lg bg-muted/30 animate-pulse">
								<div class="w-2 h-2 rounded-full bg-muted mt-2"></div>
								<div class="flex-1 space-y-2">
									<Skeleton class="h-4 w-3/4" />
									<Skeleton class="h-3 w-full" />
								</div>
								<Skeleton class="h-3 w-12" />
							</div>
						{/each}
					</div>
				{:else if submissions.length === 0}
					<div class="flex-1 flex flex-col items-center justify-center p-8 text-muted-foreground">
						<div class="w-16 h-16 rounded-full bg-muted flex items-center justify-center mb-4">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
							</svg>
						</div>
						<p class="text-sm">No submissions yet</p>
						<p class="text-xs mt-1">Waiting for first submission</p>
					</div>
				{:else}
					{#each submissions as sub}
						<button
							class="w-full px-4 py-3 text-left border-b border-border hover:bg-muted/50 transition-all duration-200 {selectedSubmission?.id === sub.id ? 'bg-primary/5' : ''}"
							onclick={() => selectSubmission(sub)}
						>
							<div class="flex items-start justify-between gap-3">
								<div class="flex items-start gap-3 flex-1 min-w-0">
									{#if sub.status === 'unread'}
										<span class="w-2 h-2 rounded-full bg-primary flex-shrink-0 mt-2 animate-pulse"></span>
									{:else}
										<span class="w-2 h-2 rounded-full bg-transparent flex-shrink-0 mt-2"></span>
									{/if}
									<div class="flex-1 min-w-0">
										<span class="font-medium text-sm truncate block {sub.status === 'unread' ? '' : 'text-muted-foreground'}">{getSubmissionPreview(sub)}</span>
										<p class="text-xs text-muted-foreground truncate mt-0.5">{getSubmissionSubject(sub)}</p>
									</div>
								</div>
								<div class="flex items-center gap-2 flex-shrink-0">
									<span class="text-xs text-muted-foreground">{formatDate(sub.created_at)}</span>
									<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-muted-foreground lg:hidden" viewBox="0 0 20 20" fill="currentColor">
										<path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
									</svg>
								</div>
							</div>
						</button>
					{/each}
				{/if}
			</div>
		{:else}
			<div class="flex-1 flex items-center justify-center text-muted-foreground p-8">
				<div class="text-center">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto mb-3 opacity-30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 4H6a2 2 0 00-2 2v12a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-2m-4-1v8m0 0l3-3m-3 3L9 8m-5 5h2.586a1 1 0 01.707.293l2.414 2.414a1 1 0 00.707.293h3.172a1 1 0 00.707-.293l2.414-2.414a1 1 0 01.707-.293H20" />
					</svg>
					<p class="text-sm">Select a form</p>
				</div>
			</div>
		{/if}
	</div>

	<!-- Panel 3: Submission Detail -->
	<div class="w-full lg:flex-1 flex flex-col bg-card border border-border rounded-xl overflow-hidden {mobileView !== 'detail' ? 'hidden lg:flex' : 'flex'}">
		{#if selectedSubmission}
			<!-- Detail Header -->
			<div class="p-4 border-b border-border flex items-center justify-between">
				<div class="min-w-0 flex-1">
					<h4 class="font-semibold truncate">{getSubmissionPreview(selectedSubmission)}</h4>
					<p class="text-xs text-muted-foreground">
						{new Date(selectedSubmission.created_at).toLocaleString()}
					</p>
				</div>
				<Button variant="ghost" size="sm" onclick={() => handleDeleteSubmission(selectedSubmission.id)} class="h-8 w-8 p-0 flex-shrink-0">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-destructive" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
					</svg>
				</Button>
			</div>

			<!-- Detail Content -->
			<div class="flex-1 overflow-y-auto p-4">
				<div class="space-y-4">
					<div>
						<h5 class="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-3">Submission Data</h5>
						<div class="bg-muted/30 rounded-xl p-4 space-y-4">
							{#each Object.entries(typeof selectedSubmission.data === 'string' ? JSON.parse(selectedSubmission.data) : selectedSubmission.data) as [key, value]}
								<div class="flex flex-col">
									<span class="text-xs font-medium text-muted-foreground capitalize">{key.replace(/_/g, ' ')}</span>
									<span class="text-sm mt-0.5 break-words">{value}</span>
								</div>
							{/each}
						</div>
					</div>

					{#if selectedSubmission.meta}
						<div>
							<h5 class="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-3">Metadata</h5>
							<div class="bg-muted/30 rounded-xl p-4 overflow-x-auto">
								<pre class="text-xs">{JSON.stringify(typeof selectedSubmission.meta === 'string' ? JSON.parse(selectedSubmission.meta) : selectedSubmission.meta, null, 2)}</pre>
							</div>
						</div>
					{/if}
				</div>
			</div>
		{:else if selectedForm}
			<div class="flex-1 flex items-center justify-center text-muted-foreground p-8">
				<div class="text-center">
					<div class="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
						</svg>
					</div>
					<p class="text-sm">Select a submission</p>
					<p class="text-xs mt-1">Click an item to view details</p>
				</div>
			</div>
		{:else}
			<div class="flex-1 flex items-center justify-center text-muted-foreground p-8">
				<div class="text-center">
					<div class="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 4H6a2 2 0 00-2 2v12a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-2m-4-1v8m0 0l3-3m-3 3L9 8" />
						</svg>
					</div>
					<p class="text-sm">Select a form to view inbox</p>
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Form Modal -->
<FormModal bind:isOpen={showFormModal} form={editingForm} onsuccess={handleFormSuccess} />
