<script lang="ts">
  /**
   * Statistics box. Besides the plain saldo it shows the two monthly
   * transfers the planning approach is built on:
   *  - to the bank account: all spendings that are paid monthly
   *  - to the savings account: 1/12 of all spendings that are paid yearly,
   *    so the money is there when the yearly spending is due.
   */
  import type { Stats } from "../lib/store.svelte";
  import { formatEuro, formatEuroSigned } from "../lib/money";

  let { stats }: { stats: Stats } = $props();

  const sign = (cents: number) => (cents < 0 ? "negative" : cents > 0 ? "positive" : "");
</script>

<aside class="stats">
  <h2>Statistics</h2>

  <section class="group">
    <h3>Monthly transfers</h3>
    <div class="tile bank">
      <span class="label">To bank account</span>
      <span class="value money">{formatEuro(stats.toBankMonthlyCents)}</span>
      <span class="sub">spendings paid per month</span>
    </div>
    <div class="tile savings">
      <span class="label">To savings account</span>
      <span class="value money">{formatEuro(stats.toSavingsMonthlyCents)}</span>
      <span class="sub">1/12 of spendings paid per year</span>
    </div>
  </section>

  <section class="group">
    <h3>Overview</h3>
    <dl>
      <div>
        <dt>Income per month</dt>
        <dd class="money">{formatEuro(stats.incomeMonthlyCents)}</dd>
      </div>
      <div>
        <dt>Average cost per month</dt>
        <dd class="money">{formatEuro(stats.spendingMonthlyCents)}</dd>
      </div>
      <div class="total">
        <dt>Saldo per month</dt>
        <dd class="money {sign(stats.saldoMonthlyCents)}">{formatEuroSigned(stats.saldoMonthlyCents)}</dd>
      </div>
    </dl>
    <dl>
      <div>
        <dt>Income per year</dt>
        <dd class="money">{formatEuro(stats.incomeYearlyCents)}</dd>
      </div>
      <div>
        <dt>Cost per year</dt>
        <dd class="money">{formatEuro(stats.spendingYearlyCents)}</dd>
      </div>
      <div class="total">
        <dt>Saldo per year</dt>
        <dd class="money {sign(stats.saldoYearlyCents)}">{formatEuroSigned(stats.saldoYearlyCents)}</dd>
      </div>
    </dl>
  </section>

  <p class="note">
    Monthly spendings are paid from the bank account. Yearly spendings are saved up month by month on the
    savings account, so the money is available when they are due.
  </p>
</aside>

<style>
  .stats {
    position: sticky;
    top: 20px; /* matches the content padding, so it never jumps when sticking */
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
    border-left-color: #6b3fa0;
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
  .note {
    margin: 0;
    font-size: 12px;
    line-height: 1.5;
    color: var(--text-3);
  }
</style>
