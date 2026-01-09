<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { Button } from "$lib/components/ui/button";
  import { Card, CardContent } from "$lib/components/ui/card";
  import { Input } from "$lib/components/ui/input";
  import { auth } from "$lib/stores/auth";
  import { toast } from "$lib/stores/toast";

  let loading = true;
  let saving = false;
  let testing = false;
  let testEmail = "";

  // Redirect if not super_admin
  $: if (!$auth.isLoading && $auth.user?.role !== "super_admin") {
    goto("/");
  }

  // Site settings
  let siteName = "";
  let siteUrl = "";
  let smtpHost = "";
  let smtpPort = "587";
  let smtpUser = "";
  let smtpPass = "";
  let smtpFrom = "";
  let smtpFromName = "";
  let smtpSecure = true;

  onMount(async () => {
    await loadSettings();
  });

  async function loadSettings() {
    loading = true;
    try {
      const token = auth.getToken();
      const res = await fetch("/api/v1/settings", {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
      const json = await res.json();

      if (json.status === "success") {
        const data = json.data;
        siteName = data.site_name || "Headless Forms";
        siteUrl = data.site_url || "";
        smtpHost = data.smtp_host || "";
        smtpPort = String(data.smtp_port || 587);
        smtpUser = data.smtp_user || "";
        smtpPass = data.smtp_password || "";
        smtpFrom = data.smtp_from || "";
        smtpFromName = data.smtp_from_name || "";
        smtpSecure = data.smtp_secure ?? true;
      }
    } catch (e) {
      console.error("Failed to load settings:", e);
    } finally {
      loading = false;
    }
  }

  async function handleSave() {
    saving = true;
    try {
      const token = auth.getToken();
      const res = await fetch("/api/v1/settings", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({
          site_name: siteName,
          site_url: siteUrl,
          smtp_host: smtpHost,
          smtp_port: parseInt(smtpPort, 10) || 587,
          smtp_user: smtpUser,
          smtp_password: smtpPass,
          smtp_from: smtpFrom,
          smtp_from_name: smtpFromName,
          smtp_secure: smtpSecure,
        }),
      });
      const json = await res.json();

      if (json.status === "success") {
        toast.success("Settings saved successfully");
      } else {
        toast.error(json.message || "Failed to save settings");
      }
    } catch (e) {
      toast.error("Failed to save settings");
    } finally {
      saving = false;
    }
  }

  async function handleTestSMTP() {
    if (!testEmail) {
      toast.error("Please enter an email address to test");
      return;
    }
    if (!smtpHost) {
      toast.error("Please configure SMTP settings first");
      return;
    }

    testing = true;
    try {
      const token = auth.getToken();
      const res = await fetch("/api/v1/settings/test-smtp", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({
          host: smtpHost,
          port: parseInt(smtpPort, 10) || 587,
          user: smtpUser,
          password: smtpPass,
          from: smtpFrom,
          test_to: testEmail,
          secure: smtpSecure,
        }),
      });
      const json = await res.json();

      if (json.status === "success") {
        toast.success(json.data?.message || "Test email sent successfully!");
      } else {
        toast.error(json.message || "SMTP test failed");
      }
    } catch (e) {
      toast.error("SMTP test request failed");
    } finally {
      testing = false;
    }
  }
</script>

