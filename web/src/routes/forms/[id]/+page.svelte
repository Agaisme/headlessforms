<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import { Card, CardContent } from "$lib/components/ui/card";
  import { Input } from "$lib/components/ui/input";
  import { Badge } from "$lib/components/ui/badge";
  import { Skeleton } from "$lib/components/ui/skeleton";
  import { auth } from "$lib/stores/auth";
  import { toast } from "$lib/stores/toast";

  let formId = "";
  let form: any = null;
  let submissions: any[] = [];
  let loading = true;
  let error: string | null = null;
  let selectedSubmission: any = null;
  let showTestForm = false;

  // Test form state
  let testEmail = "";
  let submitting = false;
  let submitSuccess = false;

  $: formId = $page.params.id ?? "";

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    loading = true;
    error = null;
    try {
      const token = auth.getToken();
      const headers: HeadersInit = token
        ? { Authorization: `Bearer ${token}` }
        : {};
      const formRes = await fetch(`/api/v1/forms/${formId}`, { headers });
      const formJson = await formRes.json();
      if (formJson.status === "success") {
        form = formJson.data;
      } else {
        error = formJson.message || "Failed to load form";
      }

      const subRes = await fetch(`/api/v1/forms/${formId}/submissions`, {
        headers,
      });
      const subJson = await subRes.json();
      if (subJson.status === "success") {
        // Handle both paginated and non-paginated response
        const submissionList = subJson.data?.submissions || subJson.data || [];
        submissions = submissionList.sort(
          (a: any, b: any) =>
            new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        );
        if (submissions.length > 0 && !selectedSubmission) {
          selectedSubmission = submissions[0];
        }
      }
    } catch (e) {
      error = "Failed to load data";
    } finally {
      loading = false;
    }
  }

  async function handleTestSubmit() {
    submitting = true;
    submitSuccess = false;

    try {
      const token = auth.getToken();
      // Build data payload, including submission key for with_key forms
      const dataPayload: any = { email: testEmail };
      if (form?.access_mode === "with_key" && form?.submission_key) {
        dataPayload._submission_key = form.submission_key;
      }

      const res = await fetch(`/api/v1/submissions/${formId}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({
          data: dataPayload,
          meta: { source: "admin-test" },
        }),
      });

      if (res.ok) {
        submitSuccess = true;
        testEmail = "";
        await loadData();
        setTimeout(() => (submitSuccess = false), 3000);
      }
    } catch (e) {
      // Silently handle error
    } finally {
      submitting = false;
    }
  }

  async function markAsRead(subId: string) {
    const token = auth.getToken();
    await fetch(`/api/v1/submissions/${subId}/read`, {
      method: "PUT",
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    await loadData();
  }

  async function handleExportCSV() {
    try {
      const token = auth.getToken();
      const res = await fetch(`/api/v1/forms/${formId}/export/csv`, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
      if (res.ok) {
        const blob = await res.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `${form?.name || "submissions"}_export.csv`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
      }
    } catch (e) {
      console.error("Export failed:", e);
      toast.error("Failed to export CSV");
    }
  }

  async function markAsUnread(subId: string) {
    const token = auth.getToken();
    await fetch(`/api/v1/submissions/${subId}/unread`, {
      method: "PUT",
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    await loadData();
  }

  async function deleteSubmission(subId: string) {
    if (!confirm("Are you sure you want to delete this submission?")) return;
    const token = auth.getToken();
    await fetch(`/api/v1/submissions/${subId}`, {
      method: "DELETE",
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    selectedSubmission = null;
    await loadData();
  }

  function getEndpointUrl() {
    return `${window.location.origin}/api/v1/submissions/${formId}`;
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
  }

  function formatDate(dateStr: string) {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) {
      return date.toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
      });
    } else if (days === 1) {
      return "Yesterday";
    } else if (days < 7) {
      return date.toLocaleDateString("en-US", { weekday: "short" });
    } else {
      return date.toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
      });
    }
  }

  function formatFullDate(dateStr: string) {
    return new Date(dateStr).toLocaleString("en-US", {
      weekday: "long",
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function getSubmissionData(sub: any): Record<string, any> {
    if (!sub || !sub.data) return {};
    try {
      const parsed =
        typeof sub.data === "string" ? JSON.parse(sub.data) : sub.data;
      // JSON.parse can return null (valid JSON), so coalesce to empty object
      return parsed ?? {};
    } catch {
      return {};
    }
  }

  function getPreviewFields(sub: any) {
    const data = getSubmissionData(sub) || {};
    const email = data?.email || data?.Email || "";
    const name =
      data?.name || data?.firstName || data?.Name || data?.first_name || "";
    const subject =
      data?.subject ||
      data?.message?.substring(0, 50) ||
      data?.comments?.substring(0, 50) ||
      "Form Submission";
    return { email, name, subject };
  }

  function getSubmissionMeta(sub: any) {
    if (!sub?.meta) return null;
    try {
      const meta =
        typeof sub.meta === "string" ? JSON.parse(sub.meta) : sub.meta;
      const server = meta?._server || {};
      const spam = meta?._spam || {};

      // Parse user agent for device info
      const ua = server.user_agent || "";
      let device = "Unknown";
      if (ua.includes("Mobile") || ua.includes("Android")) device = "📱 Mobile";
      else if (ua.includes("iPhone") || ua.includes("iPad")) device = "📱 iOS";
      else if (ua.includes("Windows")) device = "💻 Windows";
      else if (ua.includes("Mac")) device = "💻 Mac";
      else if (ua.includes("Linux")) device = "💻 Linux";

      // Get browser
      let browser = "";
      if (ua.includes("Chrome") && !ua.includes("Edg")) browser = "Chrome";
      else if (ua.includes("Firefox")) browser = "Firefox";
      else if (ua.includes("Safari") && !ua.includes("Chrome"))
        browser = "Safari";
      else if (ua.includes("Edg")) browser = "Edge";

      // Country flag emoji
      const countryFlags: Record<string, string> = {
        ID: "🇮🇩",
        US: "🇺🇸",
        GB: "🇬🇧",
        JP: "🇯🇵",
        CN: "🇨🇳",
        DE: "🇩🇪",
        FR: "🇫🇷",
        AU: "🇦🇺",
        CA: "🇨🇦",
        SG: "🇸🇬",
        MY: "🇲🇾",
        TH: "🇹🇭",
        VN: "🇻🇳",
        PH: "🇵🇭",
        IN: "🇮🇳",
        KR: "🇰🇷",
        NL: "🇳🇱",
        BR: "🇧🇷",
        RU: "🇷🇺",
        IT: "🇮🇹",
        ES: "🇪🇸",
      };
      const flag = countryFlags[server.country] || "🌐";

      return {
        ip: server.ip || "",
        country: server.country || "",
        countryFlag: flag,
        timezone: server.estimated_tz || "",
        device,
        browser,
        referer: server.referer || "",
        origin: server.origin || "",
        spamScore: spam.score ?? 0,
        isSpam: spam.is_spam ?? false,
        spamFlags: spam.flags || [],
        timestamp: server.timestamp || "",
      };
    } catch {
      return null;
    }
  }

  function selectSubmission(sub: any) {
    selectedSubmission = sub;
    if (sub.status === "unread") {
      markAsRead(sub.id);
    }
  }

  function getUnreadCount() {
    return submissions.filter((s) => s.status === "unread").length;
  }
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <Button variant="ghost" size="sm" href="/forms" class="gap-2 mb-2">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-4 w-4"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            fill-rule="evenodd"
            d="M9.707 16.707a1 1 0 01-1.414 0l-6-6a1 1 0 010-1.414l6-6a1 1 0 011.414 1.414L5.414 9H17a1 1 0 110 2H5.414l4.293 4.293a1 1 0 010 1.414z"
            clip-rule="evenodd"
          />
        </svg>
        Back
      </Button>
      <h1 class="text-2xl font-bold">{form?.name || "Loading..."}</h1>
    </div>
    <div class="flex items-center gap-2">
      <Button
        variant="ghost"
        size="sm"
        class="gap-2"
        onclick={() => (showTestForm = !showTestForm)}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-4 w-4"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z"
            fill-rule="evenodd"
            clip-rule="evenodd"
          />
        </svg>
        Test
      </Button>
      <Button variant="ghost" size="sm" class="gap-2" onclick={handleExportCSV}>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-4 w-4"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            fill-rule="evenodd"
            d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z"
            clip-rule="evenodd"
          />
        </svg>
        Export
      </Button>
      <Button
        variant="outline"
        size="sm"
        class="gap-2"
        href="/forms/{formId}/edit"
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
        Edit Form
      </Button>
    </div>
  </div>

  {#if loading}
    <div class="flex flex-col items-center justify-center py-16 gap-4">
      <Skeleton class="h-10 w-10 rounded-full" />
      <Skeleton class="h-4 w-32" />
    </div>
  {:else if error}
    <div class="bg-destructive/10 text-destructive p-4 rounded-lg">
      {error}
    </div>
  {:else}
    <!-- Settings Panel (Collapsible) -->
    {#if showTestForm}
      <Card>
        <CardContent class="p-5">
          <div class="flex items-center justify-between mb-4">
            <h3 class="font-semibold">Form Settings & Test</h3>
            <Button
              variant="ghost"
              size="icon"
              onclick={() => (showTestForm = false)}
              aria-label="Close settings"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-5 w-5"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fill-rule="evenodd"
                  d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                  clip-rule="evenodd"
                />
              </svg>
            </Button>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <!-- Endpoint Info -->
            <div>
              <p class="text-sm font-medium mb-2">API Endpoint</p>
              <div class="flex items-center gap-2">
                <code
                  class="flex-1 bg-muted px-3 py-2 rounded text-xs font-mono truncate"
                  >POST {getEndpointUrl()}</code
                >
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={() => copyToClipboard(getEndpointUrl())}
                  title="Copy URL"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    class="h-4 w-4"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
                    <path
                      d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z"
                    />
                  </svg>
                </Button>
              </div>
              <p class="text-xs text-muted-foreground mt-2">
                Status: <Badge variant="success"
                  >{form?.status || "active"}</Badge
                >
              </p>
            </div>

            <!-- Quick Test -->
            <div>
              <p class="text-sm font-medium mb-2">Quick Test</p>
              {#if submitSuccess}
                <div
                  class="bg-green-500/10 text-green-600 p-3 rounded-lg text-sm"
                >
                  ✅ Sent!
                </div>
              {:else}
                <form
                  onsubmit={(e) => {
                    e.preventDefault();
                    handleTestSubmit();
                  }}
                  class="flex gap-2"
                >
                  <Input
                    type="email"
                    class="flex-1"
                    placeholder="test@email.com"
                    bind:value={testEmail}
                    required
                  />
                  <Button type="submit" size="sm" disabled={submitting}>
                    {submitting ? "..." : "Test"}
                  </Button>
                </form>
              {/if}
            </div>
          </div>
        </CardContent>
      </Card>
    {/if}

    <!-- Inbox Section -->
    <Card class="overflow-hidden">
      <!-- Inbox Header -->
      <div
        class="flex items-center justify-between px-5 py-4 border-b border-border"
      >
        <div class="flex items-center gap-3">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-5 w-5 text-primary"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              d="M2.003 5.884L10 9.882l7.997-3.998A2 2 0 0016 4H4a2 2 0 00-1.997 1.884z"
            />
            <path d="M18 8.118l-8 4-8-4V14a2 2 0 002 2h12a2 2 0 002-2V8.118z" />
          </svg>
          <h2 class="text-lg font-semibold">Inbox</h2>
          {#if getUnreadCount() > 0}
            <Badge>{getUnreadCount()} new</Badge>
          {/if}
        </div>
        <div class="flex items-center gap-2">
          {#if submissions.length > 0}
            <a
              href={`/api/v1/forms/${formId}/export/csv`}
              download
              class="inline-flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-muted rounded-md transition-colors"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-4 w-4"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fill-rule="evenodd"
                  d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z"
                  clip-rule="evenodd"
                />
              </svg>
              Export CSV
            </a>
          {/if}
          <span class="text-sm text-muted-foreground"
            >{submissions.length} total</span
          >
        </div>
      </div>

      {#if submissions.length === 0}
        <div class="text-center py-16">
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
                d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
              />
            </svg>
          </div>
          <h3 class="text-lg font-semibold mb-2">Inbox is empty</h3>
          <p class="text-muted-foreground mb-4">
            Submissions from your landing page will appear here.
          </p>
          <Button
            variant="outline"
            size="sm"
            onclick={() => (showTestForm = true)}>Send a Test Submission</Button
          >
        </div>
      {:else}
        <div class="flex min-h-[500px]">
          <!-- Email List (Left Panel) -->
          <div
            class="w-80 border-r border-border overflow-y-auto max-h-[600px]"
          >
            {#each submissions as sub}
              {@const preview = getPreviewFields(sub)}
              <button
                class="w-full text-left px-4 py-3 border-b border-border/50 hover:bg-muted/50 transition-colors
								       {selectedSubmission?.id === sub.id
                  ? 'bg-primary/5 border-l-2 border-l-primary'
                  : ''}
								       {sub.status === 'unread' ? 'bg-primary/5' : ''}"
                onclick={() => selectSubmission(sub)}
              >
                <div class="flex items-start justify-between gap-2">
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                      {#if sub.status === "unread"}
                        <div class="w-2 h-2 rounded-full bg-primary"></div>
                      {/if}
                      <p
                        class="font-semibold text-sm truncate {sub.status ===
                        'unread'
                          ? ''
                          : 'font-normal'}"
                      >
                        {preview.name || preview.email || "Anonymous"}
                      </p>
                    </div>
                    {#if preview.email && preview.name}
                      <p class="text-xs text-muted-foreground truncate pl-4">
                        {preview.email}
                      </p>
                    {/if}
                    <p class="text-sm text-muted-foreground truncate mt-1 pl-4">
                      {preview.subject}
                    </p>
                  </div>
                  <span class="text-xs text-muted-foreground whitespace-nowrap">
                    {formatDate(sub.created_at)}
                  </span>
                </div>
              </button>
            {/each}
          </div>

          <!-- Email Detail (Right Panel) -->
          <div class="flex-1 overflow-y-auto max-h-[600px]">
            {#if selectedSubmission}
              {@const data = getSubmissionData(selectedSubmission)}
              {@const preview = getPreviewFields(selectedSubmission)}

              <!-- Email Header -->
              <div
                class="px-6 py-4 border-b border-border bg-card sticky top-0"
              >
                <div class="flex items-start justify-between">
                  <div>
                    <h3 class="text-lg font-semibold">
                      {preview.name || preview.email || "Form Submission"}
                    </h3>
                    {#if preview.email}
                      <p class="text-sm text-muted-foreground">
                        {preview.email}
                      </p>
                    {/if}
                    <p class="text-xs text-muted-foreground mt-1">
                      {formatFullDate(selectedSubmission.created_at)}
                    </p>
                  </div>
                  <div class="flex items-center gap-1">
                    {#if selectedSubmission.status === "read"}
                      <Button
                        variant="ghost"
                        size="icon"
                        onclick={() => markAsUnread(selectedSubmission.id)}
                        title="Mark as unread"
                      >
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          class="h-4 w-4"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                          />
                        </svg>
                      </Button>
                    {:else}
                      <Button
                        variant="ghost"
                        size="icon"
                        onclick={() => markAsRead(selectedSubmission.id)}
                        title="Mark as read"
                      >
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          class="h-4 w-4"
                          viewBox="0 0 20 20"
                          fill="currentColor"
                        >
                          <path
                            fill-rule="evenodd"
                            d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                            clip-rule="evenodd"
                          />
                        </svg>
                      </Button>
                    {/if}
                    <Button
                      variant="ghost"
                      size="icon"
                      class="text-destructive"
                      onclick={() => deleteSubmission(selectedSubmission.id)}
                      title="Delete"
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        class="h-4 w-4"
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
                  </div>
                </div>
              </div>

              <!-- Submitter Info Card -->
              {@const meta = getSubmissionMeta(selectedSubmission)}
              {#if meta}
                <div class="px-6 py-3 bg-muted/50 border-b border-border">
                  <div
                    class="flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted-foreground"
                  >
                    {#if meta.country}
                      <span class="flex items-center gap-1" title="Location">
                        <span>{meta.countryFlag}</span>
                        <span>{meta.country}</span>
                        {#if meta.timezone}
                          <span class="text-muted-foreground/70"
                            >({meta.timezone})</span
                          >
                        {/if}
                      </span>
                    {/if}
                    {#if meta.device}
                      <span title="Device"
                        >{meta.device}{meta.browser
                          ? ` • ${meta.browser}`
                          : ""}</span
                      >
                    {/if}
                    {#if meta.ip}
                      <span class="font-mono" title="IP Address">{meta.ip}</span
                      >
                    {/if}
                    {#if meta.origin}
                      <span title="Origin">{meta.origin}</span>
                    {/if}
                  </div>
                  {#if meta.isSpam || meta.spamScore > 0}
                    <div class="mt-2 flex items-center gap-2">
                      {#if meta.isSpam}
                        <Badge variant="destructive" class="text-xs"
                          >⚠️ Spam ({meta.spamScore})</Badge
                        >
                      {:else if meta.spamScore > 30}
                        <Badge
                          variant="outline"
                          class="text-xs text-yellow-500 border-yellow-500"
                          >Suspicious ({meta.spamScore})</Badge
                        >
                      {:else if meta.spamScore > 0}
                        <Badge variant="outline" class="text-xs"
                          >Score: {meta.spamScore}</Badge
                        >
                      {/if}
                      {#if meta.spamFlags.length > 0}
                        <span class="text-xs text-muted-foreground">
                          {meta.spamFlags.join(", ")}
                        </span>
                      {/if}
                    </div>
                  {/if}
                </div>
              {/if}

              <!-- Email Body -->
              <div class="p-6">
                <div class="space-y-4">
                  {#each Object.entries(data) as [key, value]}
                    <div class="border-b border-border/50 pb-3">
                      <p
                        class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1"
                      >
                        {key
                          .replace(/([A-Z])/g, " $1")
                          .replace(/_/g, " ")
                          .trim()}
                      </p>
                      <div class="text-foreground">
                        {#if Array.isArray(value)}
                          <span class="flex flex-wrap gap-1.5">
                            {#each value as item}
                              <Badge variant="outline">{item}</Badge>
                            {/each}
                          </span>
                        {:else if typeof value === "object"}
                          <pre
                            class="text-xs bg-muted p-2 rounded mt-1">{JSON.stringify(
                              value,
                              null,
                              2
                            )}</pre>
                        {:else}
                          {value}
                        {/if}
                      </div>
                    </div>
                  {/each}
                </div>

                <!-- Actions -->
                <div class="mt-6 pt-4 border-t border-border flex gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    class="gap-2"
                    onclick={() =>
                      copyToClipboard(JSON.stringify(data, null, 2))}
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      class="h-4 w-4"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z"
                      />
                      <path
                        d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z"
                      />
                    </svg>
                    Copy JSON
                  </Button>
                </div>
              </div>
            {:else}
              <div
                class="flex items-center justify-center h-full text-muted-foreground"
              >
                <p>Select a submission to view details</p>
              </div>
            {/if}
          </div>
        </div>
      {/if}
    </Card>
  {/if}
</div>
