<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button';
	import Toast from '$lib/components/Toast.svelte';

	// Pages that don't require authentication
	const publicPaths = ['/login', '/register', '/forgot-password', '/reset-password'];
	// Auth pages that should redirect if already logged in
	const authOnlyPaths = ['/login', '/register', '/forgot-password', '/reset-password'];
	
	let isPublicPage = false;
	$: isPublicPage = publicPaths.includes($page.url.pathname);
	
	// Check auth state
	$: {
		if (!$auth.isLoading) {
			// Redirect unauthenticated users from protected pages
			if (!$auth.isAuthenticated && !isPublicPage) {
				goto('/login');
			}
			// Redirect authenticated users away from login/register pages
			if ($auth.isAuthenticated && authOnlyPaths.includes($page.url.pathname)) {
				goto('/');
			}
		}
	}

	function handleLogout() {
		auth.logout();
		goto('/login');
	}
</script>

{#if isPublicPage}
	<!-- Public pages (login/register) - no sidebar -->
	<slot />
{:else if $auth.isLoading}
	<!-- Loading state -->
	<div class="min-h-screen flex items-center justify-center bg-background">
		<div class="flex flex-col items-center gap-4">
			<div class="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-primary"></div>
			<p class="text-muted-foreground">Loading...</p>
		</div>
	</div>
{:else if $auth.isAuthenticated}
	<!-- Authenticated layout with sidebar -->
	<div class="flex min-h-screen bg-background">
		<!-- Sidebar -->
		<aside class="w-64 bg-sidebar border-r border-sidebar-border flex flex-col">
			<!-- Logo -->
			<div class="h-16 flex items-center px-6 border-b border-sidebar-border">
				<a href="/" class="flex items-center gap-2">
					<div class="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-primary-foreground" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M5 3a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2V5a2 2 0 00-2-2H5zm0 2h10v7h-2l-1 2H8l-1-2H5V5z" clip-rule="evenodd" />
						</svg>
					</div>
					<span class="text-lg font-bold text-sidebar-foreground">HeadlessForms</span>
				</a>
			</div>

			<!-- Navigation -->
			<nav class="flex-1 p-4 space-y-1">
				<a href="/" class="sidebar-link" class:active={$page.url.pathname === '/'}>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
						<path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
					</svg>
					<span>Dashboard</span>
				</a>
				<a href="/forms" class="sidebar-link" class:active={$page.url.pathname === '/forms' || $page.url.pathname.startsWith('/forms/')}>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
						<path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
						<path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clip-rule="evenodd" />
					</svg>
					<span>Forms</span>
				</a>
				{#if $auth.user?.role === 'admin' || $auth.user?.role === 'super_admin'}
					<a href="/users" class="sidebar-link" class:active={$page.url.pathname === '/users'}>
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
							<path d="M9 6a3 3 0 11-6 0 3 3 0 016 0zM17 6a3 3 0 11-6 0 3 3 0 016 0zM12.93 17c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z" />
						</svg>
						<span>Users</span>
					</a>
				{/if}
				{#if $auth.user?.role === 'super_admin'}
					<a href="/settings" class="sidebar-link" class:active={$page.url.pathname === '/settings'}>
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clip-rule="evenodd" />
						</svg>
						<span>Settings</span>
					</a>
				{/if}
				<a href="/profile" class="sidebar-link" class:active={$page.url.pathname === '/profile'}>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd" />
					</svg>
					<span>Profile</span>
				</a>
			</nav>

			<!-- User & Footer -->
			<div class="p-4 border-t border-sidebar-border">
				<div class="flex items-center gap-3 mb-3">
					<div class="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-primary font-medium text-sm">
						{$auth.user?.name?.[0]?.toUpperCase() || $auth.user?.email?.[0]?.toUpperCase() || '?'}
					</div>
					<div class="flex-1 min-w-0">
						<p class="text-sm font-medium text-sidebar-foreground truncate">{$auth.user?.name || 'User'}</p>
						<p class="text-xs text-sidebar-foreground/50 truncate">{$auth.user?.email}</p>
					</div>
				</div>
				<Button variant="ghost" size="sm" class="w-full justify-start gap-2 text-sidebar-foreground/70" onclick={handleLogout}>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M3 3a1 1 0 00-1 1v12a1 1 0 001 1h5a1 1 0 100-2H4V5h4a1 1 0 100-2H3zm11 4a1 1 0 10-1.414 1.414L14.586 10l-2 2a1 1 0 101.414 1.414l2.707-2.707a1 1 0 000-1.414l-2.707-2.707A1 1 0 0014 7z" clip-rule="evenodd" />
						<path fill-rule="evenodd" d="M10 10a1 1 0 011-1h6a1 1 0 110 2h-6a1 1 0 01-1-1z" clip-rule="evenodd" />
					</svg>
					Sign Out
				</Button>
				<div class="text-xs text-sidebar-foreground/50 mt-3 flex items-center justify-between">
					<span>v1.0.0</span>
					<span class="px-1.5 py-0.5 rounded bg-primary/10 text-primary text-[10px] font-medium">{$auth.user?.role}</span>
				</div>
			</div>
		</aside>

		<!-- Main Content -->
		<main class="flex-1 flex flex-col">
			<!-- Top Bar -->
			<header class="h-16 bg-card border-b border-border flex items-center justify-between px-6">
				<div class="flex items-center gap-4">
					<h1 class="text-lg font-semibold text-foreground">
						{#if $page.url.pathname === '/'}
							Dashboard
						{:else if $page.url.pathname === '/forms'}
							Forms
						{:else if $page.url.pathname.startsWith('/forms/')}
							Form Details
						{:else if $page.url.pathname === '/users'}
							Users
						{:else if $page.url.pathname === '/settings'}
							Settings
						{:else if $page.url.pathname === '/profile'}
							Profile
						{:else}
							HeadlessForms
						{/if}
					</h1>
				</div>
			</header>

			<!-- Page Content -->
			<div class="flex-1 p-6 overflow-auto bg-muted/30">
				<slot />
			</div>
		</main>
	</div>
{:else}
	<!-- Redirect to login (handled by reactive statement) -->
	<div class="min-h-screen flex items-center justify-center bg-background">
		<p class="text-muted-foreground">Redirecting to login...</p>
	</div>
{/if}

<!-- Global Toast Notifications -->
<Toast />
