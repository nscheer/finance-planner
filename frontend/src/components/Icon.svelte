<script lang="ts">
  /** Tiny inline icon set (stroke icons, 24px viewBox). */
  let { name, size = 16 }: { name: IconName; size?: number } = $props();

  export type IconName =
    | "plus" | "edit" | "trash" | "chevron" | "grip" | "download" | "upload"
    | "expand" | "collapse" | "close" | "info" | "check" | "alert" | "globe"
    | "pause" | "play" | "copy" | "note" | "search" | "keyboard" | "history" | "sparkles" | "target" | "sun" | "printer" | "command" | "checkbox";

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
    globe: "M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20ZM2 12h20M12 2a15 15 0 0 1 0 20M12 2a15 15 0 0 0 0 20",
    pause: "M8 5v14M16 5v14",
    play: "m7 4 13 8-13 8Z",
    copy: "M9 9h11v11H9zM5 15H4V4h11v1",
    note: "M14 3v5h5M6 3h8l5 5v13H6zM9 13h6M9 17h6",
    search: "M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16ZM21 21l-4.3-4.3",
    keyboard: "M3 6h18v12H3zM7 10h.01M11 10h.01M15 10h.01M7 14h10",
    history: "M3 12a9 9 0 1 0 3-6.7M3 3v5h5M12 7v5l3 3",
    sparkles: "M12 3l1.8 5.2L19 10l-5.2 1.8L12 17l-1.8-5.2L5 10l5.2-1.8ZM19 17l.7 2 2 .7-2 .7-.7 2-.7-2-2-.7 2-.7Z",
    target: "M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20ZM12 18a6 6 0 1 0 0-12 6 6 0 0 0 0 12ZM12 14a2 2 0 1 0 0-4 2 2 0 0 0 0 4Z",
    sun: "M12 16a4 4 0 1 0 0-8 4 4 0 0 0 0 8ZM12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4",
    printer: "M6 9V3h12v6M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2M6 14h12v7H6z",
    // terminal prompt: ">_" – the usual sign for a command palette
    command: "M4 17l6-5-6-5M12 19h8",
    checkbox: "M4 6a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2zM8 12l3 3 5-6",
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
