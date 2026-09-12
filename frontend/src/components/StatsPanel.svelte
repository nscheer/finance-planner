<script lang="ts">
  /**
   * Statistics box. Besides the plain saldo it shows the two monthly
   * transfers the planning approach is built on:
   *  - to the bank account: all spendings that are paid monthly
   *  - to the savings account: 1/12 of all spendings that are paid yearly,
   *    so the money is there when the yearly spending is due.
   */
  import Icon from "./Icon.svelte";
  import Timeline from "./Timeline.svelte";
  import Charts from "./Charts.svelte";
  import { type Stats, type CategoryView, type EntryView, Kind, openDialog } from "../lib/store.svelte";
  import { t, plural, formatEuro, formatEuroSigned, formatPercent } from "../lib/i18n.svelte";

  let { stats, spending }: { stats: Stats; spending: CategoryView[] } = $props();

  /** Opens the edit dialog for one of the "biggest levers". */
  function editLever(id: string) {
    let entry: EntryView | undefined;
    for (const c of spending) entry = entry ?? (c.entries ?? []).find((e) => e.id === id);
    if (entry) openDialog({ type: "entry", kind: Kind.KindSpending, entry });
  }

  const sign = (cents: number) => (cents < 0 ? "negative" : cents > 0 ? "positive" : "");
</script>

