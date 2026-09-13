<script lang="ts">
  /**
   * 12-month strip: bars show what is due in each month, the step line shows
   * the balance the savings account holds at the end of the month (steady
   * state). Hovering a month shows its values below the chart.
   */
  import type { Stats } from "../lib/store.svelte";
  import { t, formatEuro, monthShort, monthName } from "../lib/i18n.svelte";

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

  /** Row indices of the first half year; the second half is +6. */
  const firstHalf = [0, 1, 2, 3, 4, 5];
  const dash = "–";
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
    <!-- Always two lines, so hovering never reflows the box. -->
    <div class="readout" aria-live="polite">
      <div class="line" title={t("tip.timelineDue")}>
        <span class="swatch due"></span>
        <span class="label">{t("stats.timelineDue")}{hovered !== null ? ` (${monthName(hovered + 1)})` : ""}</span>
        <span class="money value">{hovered !== null ? formatEuro(months[hovered].dueCents) : ""}</span>
      </div>
      <div class="line" title={t("tip.timelineSaved")}>
        <span class="swatch saved"></span>
        <span class="label">{t("stats.timelineSaved")}</span>
        <span class="money value">{hovered !== null ? formatEuro(months[hovered].savedCents) : ""}</span>
      </div>
    </div>
    <!-- Paper cannot be hovered, so the figures are listed. Two half years
         side by side keep the block short. -->
    <table class="print-values">
      <thead>
        <tr>
          <th scope="col">{t("stats.timelineMonth")}</th>
          <th scope="col" class="num">{t("stats.timelineDue")}</th>
          <th scope="col" class="num gap-after">{t("stats.timelineSaved")}</th>
          <th scope="col">{t("stats.timelineMonth")}</th>
          <th scope="col" class="num">{t("stats.timelineDue")}</th>
          <th scope="col" class="num">{t("stats.timelineSaved")}</th>
        </tr>
      </thead>
      <tbody>
        {#each firstHalf as i (i)}
          <tr>
            <th scope="row">{monthShort(i + 1)}</th>
            <td class="num">{months[i].dueCents > 0 ? formatEuro(months[i].dueCents) : dash}</td>
            <td class="num gap-after">{formatEuro(months[i].savedCents)}</td>
            <th scope="row">{monthShort(i + 7)}</th>
            <td class="num">{months[i + 6].dueCents > 0 ? formatEuro(months[i + 6].dueCents) : dash}</td>
            <td class="num">{formatEuro(months[i + 6].savedCents)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
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
    stroke: var(--savings);
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
    flex-direction: column;
    gap: 2px;
    font-size: 11.5px;
    color: var(--text-2);
  }
  .line {
    display: flex;
    align-items: center;
    gap: 6px;
    min-height: 18px;
  }
  .line .label {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .line .value {
    font-weight: 500;
    color: var(--text);
  }
  .swatch {
    flex-shrink: 0;
    width: 8px;
    height: 8px;
    border-radius: 2px;
  }
  .swatch.due {
    background: var(--spending);
  }
  .swatch.saved {
    background: var(--savings);
  }
  /* Only on paper: the numbers the chart shows on hover. */
  .print-values {
    display: none;
  }
  @media print {
    .print-values {
      display: table;
      width: 100%;
      margin-top: 6px;
      border-collapse: collapse;
      font-size: 8.5pt;
      break-inside: avoid;
    }
    .print-values th,
    .print-values td {
      padding: 2px 0;
      border-bottom: 1px solid var(--border);
      font-weight: 400;
      text-align: left;
      white-space: nowrap;
    }
    .print-values thead th {
      font-weight: 600;
      color: var(--text-2);
      border-bottom-color: var(--border-strong);
    }
    .print-values .num {
      text-align: right;
      font-variant-numeric: tabular-nums;
      padding-left: 10px;
    }
    .print-values .gap-after {
      padding-right: 22px;
    }
  }

  .empty {
    margin: 0;
    font-size: 12px;
    color: var(--text-3);
  }
</style>
