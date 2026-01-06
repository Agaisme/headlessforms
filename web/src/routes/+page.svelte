<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { auth } from '$lib/stores/auth';

	let stats: any = null;
	let loading = true;
	let recentActivity: any[] = [];

	onMount(async () => {
		await loadStats();
	});

	async function loadStats() {
		loading = true;
		try {
			const token = auth.getToken();
			const res = await fetch('/api/v1/stats', {
				headers: token ? { 'Authorization': `Bearer ${token}` } : {}
			});
			const json = await res.json();
			if (json.status === 'success') {
				stats = json.data;
			}
		} catch (e) {
			console.error('Failed to load stats');
		} finally {
			loading = false;
		}
	}
</script>

<div class="space-y-6">
	<!-- Welcome Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold">Welcome back{$auth.user?.name ? `, ${$auth.user.name}` : ''}!</h1>
			<p class="text-muted-foreground">Here's an overview of your form submissions.</p>
		</div>
		<Button href="/forms" class="gap-2">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M10.293 5.293a1 1 0 011.414 0l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414-1.414L12.586 11H5a1 1 0 110-2h7.586l-2.293-2.293a1 1 0 010-1.414z" clip-rule="evenodd" />
			</svg>
			Go to Forms
		</Button>
	</div>

	<!-- Stats Cards -->
	{#if loading}
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
			{#each [1, 2, 3, 4] as _}
				<Card>
					<CardContent class="p-5">
						<Skeleton class="h-4 w-24 mb-2" />
						<Skeleton class="h-8 w-16" />
					</CardContent>
				</Card>
			{/each}
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
			<Card class="hover:border-primary/50 transition-colors">
				<CardContent class="p-5">
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm text-muted-foreground font-medium">Total Forms</p>
							<p class="text-3xl font-bold mt-1">{stats?.total_forms ?? 0}</p>
						</div>
						<div class="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-primary" viewBox="0 0 20 20" fill="currentColor">
								<path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
								<path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clip-rule="evenodd" />
							</svg>
						</div>
					</div>
					<p class="text-xs text-muted-foreground mt-2">{stats?.active_forms ?? 0} active</p>
				</CardContent>
			</Card>

			<Card class="hover:border-green-500/50 transition-colors">
				<CardContent class="p-5">
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm text-muted-foreground font-medium">Total Submissions</p>
							<p class="text-3xl font-bold mt-1">{stats?.total_submissions ?? 0}</p>
						</div>
						<div class="w-12 h-12 rounded-xl bg-green-500/10 flex items-center justify-center">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-green-500" viewBox="0 0 20 20" fill="currentColor">
								<path d="M2.003 5.884L10 9.882l7.997-3.998A2 2 0 0016 4H4a2 2 0 00-1.997 1.884z" />
								<path d="M18 8.118l-8 4-8-4V14a2 2 0 002 2h12a2 2 0 002-2V8.118z" />
							</svg>
						</div>
					</div>
					<p class="text-xs text-muted-foreground mt-2">{stats?.submissions_today ?? 0} today</p>
				</CardContent>
			</Card>

			<Card class="hover:border-yellow-500/50 transition-colors">
				<CardContent class="p-5">
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm text-muted-foreground font-medium">Unread</p>
							<p class="text-3xl font-bold mt-1 {(stats?.unread_submissions ?? 0) > 0 ? 'text-yellow-500' : ''}">{stats?.unread_submissions ?? 0}</p>
						</div>
						<div class="w-12 h-12 rounded-xl bg-yellow-500/10 flex items-center justify-center">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-yellow-500" viewBox="0 0 20 20" fill="currentColor">
								<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
							</svg>
						</div>
					</div>
					{#if (stats?.unread_submissions ?? 0) > 0}
						<Button variant="link" size="sm" href="/forms" class="text-xs p-0 h-auto mt-2">View inbox →</Button>
					{:else}
						<p class="text-xs text-muted-foreground mt-2">All caught up!</p>
					{/if}
				</CardContent>
			</Card>

			<Card class="hover:border-blue-500/50 transition-colors">
				<CardContent class="p-5">
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm text-muted-foreground font-medium">This Week</p>
							<p class="text-3xl font-bold mt-1">{stats?.submissions_this_week ?? 0}</p>
						</div>
						<div class="w-12 h-12 rounded-xl bg-blue-500/10 flex items-center justify-center">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-blue-500" viewBox="0 0 20 20" fill="currentColor">
								<path fill-rule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clip-rule="evenodd" />
							</svg>
						</div>
					</div>
					<p class="text-xs text-muted-foreground mt-2">vs last week</p>
				</CardContent>
			</Card>
		</div>
	{/if}

	<!-- Submission Trends Chart -->
	{#if stats?.daily_submissions && stats.daily_submissions.length > 0}
		<Card>
			<CardContent class="p-5">
				<div class="flex items-center justify-between mb-4">
					<div>
						<h3 class="font-semibold">Submission Trends</h3>
						<p class="text-xs text-muted-foreground">Last 7 days</p>
					</div>
					<Badge variant="outline">{stats.submissions_this_week || 0} this week</Badge>
				</div>
				<div class="flex items-end justify-between gap-2 h-40">
					{#each stats.daily_submissions as day}
						{@const maxCount = Math.max(...stats.daily_submissions.map(d => d.count), 1)}
						{@const heightPercent = maxCount > 0 ? (day.count / maxCount) * 100 : 0}
						<div class="flex-1 flex flex-col items-center gap-2">
							<div 
								class="w-full rounded-t transition-all bg-primary/20 hover:bg-primary/40 relative group cursor-pointer"
								style="height: {Math.max(heightPercent, 8)}%"
							>
								<div class="absolute -top-6 left-1/2 -translate-x-1/2 bg-foreground text-background text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
									{day.count} submissions
								</div>
								{#if day.count > 0}
									<div class="w-full h-full flex items-end justify-center pb-1">
										<span class="text-xs font-medium text-primary">{day.count}</span>
									</div>
								{/if}
							</div>
							<span class="text-[10px] text-muted-foreground">
								{new Date(day.date).toLocaleDateString('en-US', { weekday: 'short' })}
							</span>
						</div>
					{/each}
				</div>
			</CardContent>
		</Card>
	{/if}

	<!-- Quick Actions -->
	<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
		<Card class="hover:border-primary/50 transition-colors cursor-pointer group">
			<a href="/forms" class="block">
				<CardContent class="p-5 flex items-center gap-4">
					<div class="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-primary" viewBox="0 0 20 20" fill="currentColor">
							<path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
						</svg>
					</div>
					<div>
						<h3 class="font-semibold">View Forms</h3>
						<p class="text-sm text-muted-foreground">Manage forms & view submissions</p>
					</div>
				</CardContent>
			</a>
		</Card>

		{#if $auth.user?.role === 'admin' || $auth.user?.role === 'super_admin'}
			<Card class="hover:border-primary/50 transition-colors cursor-pointer group">
				<a href="/users" class="block">
					<CardContent class="p-5 flex items-center gap-4">
						<div class="w-12 h-12 rounded-xl bg-purple-500/10 flex items-center justify-center group-hover:bg-purple-500/20 transition-colors">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-purple-500" viewBox="0 0 20 20" fill="currentColor">
								<path d="M9 6a3 3 0 11-6 0 3 3 0 016 0zM17 6a3 3 0 11-6 0 3 3 0 016 0zM12.93 17c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z" />
							</svg>
						</div>
						<div>
							<h3 class="font-semibold">Manage Users</h3>
							<p class="text-sm text-muted-foreground">Add or edit user accounts</p>
						</div>
					</CardContent>
				</a>
			</Card>
		{/if}

		{#if $auth.user?.role === 'super_admin'}
			<Card class="hover:border-primary/50 transition-colors cursor-pointer group">
				<a href="/settings" class="block">
					<CardContent class="p-5 flex items-center gap-4">
						<div class="w-12 h-12 rounded-xl bg-gray-500/10 flex items-center justify-center group-hover:bg-gray-500/20 transition-colors">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-gray-500" viewBox="0 0 20 20" fill="currentColor">
								<path fill-rule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clip-rule="evenodd" />
							</svg>
						</div>
						<div>
							<h3 class="font-semibold">Settings</h3>
							<p class="text-sm text-muted-foreground">Configure site & SMTP</p>
						</div>
					</CardContent>
				</a>
			</Card>
		{/if}
	</div>
</div>
