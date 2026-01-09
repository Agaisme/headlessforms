<script lang="ts">
  import { onMount } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import { Card, CardContent } from "$lib/components/ui/card";
  import { Badge } from "$lib/components/ui/badge";
  import { Skeleton } from "$lib/components/ui/skeleton";
  import {
    Table,
    TableHeader,
    TableBody,
    TableRow,
    TableHead,
    TableCell,
  } from "$lib/components/ui/table";
  import { toast } from "$lib/stores/toast";
  import { auth } from "$lib/stores/auth";
  import UserModal from "$lib/components/UserModal.svelte";

  interface User {
    id: string;
    email: string;
    name?: string;
    role: string;
    created_at?: string;
  }

  let users: User[] = [];
  let loading = true;

  // Modal state
  let showUserModal = false;
  let editingUser: User | null = null;

  onMount(async () => {
    await loadUsers();
  });

  function openCreateModal() {
    editingUser = null;
    showUserModal = true;
  }

  function openEditModal(user: User) {
    editingUser = user;
    showUserModal = true;
  }

  function handleUserSuccess(_user: User) {
    loadUsers();
  }

  async function loadUsers() {
    loading = true;
    try {
      const token = auth.getToken();
      const res = await fetch("/api/v1/users", {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
      const json = await res.json();

      if (json.status === "success") {
        users = json.data?.users || [];
      } else {
        toast.error(json.message || "Failed to load users");
      }
    } catch (e) {
      toast.error("Failed to load users");
    } finally {
      loading = false;
    }
  }

  async function handleDeleteUser(userId: string, userEmail: string) {
    if (
      !confirm(
        `Are you sure you want to delete "${userEmail}"? This action cannot be undone.`
      )
    ) {
      return;
    }

    try {
      const token = auth.getToken();
      const res = await fetch(`/api/v1/users/${userId}`, {
        method: "DELETE",
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });

      const json = await res.json();
      if (json.status === "success") {
        toast.success("User deleted successfully");
        await loadUsers();
      } else {
        toast.error(json.message || "Failed to delete user");
      }
    } catch (e) {
      toast.error("Failed to delete user");
    }
  }

  function getRoleBadgeVariant(role: string) {
    if (role === "super_admin") return "default";
    if (role === "admin") return "secondary";
    return "outline";
  }

  function formatRole(role: string) {
    if (role === "super_admin") return "Super Admin";
    if (role === "admin") return "Admin";
    return "User";
  }
</script>

<div class="space-y-6">
  <Card>
    <!-- Header -->
    <div
      class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-5 border-b border-border"
    >
      <div>
        <h2 class="text-lg font-semibold">User Management</h2>
        <p class="text-sm text-muted-foreground mt-0.5">
          Manage users and their access roles
        </p>
      </div>
      <Button onclick={openCreateModal} class="gap-2 whitespace-nowrap">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-4 w-4"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            d="M8 9a3 3 0 100-6 3 3 0 000 6zM8 11a6 6 0 016 6H2a6 6 0 016-6zM16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z"
          />
        </svg>
        Add User
      </Button>
    </div>

    {#if loading}
      <div class="flex justify-center py-16">
        <div class="flex flex-col items-center gap-4">
          <Skeleton class="h-10 w-10 rounded-full" />
          <Skeleton class="h-4 w-32" />
        </div>
      </div>
    {:else if users.length === 0}
      <div class="text-center py-20">
        <div
          class="w-20 h-20 rounded-full bg-muted flex items-center justify-center mx-auto mb-4"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-10 w-10 text-muted-foreground/50"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.5"
              d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
            />
          </svg>
        </div>
        <h3 class="text-lg font-semibold mb-2">No users yet</h3>
        <p class="text-muted-foreground mb-6 max-w-sm mx-auto">
          Create your first user to start managing access.
        </p>
        <Button onclick={openCreateModal}>Add Your First User</Button>
      </div>
    {:else}
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>User</TableHead>
            <TableHead>Role</TableHead>
            <TableHead class="hidden md:table-cell">Created</TableHead>
            <TableHead class="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {#each users as user}
            <TableRow class="group">
              <TableCell>
                <div class="flex items-center gap-3">
                  <div
                    class="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary font-medium"
                  >
                    {user.name?.[0]?.toUpperCase() ||
                      user.email?.[0]?.toUpperCase() ||
                      "?"}
                  </div>
                  <div>
                    <p class="font-medium">
                      {user.name || "No name"}
                      {#if user.id === $auth.user?.id}
                        <span class="text-xs text-primary ml-1">(You)</span>
                      {/if}
                    </p>
                    <p class="text-sm text-muted-foreground">{user.email}</p>
                  </div>
                </div>
              </TableCell>
              <TableCell>
                <Badge variant={getRoleBadgeVariant(user.role)}>
                  {formatRole(user.role)}
                </Badge>
              </TableCell>
              <TableCell class="hidden md:table-cell text-muted-foreground">
                {user.created_at
                  ? new Date(user.created_at).toLocaleDateString()
                  : "-"}
              </TableCell>
              <TableCell class="text-right">
                <div
                  class="flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <Button
                    variant="ghost"
                    size="sm"
                    onclick={() => openEditModal(user)}
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      class="h-4 w-4"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z"
                      />
                    </svg>
                  </Button>
                  {#if user.id !== $auth.user?.id}
                    <Button
                      variant="ghost"
                      size="sm"
                      onclick={() => handleDeleteUser(user.id, user.email)}
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        class="h-4 w-4 text-destructive"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                      >
                        <path
                          fill-rule="evenodd"
                          d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z"
                          clip-rule="evenodd"
                        />
                      </svg>
                    </Button>
                  {/if}
                </div>
              </TableCell>
            </TableRow>
          {/each}
        </TableBody>
      </Table>
    {/if}
  </Card>
</div>

<!-- User Modal -->
<UserModal
  bind:isOpen={showUserModal}
  user={editingUser}
  onsuccess={handleUserSuccess}
/>
