<script>
  import { store } from '../lib/store.svelte.js';
  import { LINES, lineStatus } from '../lib/lines.js';
</script>

<section aria-labelledby="lines-title">
  <h2 id="lines-title" class="label">Linhas</h2>
  <div class="lines">
    {#each store.lines as line (line.id)}
      {@const status = lineStatus(store.status, store.statusFailed, line.id)}
      <button
        class="line"
        data-line={line.id}
        aria-pressed={store.line === line.id}
        onclick={() => store.setLine(line.id)}
      >
        <span class="line__bar" aria-hidden="true"></span>
        <span class="line__name">{LINES[line.id].name}</span>
        <span class="line__status" data-state={status.state} title={status.label}>
          {status.label}
        </span>
      </button>
    {/each}
  </div>
</section>

<style>
  .lines {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
    margin-top: 8px;
  }

  .line {
    position: relative;
    display: grid;
    gap: 2px;
    min-height: 72px;
    padding: 12px 14px 12px 22px;
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface);
    text-align: left;
    cursor: pointer;
    transition:
      border-color var(--t-out) var(--ease),
      transform var(--t-out) var(--ease);
  }

  .line__bar {
    position: absolute;
    left: 10px;
    top: 12px;
    bottom: 12px;
    width: 4px;
    border-radius: 2px;
    background: var(--c);
  }

  .line__name {
    font-size: 17px;
    font-weight: 600;
    letter-spacing: -0.01em;
  }

  .line__status {
    display: flex;
    align-items: center;
    gap: 6px;
    overflow: hidden;
    font-size: 13px;
    font-weight: 500;
    color: var(--muted);
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .line__status::before {
    content: '';
    width: 7px;
    height: 7px;
    flex: none;
    border-radius: 50%;
    background: var(--hair-strong);
  }

  .line__status[data-state='ok']::before { background: var(--ok); }
  .line__status[data-state='warn'] { color: var(--warn); }
  .line__status[data-state='warn']::before { background: var(--warn); }
  .line__status[data-state='alert'] { color: var(--alert); }
  .line__status[data-state='alert']::before { background: var(--alert); }

  .line[aria-pressed='true'] {
    border-color: var(--c);
    box-shadow: inset 0 0 0 1px var(--c);
  }

  @media (hover: hover) and (pointer: fine) {
    .line:hover {
      border-color: var(--hair-strong);
      transform: translateY(-1px);
      transition-duration: var(--t-in);
    }

    .line[aria-pressed='true']:hover {
      border-color: var(--c);
    }
  }

  .line:active {
    transform: translateY(1px) scale(0.985);
  }
</style>
