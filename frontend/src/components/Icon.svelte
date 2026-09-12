<script lang="ts">
  /** Tiny inline icon set (stroke icons, 24px viewBox). */
  let { name, size = 16 }: { name: IconName; size?: number } = $props();

  export type IconName =
    | "plus" | "edit" | "trash" | "chevron" | "grip" | "download" | "upload"
    | "expand" | "collapse" | "close" | "info" | "check" | "alert";

  const paths: Record<IconName, string> = {
    plus: "M12 5v14M5 12h14",
    edit: "M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z",
    trash: "M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6M10 11v6M14 11v6",
    chevron: "m9 6 6 6-6 6",
    grip: "M9 5h.01M9 12h.01M9 19h.01M15 5h.01M15 12h.01M15 19h.01",
    download: "M12 3v12M6 11l6 6 6-6M4 21h16",
    upload: "M12 21V9M6 13l6-6 6 6M4 3h16",
    expand: "M4 14v6h6M20 10V4h-6M20 4l-7 7M4 20l7-7",
    collapse: "M10 20v-6H4M14 4v6h6M20 4l-6 6M4 20l6-6",
    close: "M18 6 6 18M6 6l12 12",
    info: "M12 16v-4M12 8h.01M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20Z",
    check: "m5 12 5 5L20 7",
    alert: "M12 9v4M12 17h.01M10.3 3.9 1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0Z",
  };
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width={name === "grip" ? 3 : 2}
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"
>
  <path d={paths[name]} />
</svg>
