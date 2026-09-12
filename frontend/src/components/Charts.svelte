<script lang="ts">
  /**
   * Donut of spending by category (largest first, at most 8 slots, the rest
   * folded into "Other") with a legend that names every slice. Inline SVG
   * using the categorical palette tokens.
   */
  import type { CategoryView } from "../lib/store.svelte";
  import { t, formatEuro, formatPercent } from "../lib/i18n.svelte";

  let { spending }: { spending: CategoryView[] } = $props();

  interface Slice {
    name: string;
    cents: number;
    color: string;
  }

  const slices = $derived.by((): Slice[] => {
    const sorted = spending.filter((c) => c.monthlyCents > 0).sort((a, b) => b.monthlyCents - a.monthlyCents);
    const top = sorted.slice(0, 8).map((c, i) => ({ name: c.name, cents: c.monthlyCents, color: `var(--series-${i + 1})` }));
    const rest = sorted.slice(8).reduce((sum, c) => sum + c.monthlyCents, 0);
    if (rest > 0) top.push({ name: t("stats.otherCategories"), cents: rest, color: "var(--series-other)" });
    return top;
  });
  const total = $derived(slices.reduce((sum, s) => sum + s.cents, 0));

  // Donut geometry: stroke-dasharray segments on a circle, with a 2px gap
  // between segments drawn by the surface-colored background circle.
  const R = 40;
  const C = 2 * Math.PI * R;
  const segments = $derived.by(() => {
    let offset = 0;
    return slices.map((s) => {
      const len = (s.cents / total) * C;
      const seg = { ...s, dash: `${Math.max(0, len - 2)} ${C - Math.max(0, len - 2)}`, offset: -offset };
      offset += len;
      return seg;
    });
  });

  let hovered = $state<number | null>(null);
</script>

<div class="charts">
  <h3>{t("stats.byCategory")}</h3>
  {#if slices.length === 0}
    <p class="empty">{t("stats.chartsEmpty")}</p>
  {:else}
    <div class="donut-row">
      <svg viewBox="0 0 100 100" width="104" height="104" role="img" aria-label={t("stats.byCategory")}>
        <circle cx="50" cy="50" r={R} fill="none" stroke="var(--border)" stroke-width="12" />
        {#each segments as seg, i (i)}
          <circle
            cx="50"
            cy="50"
            r={R}
            fill="none"
            stroke={seg.color}
            stroke-width={hovered === i ? 14 : 12}
            stroke-dasharray={seg.dash}
            stroke-dashoffset={seg.offset}
            transform="rotate(-90 50 50)"
            role="presentation"
            onmouseenter={() => (hovered = i)}
            onmouseleave={() => (hovered = null)}
          >
            <title>{seg.name}: {formatEuro(seg.cents)}</title>
          </circle>
        {/each}
        <text x="50" y="47" text-anchor="middle" class="center-label">{hovered !== null ? formatPercent(slices[hovered].cents / total) : "100%"}</text>
        <text x="50" y="59" text-anchor="middle" class="center-sub">{hovered !== null ? slices[hovered].name.slice(0, 12) : t("kind.spending")}</text>
      </svg>
      <ul class="legend">
        {#each slices as s, i (i)}
          <li class:hover={hovered === i} onmouseenter={() => (hovered = i)} onmouseleave={() => (hovered = null)}>
            <span class="swatch" style:background={s.color}></span>
            <span class="name">{s.name}</span>
            <span class="value money">{formatEuro(s.cents)}</span>
          </li>
        {/each}
      </ul>
    </div>
  {/if}

</div>

<style>
  .charts {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  h3 {
    margin: 6px 0 0;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--text-3);
  }
  .empty {
    margin: 0;
    font-size: 12px;
    color: var(--text-3);
  }
  .donut-row {
    display: flex;
    gap: 10px;
    align-items: flex-start;
  }
  svg {
    flex-shrink: 0;
  }
  circle {
    transition: stroke-width 0.1s;
  }
  .center-label {
    font-size: 13px;
    font-weight: 700;
    fill: var(--text);
  }
  .center-sub {
    font-size: 7px;
    fill: var(--text-3);
  }
  .legend {
    flex: 1;
    min-width: 0;
    margin: 0;
    padding: 0;
    list-style: none;
    font-size: 11.5px;
  }
  .legend li {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 2px 4px;
    border-radius: 4px;
  }
  .legend li.hover {
    background: var(--surface-3);
  }
  .swatch {
    flex-shrink: 0;
    width: 9px;
    height: 9px;
    border-radius: 2px;
  }
  .legend .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--text-2);
  }
  .legend .value {
    color: var(--text);
    font-weight: 500;
  }
</style>