<aside class="stats">
  <h2>{t("stats.title")}</h2>

  <section class="group">
    <h3>{t("stats.transfers")}</h3>
    <div class="tile bank" title={t("tip.toBank")}>
      <span class="label">{t("stats.toBank")}</span>
      <span class="value money">{formatEuro(stats.toBankMonthlyCents)}</span>
      <span class="sub">{t("stats.toBankSub")}</span>
    </div>
    <div class="tile savings" title={t("tip.toSavings")}>
      <span class="label">{t("stats.toSavings")}</span>
      <span class="value money">{formatEuro(stats.toSavingsMonthlyCents)}</span>
      <span class="sub">{t("stats.toSavingsSub")}</span>
    </div>
  </section>

  <section class="group">
    <h3>{t("stats.overview")}</h3>
    <dl>
      <div title={t("tip.incomePerMonth")}>
        <dt>{t("stats.incomePerMonth")}</dt>
        <dd class="money">{formatEuro(stats.incomeMonthlyCents)}</dd>
      </div>
      <div title={t("tip.avgCostPerMonth")}>
        <dt>{t("stats.avgCostPerMonth")}</dt>
        <dd class="money">{formatEuro(stats.spendingMonthlyCents)}</dd>
      </div>
      <div class="total" title={t("tip.saldoPerMonth")}>
        <dt>{t("stats.saldoPerMonth")}</dt>
        <dd class="money {sign(stats.saldoMonthlyCents)}">{formatEuroSigned(stats.saldoMonthlyCents)}</dd>
      </div>
      <div class="goal" title={t("tip.goal")}>
        <dt>
          {t("stats.goal")}
          <button class="icon-btn small" type="button" title={t("stats.goalEdit")} aria-label={t("stats.goalEdit")} onclick={() => openDialog({ type: "goal" })}><Icon name="target" size={13} /></button>
        </dt>
        <dd class="money">{stats.savingsGoalCents > 0 ? formatEuro(stats.savingsGoalCents) : t("stats.goalNone")}</dd>
      </div>
      {#if stats.savingsGoalCents > 0}
        <div class="total" title={t("tip.remainingAfterGoal")}>
          <dt>{t("stats.remainingAfterGoal")}</dt>
          <dd class="money {sign(stats.remainingAfterGoalCents)}">{formatEuroSigned(stats.remainingAfterGoalCents)}</dd>
        </div>
      {/if}
    </dl>
    <dl>
      <div title={t("tip.incomePerYear")}>
        <dt>{t("stats.incomePerYear")}</dt>
        <dd class="money">{formatEuro(stats.incomeYearlyCents)}</dd>
      </div>
      <div title={t("tip.costPerYear")}>
        <dt>{t("stats.costPerYear")}</dt>
        <dd class="money">{formatEuro(stats.spendingYearlyCents)}</dd>
      </div>
      <div class="total" title={t("tip.saldoPerYear")}>
        <dt>{t("stats.saldoPerYear")}</dt>
        <dd class="money {sign(stats.saldoYearlyCents)}">{formatEuroSigned(stats.saldoYearlyCents)}</dd>
      </div>
    </dl>
  </section>

  <section class="group">
    <h3>{t("stats.timeline")}</h3>
    <Timeline {stats} />
    {#if stats.peakBufferCents > 0}
      <dl class="compact">
        <div class="total" title={t("tip.peakBuffer")}>
          <dt>{t("stats.peakBuffer")}</dt>
          <dd class="money">{formatEuro(stats.peakBufferCents)}</dd>
        </div>
      </dl>
    {/if}
    {#if stats.unscheduledCount > 0}
      <p class="note">{plural("stats.unscheduled", stats.unscheduledCount)}</p>
    {/if}
    {#if stats.pausedCount > 0}
      <p class="note">{plural("stats.paused", stats.pausedCount)}</p>
    {/if}
  </section>

  <section class="group">
    <h3 title={t("stats.leversHint")}>{t("stats.levers")}</h3>
    {#if (stats.topSpendings ?? []).length === 0}
      <p class="note">{t("stats.leversEmpty")}</p>
    {:else}
      <ol class="levers">
        {#each stats.topSpendings ?? [] as lever, i (lever.id)}
          {@const values = { yearly: formatEuro(lever.yearlyCents), monthly: formatEuro(lever.monthlyCents), share: formatPercent(lever.shareOfSpending), income: formatPercent(lever.shareOfIncome) }}
          <li>
            <button
              type="button"
              class="lever"
              title={lever.shareOfIncome > 0 ? t("tip.lever", values) : t("tip.leverNoIncome", values)}
              onclick={() => editLever(lever.id)}
            >
              <span class="rank">{i + 1}</span>
              <span class="lever-text">
                <span class="lever-name">{lever.name}</span>
                <span class="lever-cat">{lever.categoryName}</span>
              </span>
              <span class="lever-values">
                <span class="money lever-amount">{formatEuro(lever.yearlyCents)}</span>
                <span class="lever-share">{formatPercent(lever.shareOfSpending)}</span>
              </span>
            </button>
          </li>
        {/each}
      </ol>
    {/if}
  </section>

  <section class="group">
    <Charts {spending} />
  </section>

  <p class="note">{t("stats.note")}</p>
</aside>

<style>
  .stats {
    position: sticky;
    top: 20px;
    margin-top: 20px; /* same gap as the tables column; keeps both boxes level */
    display: flex;
    flex-direction: column;
    gap: 18px;
    padding: 18px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
  }
  h2 {
    font-size: 15px;
  }
  h3 {
    margin-bottom: 8px;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--text-3);
  }
  .group {
    display: flex;
    flex-direction: column;
  }
  .tile {
    display: flex;
    flex-direction: column;
    padding: 12px 14px;
    border-radius: var(--radius-sm);
    border-left: 4px solid var(--accent);
    background: var(--surface-2);
  }
  .tile + .tile {
    margin-top: 8px;
  }
  .tile.bank {
    border-left-color: var(--accent);
  }
  .tile.savings {
    border-left-color: var(--savings);
  }
  .label {
    font-size: 12.5px;
    font-weight: 500;
    color: var(--text-2);
  }
  .value {
    font-size: 22px;
    font-weight: 600;
    line-height: 1.3;
  }
  .sub {
    font-size: 11.5px;
    color: var(--text-3);
  }
  dl {
    margin: 0;
    padding: 0;
  }
  dl + dl {
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--border);
  }
  dl > div {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    padding: 3px 0;
  }
  dt {
    color: var(--text-2);
  }
  dd {
    margin: 0;
    font-weight: 500;
  }
  .total dt {
    font-weight: 600;
    color: var(--text);
  }
  .total dd {
    font-weight: 700;
    font-size: 15px;
  }
  .goal dt {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .icon-btn.small {
    width: 22px;
    height: 22px;
  }
  dl.compact {
    margin-top: 6px;
  }
  dl.compact .total dt {
    font-weight: 500;
    font-size: 12.5px;
  }
  dl.compact .total dd {
    font-size: 13px;
  }
  .levers {
    margin: 0;
    padding: 0;
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .lever {
    display: grid;
    grid-template-columns: 18px 1fr auto;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 4px 6px;
    border: 0;
    border-radius: var(--radius-sm);
    background: transparent;
    text-align: left;
    cursor: pointer;
  }
  .lever:hover {
    background: var(--surface-3);
  }
  .rank {
    font-size: 11px;
    font-weight: 700;
    color: var(--text-3);
    text-align: center;
  }
  .lever-text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .lever-name {
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .lever-cat {
    font-size: 11px;
    color: var(--text-3);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .lever-values {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
  }
  .lever-amount {
    font-weight: 600;
  }
  .lever-share {
    font-size: 11px;
    color: var(--text-3);
  }
  .note {
    margin: 0;
    font-size: 12px;
    line-height: 1.5;
    color: var(--text-3);
  }
</style>
