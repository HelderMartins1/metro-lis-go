<script module>
  // The last positions per line, for this session: going back to a line shows them at once.
  const lastSeen = new Map();
</script>

<script>
  import { get } from '../lib/api.js';
  import { onReconnect } from '../lib/connection.svelte.js';
  import { store } from '../lib/store.svelte.js';
  import { clock } from '../lib/clock.svelte.js';
  import { LINES } from '../lib/lines.js';
  import { ago, clockTime } from '../lib/format.js';
  import Icon from './Icon.svelte';
  import LineDots from './LineDots.svelte';

  const REFRESH = 15000;
  // Trains drift to their next station and stop there; past this, hide them rather than mislead.
  const TOO_OLD = 3 * 60 * 1000;

  const ROW = 56; // px per station; the drawing below is laid out in these units

  let { lineId } = $props();

  const line = $derived(store.lines.find((l) => l.id === lineId));

  const saved = (() => lastSeen.get(lineId))();
  let trains = $state.raw(saved?.trains ?? null);
  let fetchedAt = $state(saved?.fetchedAt ?? 0);
  let failed = $state(false);
  let selected = $state(null);

  const tooOld = $derived(trains !== null && clock.now - fetchedAt > TOO_OLD);

  async function load() {
    try {
      const data = await get(`${lineId}/positions`);
      trains = data.trains ?? [];
      fetchedAt = Date.now();
      failed = false;
      lastSeen.set(lineId, { trains, fetchedAt });
    } catch {
      failed = true;
    }
  }

  // App.svelte keys this component by line, so this runs once per line.
  $effect(() => {
    load();
    const timer = setInterval(() => !document.hidden && load(), REFRESH);
    const onVisible = () => !document.hidden && Date.now() - fetchedAt > REFRESH / 2 && load();
    document.addEventListener('visibilitychange', onVisible);
    const stop = onReconnect(load);
    return () => {
      clearInterval(timer);
      document.removeEventListener('visibilitychange', onVisible);
      stop();
    };
  });

  // Where each train is now, in stations from the top of the drawing.
  // "up" trains run towards the last station (down the screen), "down" ones back up.
  const placed = $derived.by(() => {
    if (!trains || tooOld) return [];
    const elapsed = Math.max(0, (clock.now - fetchedAt) / 1000);
    return trains.map((t) => {
      const left = Math.max(0, t.nextSeconds - elapsed);
      const along = Math.min(left / Math.max(t.segmentSeconds, 1), 1);
      return {
        ...t,
        key: `${t.direction}-${t.id}`,
        left,
        position: t.direction === 'up' ? t.next - along : t.next + along,
      };
    });
  });

  const chosen = $derived(placed.find((t) => t.key === selected));
  const first = $derived(line?.stations[0].name);
  const last = $derived(line?.stations.at(-1).name);
</script>