<div class="max-w-4xl mx-auto space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-bold">Site Settings</h1>
      <p class="text-muted-foreground">Configure your HeadlessForms instance</p>
    </div>
  </div>

  {#if loading}
    <div class="flex justify-center py-16">
      <div class="loading loading-spinner loading-lg"></div>
    </div>
  {:else}
    <!-- General Settings -->
    <Card>
      <CardContent class="p-6">
        <h2 class="text-lg font-semibold mb-4">General</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="form-control">
            <label class="label" for="site-name">
              <span class="label-text font-medium">Site Name</span>
            </label>
            <Input
              id="site-name"
              type="text"
              bind:value={siteName}
              placeholder="Headless Forms"
            />
          </div>
          <div class="form-control">
            <label class="label" for="site-url">
              <span class="label-text font-medium">Site URL</span>
            </label>
            <Input
              id="site-url"
              type="url"
              bind:value={siteUrl}
              placeholder="https://forms.example.com"
            />
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Email Settings -->
    <Card>
      <CardContent class="p-6">
        <h2 class="text-lg font-semibold mb-4">Email Configuration (SMTP)</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="form-control">
            <label class="label" for="smtp-host">
              <span class="label-text font-medium">SMTP Host</span>
            </label>
            <Input
              id="smtp-host"
              type="text"
              bind:value={smtpHost}
              placeholder="smtp.example.com"
            />
          </div>
          <div class="form-control">
            <label class="label" for="smtp-port">
              <span class="label-text font-medium">SMTP Port</span>
            </label>
            <Input
              id="smtp-port"
              type="number"
              bind:value={smtpPort}
              placeholder="587"
            />
          </div>
          <div class="form-control">
            <label class="label" for="smtp-user">
              <span class="label-text font-medium">SMTP Username</span>
            </label>
            <Input
              id="smtp-user"
              type="text"
              bind:value={smtpUser}
              placeholder="user@example.com"
            />
          </div>
          <div class="form-control">
            <label class="label" for="smtp-pass">
              <span class="label-text font-medium">SMTP Password</span>
            </label>
            <Input
              id="smtp-pass"
              type="password"
              bind:value={smtpPass}
              placeholder="••••••••"
            />
          </div>
          <div class="form-control">
            <label class="label" for="smtp-from">
              <span class="label-text font-medium">From Address</span>
            </label>
            <Input
              id="smtp-from"
              type="email"
              bind:value={smtpFrom}
              placeholder="noreply@example.com"
            />
          </div>
          <div class="form-control">
            <label class="label" for="smtp-from-name">
              <span class="label-text font-medium">From Name</span>
            </label>
            <Input
              id="smtp-from-name"
              type="text"
              bind:value={smtpFromName}
              placeholder="Headless Forms"
            />
          </div>
          <div class="form-control flex items-center gap-2 md:col-span-2">
            <input
              type="checkbox"
              id="smtp-secure"
              bind:checked={smtpSecure}
              class="checkbox checkbox-primary"
            />
            <label for="smtp-secure" class="label cursor-pointer">
              <span class="label-text">Use TLS/SSL</span>
            </label>
          </div>
        </div>

        <!-- SMTP Test Section -->
        <div class="mt-6 pt-6 border-t">
          <h3 class="font-medium mb-3">Test SMTP Configuration</h3>
          <div class="flex flex-col sm:flex-row gap-3">
            <Input
              type="email"
              bind:value={testEmail}
              placeholder="Enter email to send test to..."
              class="flex-1"
            />
            <Button
              onclick={handleTestSMTP}
              disabled={testing || !smtpHost}
              variant="outline"
              class="gap-2"
            >
              {#if testing}
                <span class="loading loading-spinner loading-sm"></span>
              {:else}
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  class="h-4 w-4"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    d="M2.003 5.884L10 9.882l7.997-3.998A2 2 0 0016 4H4a2 2 0 00-1.997 1.884z"
                  />
                  <path
                    d="M18 8.118l-8 4-8-4V14a2 2 0 002 2h12a2 2 0 002-2V8.118z"
                  />
                </svg>
              {/if}
              Send Test Email
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- System Info -->
    <Card>
      <CardContent class="p-6">
        <h2 class="text-lg font-semibold mb-4">System Information</h2>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
          <div>
            <p class="text-muted-foreground">Version</p>
            <p class="font-medium">v1.0.0</p>
          </div>
          <div>
            <p class="text-muted-foreground">Database</p>
            <p class="font-medium">SQLite</p>
          </div>
          <div>
            <p class="text-muted-foreground">Your Role</p>
            <p class="font-medium text-primary">{$auth.user?.role}</p>
          </div>
          <div>
            <p class="text-muted-foreground">Your Email</p>
            <p class="font-medium">{$auth.user?.email}</p>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Save Button -->
    <div class="flex justify-end">
      <Button onclick={handleSave} disabled={saving} class="gap-2">
        {#if saving}
          <span class="loading loading-spinner loading-sm"></span>
        {/if}
        Save Settings
      </Button>
    </div>
  {/if}
</div>
