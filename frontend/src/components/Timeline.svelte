<script lang="ts">
  /**
   * 12-month strip: bars show what is due in each month, the step line shows
   * the balance the savings account holds at the end of the month (steady
   * state). Hovering a month shows its values below the chart.
   */
  import type { Stats } from "../lib/store.svelte";
  import { t, formatEuro, monthShort } from "../lib/i18n.svelte";

  let { stats }: { stats: Stats } = $props();

  const months = $derived(stats.timeline ?? []);
  const maxValue = $derived(Math.max(1, ...months.map((m) => Math.max(m.dueCents, m.savedCents))));
  const hasData = $derived(months.some((m) => m.dueCents > 0 || m.savedCents > 0));

  // Geometry (viewBox units).
  const W = 288;
  const H = 96;
  const PAD_TOP = 8;
  const PAD_BOTTOM = 18;
  const plotH = H - PAD_TOP - PAD_BOTTOM;
  const colW = W / 12;
  const barW = colW * 0.55;

  const y = (cents: number) => PAD_TOP + plotH - (cents / maxValue) * plotH;

  /** Step line through the month-end balances. */
  const linePath = $derived.by(() => {
    if (!hasData) return "";
    let d = "";
    months.forEach((m, i) => {
      const x0 = i * colW;
      const x1 = x0 + colW;
      const yy = y(m.savedCents);
      d += (i === 0 ? `M${x0},${yy}` : `L${x0},${yy}`) + ` L${x1},${yy}`;
    });
    return d;
  });

  let hovered = $state<number | null>(null);
  const currentMonth = new Date().getMonth();
</script>

<div class="timeline">
  {#if hasData}
    <svg viewBox="0 0 {W} {H}" width="100%" role="img" aria-label={t("stats.timeline")}>
      <line class="baseline" x1="0" x2={W} y1={PAD_TOP + plotH} y2={PAD_TOP + plotH} />
      {#each months as m, i (i)}
        {@const x = i * colW}
        <g class="col" class:hover={hovered === i} onmouseenter={() => (hovered = i)} onmouseleave={() => (hovered = null)} role="presentation">
          <rect class="hit" x={x} y="0" width={colW} height={H} />
          {#if m.dueCents > 0}
            <rect class="due" x={x + (colW - barW) / 2} y={y(m.dueCents)} width={barW} height={PAD_TOP + plotH - y(m.dueCents)} rx="2" />
          {/if}
          <text class="label" class:current={i === currentMonth} x={x + colW / 2} y={H - 5} text-anchor="middle">{monthShort(i + 1).slice(0, 1)}</text>
        </g>
      {/each}
      <path class="saved" d={linePath} />
    </svg>
    <div class="readout" aria-live="polite">
      {#if hovered !== null}
        <strong>{monthShort(hovered + 1)}</strong>
        <span><span class="swatch due"></span>{t("stats.timelineDue")}: <span class="money">{formatEuro(months[hovered].dueCents)}</span></span>
        <span><span class="swatch saved"></span>{t("stats.timelineSaved")}: <span class="money">{formatEuro(months[hovered].savedCents)}</span></span>
      {:else}
        <span><span class="swatch due"></span>{t("stats.timelineDue")}</span>
        <span><span class="swatch saved"></span>{t("stats.timelineSaved")}</span>
      {/if}
    </div>
  {:else}
    <p class="empty">{t("stats.timelineEmpty")}</p>
  {/if}
</div>

<style>
  .timeline {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  svg {
    display: block;
    overflow: visible;
  }
  .baseline {
    stroke: var(--border);
    stroke-width: 1;
  }
  .hit {
    fill: transparent;
  }
  .col.hover .hit {
    fill: var(--surface-3);
  }
  .due {
    fill: var(--spending);
  }
  .saved {
    fill: none;
    stroke: #6b3fa0;
    stroke-width: 2;
    stroke-linejoin: round;
  }
  .label {
    font-size: 8px;
    fill: var(--text-3);
  }
  .label.current {
    fill: var(--text);
    font-weight: 700;
  }
  .readout {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 12px;
    min-height: 18px;
    font-size: 11.5px;
    color: var(--text-2);
  }
  .swatch {
    display: inline-block;
    width: 8px;
    height: 8px;
    margin-right: 4px;
    border-radius: 2px;
    vertical-align: middle;
  }
  .swatch.due {
    background: var(--spending);
  }
  .swatch.saved {
    background: #6b3fa0;
  }
  .empty {
    margin: 0;
    font-size: 12px;
    color: var(--text-3);
  }
</style>