{#if line}
  <div class="map" data-line={line.id}>
    <header class="head">
      <a class="back" href="#/"><Icon name="back" size={18} /> Linhas</a>
      <p class="kicker"><span class="dot"></span>Linha {LINES[line.id].name}</p>
      <h1>Comboios na linha</h1>
      <p class="meta">
        {#if trains === null && failed}
          Não conseguimos obter as posições.
        {:else if tooOld}
          <span class="stale">Posições de {ago(clock.now - fetchedAt)}: já não servem, voltam com a ligação.</span>
        {:else if trains === null}
          A calcular posições…
        {:else}
          {trains.length} {trains.length === 1 ? 'comboio' : 'comboios'} · atualizado {ago(clock.now - fetchedAt)}
          {#if failed}<span class="stale"> · sem ligação</span>{/if}
        {/if}
      </p>
    </header>

    <ul class="legend" aria-label="Sentidos">
      <li><span class="chip chip--down" aria-hidden="true"><Icon name="arrow-up" size={14} /></span>Sentido {first}</li>
      <li><span class="chip chip--up" aria-hidden="true"><Icon name="arrow-down" size={14} /></span>Sentido {last}</li>
    </ul>

    <div class="info" aria-live="polite">
      {#if chosen}
        <p>
          <strong>{chosen.left <= 0 ? 'Na estação' : `Chega em ${clockTime(chosen.left)}`}</strong>
          a {line.stations[chosen.next].name}
          <span class="info__sub">Comboio {chosen.id} · sentido {chosen.destination}</span>
        </p>
        <button class="icon-btn" aria-label="Fechar detalhes do comboio" onclick={() => (selected = null)}>
          <Icon name="close" />
        </button>
      {:else}
        <p class="info__hint">
          Toca num comboio para ver quando chega. A posição é aproximada, calculada pelos tempos de espera.
        </p>
      {/if}
    </div>

    {#if trains === null && failed}
      <div class="message">
        <p>O serviço do Metro pode estar em baixo. Tentamos outra vez daqui a pouco.</p>
        <button class="btn btn--primary" onclick={load}>Tentar outra vez</button>
      </div>
    {:else}
      <div class="diagram" style:--rows={line.stations.length} style:--row="{ROW}px">
        <span class="track" aria-hidden="true"></span>

        <ol class="stations">
          {#each line.stations as s, i (s.code)}
            {@const others = store.stations.get(s.code).lines.filter((id) => id !== line.id)}
            <li class:interchange={others.length > 0} class:terminus={i === 0 || i === line.stations.length - 1}>
              <span class="node" aria-hidden="true"></span>
              <a href="#/{s.code}">
                <span class="name">{s.name}</span>
                {#if others.length}<LineDots lines={others} />{/if}
              </a>
            </li>
          {/each}
        </ol>

        {#each placed as t (t.key)}
          <button
            class="train train--{t.direction}"
            class:is-selected={t.key === selected}
            style:--y={t.position}
            aria-label="Comboio {t.id}, sentido {t.destination}, chega a {line.stations[t.next].name} em {Math.ceil(t.left / 60)} min"
            aria-pressed={t.key === selected}
            onclick={() => (selected = selected === t.key ? null : t.key)}
          >
            <span class="train__body">
              <Icon name={t.direction === 'up' ? 'arrow-down' : 'arrow-up'} size={14} />
            </span>
          </button>
        {/each}
      </div>
    {/if}
  </div>
{/if}

<style>
  .head {
    display: grid;
    gap: 4px;
  }

  .back {
    display: none;
  }

  .kicker {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 500;
    color: var(--muted);
  }

  h1 {
    font-size: clamp(30px, 4vw, 44px);
    font-weight: 700;
    line-height: 1.05;
    letter-spacing: -0.03em;
  }

  .meta {
    font-size: 14px;
    color: var(--muted);
  }

  .stale {
    color: var(--warn);
  }

  .legend {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 20px;
    margin: 20px 0 0;
    padding: 0;
    list-style: none;
    font-size: 14px;
    font-weight: 500;
  }

  .legend li {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .chip {
    display: grid;
    place-items: center;
    width: 24px;
    height: 24px;
    border-radius: 50%;
  }

  /* One direction filled, the other outlined, so the two read apart without colour. */
  .chip--up,
  .train--up .train__body {
    background: var(--c);
    color: var(--on-c, #fff);
  }

  .chip--down,
  .train--down .train__body {
    border: 3px solid var(--c);
    background: var(--surface);
    color: var(--ink);
  }

  [data-line='yellow'] {
    --on-c: #17181a;
  }

  /* Stays in view while the diagram scrolls, so a chosen train's details never leave the screen. */
  .info {
    position: sticky;
    top: 8px;
    z-index: 5;
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 72px;
    margin-top: 16px;
    padding: 12px 12px 12px 16px;
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface);
  }

  .info p {
    flex: 1;
    line-height: 1.35;
  }

  .info strong {
    font-variant-numeric: tabular-nums;
  }

  .info__sub {
    display: block;
    margin-top: 2px;
    font-size: 14px;
    color: var(--muted);
  }

  .info__hint {
    font-size: 14px;
    color: var(--muted);
  }

  .message {
    display: grid;
    justify-items: start;
    gap: 12px;
    margin-top: 24px;
  }

  /* The drawing: two lanes either side of the line, station names to the right.
     x positions: down lane 22px, line 54px, up lane 86px, names from 116px. */
  .diagram {
    position: relative;
    height: calc(var(--rows) * var(--row));
    margin-top: 16px;
  }

  .track {
    position: absolute;
    left: 51px;
    top: calc(var(--row) / 2);
    bottom: calc(var(--row) / 2);
    width: 6px;
    border-radius: 3px;
    background: var(--c);
  }

  .stations {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .stations li {
    position: relative;
    height: var(--row);
  }

  .node {
    position: absolute;
    left: 46px;
    top: 50%;
    width: 16px;
    height: 16px;
    margin-top: -8px;
    border: 4px solid var(--c);
    border-radius: 50%;
    background: var(--bg);
  }

  .interchange .node {
    left: 44px;
    width: 20px;
    height: 20px;
    margin-top: -10px;
    border-color: var(--ink);
    background: var(--surface);
  }

  .stations a {
    display: flex;
    align-items: center;
    gap: 10px;
    height: 100%;
    margin-left: 116px;
    text-decoration: none;
  }

  .terminus .name {
    font-weight: 600;
  }

  @media (hover: hover) and (pointer: fine) {
    .stations a:hover .name {
      text-decoration: underline;
      text-underline-offset: 3px;
    }
  }

  /* Each train is a 44px tap target centred on its spot; only transform moves. */
  .train {
    position: absolute;
    top: 0;
    z-index: 3;
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    padding: 0;
    border: 0;
    background: none;
    cursor: pointer;
    transform: translateY(calc(var(--y) * var(--row) + var(--row) / 2 - 22px));
    transition: transform 1s linear;
  }

  .train--down {
    left: 0;
  }

  .train--up {
    left: 64px;
  }

  .train__body {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    box-shadow: 0 0 0 3px var(--bg);
    transition: scale var(--t-out) var(--ease);
  }

  .train.is-selected .train__body {
    scale: 1.25;
    box-shadow:
      0 0 0 3px var(--bg),
      0 0 0 5px var(--ink);
  }

  .train:focus-visible {
    outline-offset: -2px;
  }

  @media (max-width: 819px) {
    .back {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      justify-self: start;
      min-height: 44px;
      margin-left: -6px;
      padding-right: 8px;
      font-weight: 500;
      text-decoration: none;
    }

    h1 {
      font-size: 30px;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .train {
      transition: none;
    }
  }
</style>
