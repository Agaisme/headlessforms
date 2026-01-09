<script lang="ts">
  import { page } from "$app/stores";
  import { Button } from "$lib/components/ui/button";
  import {
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    CardDescription,
  } from "$lib/components/ui/card";

  // Get error from page store
  $: error = $page.error;
  $: status = $page.status || 500;
</script>

<svelte:head>
  <title>Error {status} - HeadlessForms</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-background p-4">
  <div class="w-full max-w-md">
    <!-- Error Icon -->
    <div class="text-center mb-8">
      <div
        class="inline-flex items-center justify-center w-20 h-20 rounded-full bg-destructive/10 mb-4"
      >
        {#if status === 404}
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-10 w-10 text-destructive"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fill-rule="evenodd"
              d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
              clip-rule="evenodd"
            />
          </svg>
        {:else if status === 403}
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-10 w-10 text-destructive"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fill-rule="evenodd"
              d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z"
              clip-rule="evenodd"
            />
          </svg>
        {:else}
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-10 w-10 text-destructive"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fill-rule="evenodd"
              d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z"
              clip-rule="evenodd"
            />
          </svg>
        {/if}
      </div>
      <h1 class="text-4xl font-bold text-foreground mb-2">{status}</h1>
      <p class="text-muted-foreground">
        {#if status === 404}
          Page not found
        {:else if status === 403}
          Access denied
        {:else if status === 500}
          Internal server error
        {:else}
          Something went wrong
        {/if}
      </p>
    </div>

    <Card class="shadow-lg">
      <CardHeader class="text-center pb-2">
        <CardTitle class="text-lg">
          {#if status === 404}
            We couldn't find that page
          {:else if status === 403}
            You don't have access
          {:else}
            An error occurred
          {/if}
        </CardTitle>
        <CardDescription>
          {#if error?.message}
            {error.message}
          {:else if status === 404}
            The page you're looking for doesn't exist or has been moved.
          {:else if status === 403}
            You don't have permission to access this resource.
          {:else}
            Something unexpected happened. Please try again.
          {/if}
        </CardDescription>
      </CardHeader>
      <CardContent class="pt-4 flex flex-col gap-3">
        <Button href="/" class="w-full">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-4 w-4 mr-2"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z"
            />
          </svg>
          Go to Dashboard
        </Button>
        <Button variant="outline" onclick={() => history.back()} class="w-full">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-4 w-4 mr-2"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fill-rule="evenodd"
              d="M9.707 16.707a1 1 0 01-1.414 0l-6-6a1 1 0 010-1.414l6-6a1 1 0 011.414 1.414L5.414 9H17a1 1 0 110 2H5.414l4.293 4.293a1 1 0 010 1.414z"
              clip-rule="evenodd"
            />
          </svg>
          Go Back
        </Button>
      </CardContent>
    </Card>

    <!-- Debug info in development -->
    {#if import.meta.env.DEV && error?.message}
      <details class="mt-6 text-xs">
        <summary
          class="text-muted-foreground cursor-pointer hover:text-foreground"
        >
          Technical details
        </summary>
        <pre
          class="mt-2 p-4 bg-muted rounded-lg overflow-auto max-h-48 text-muted-foreground">
Error: {error.message}
Status: {status}
				</pre>
      </details>
    {/if}
  </div>
</div>
