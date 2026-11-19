<script>
  import { get } from '../lib/api.js';
  import { store } from '../lib/store.svelte.js';
  import { connection, onReconnect } from '../lib/connection.svelte.js';
  import { push } from '../lib/push.svelte.js';
  import { clock } from '../lib/clock.svelte.js';
  import { LINES } from '../lib/lines.js';
  import { ago } from '../lib/format.js';
  import { groupBoard } from '../lib/board.js';
  import DirectionCard from './DirectionCard.svelte';
  import Icon from './Icon.svelte';
  import LineDots from './LineDots.svelte';

  const REFRESH = 20000;
  // Past this, counting down from the last times would be a guess, not a time.
  const TOO_OLD = 10 * 60 * 1000;

  let { station } = $props();

  // Start from the last times seen here, if any, while the new ones load.
  // App.svelte keys this component by station, so this runs once per station.
  const saved = (() => store.lastBoard(station.code))();

  let entries = $state.raw(saved?.entries ?? null);
  let fetchedAt = $state(saved?.fetchedAt ?? 0);
  let failed = $state(false);
  let loading = $state(false);
  // True once this visit has had an answer of its own.
  let fresh = $state(false);

  const age = $derived(clock.now - fetchedAt);
  const tooOld = $derived(entries !== null && age > TOO_OLD);

  const board = $derived(entries && groupBoard(entries, station, store.lines, store.destinations));
  const count = $derived(
    board ? board.groups.reduce((n, g) => n + g.cards.length, board.unknown.length) : 0,
  );
  const favourite = $derived(store.isFavourite(station.code));

  async function load() {
    loading = true;
    try {
      const data = await get(`${station.code}/trains`);
      entries = Array.isArray(data) ? data : [];
      fetchedAt = Date.now();
      failed = false;
      fresh = true;
      store.saveBoard(station.code, entries, fetchedAt);
    } catch {
      failed = true;
    } finally {
      loading = false;
    }
  }

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
</script>

<header class="head">
  <a class="back" href="#/"><Icon name="back" size={18} /> Linhas</a>
  <div class="title">
    <p class="lines">
      <LineDots lines={station.lines} />
      {station.lines.map((id) => `Linha ${LINES[id].name}`).join(' · ')}
    </p>
    <h1>{station.name}</h1>
  </div>
  <div class="actions">
    <button
      class="icon-btn"
      aria-pressed={favourite}
      aria-label={favourite ? 'Remover das tuas estações' : 'Guardar nas tuas estações'}
      title={favourite ? 'Remover das tuas estações' : 'Guardar nas tuas estações'}
      onclick={() => store.toggleFavourite(station.code)}
    >
      <Icon name="star" class={favourite ? 'filled' : ''} />
    </button>
    <button
      class="icon-btn"
      aria-label="Atualizar tempos"
      title="Atualizar tempos"
      disabled={loading}
      onclick={load}
    >
      <Icon name="refresh" class={loading ? 'spin' : ''} />
    </button>
  </div>
</header>

{#if entries === null && failed}
  <div class="message">
    <h2>Ainda não temos tempos desta estação.</h2>
    {#if connection.state === 'offline'}
      <p>Estás offline e não há tempos guardados daqui. Aparecem assim que a rede voltar.</p>
    {:else}
      <p>O serviço do Metro não está a responder. Os tempos aparecem assim que responder.</p>
    {/if}
    <button class="btn btn--primary" onclick={load}>Tentar outra vez</button>
  </div>
{:else if tooOld}
  <div class="message">
    <h2>Os tempos guardados já não servem.</h2>
    <p>
      São de {ago(age)}, demasiado antigos para contar a partir deles.
      {connection.state === 'online' ? 'A atualizar…' : 'Atualizam assim que houver ligação.'}
    </p>
    <button class="btn btn--primary" onclick={load} disabled={loading}>Tentar outra vez</button>
  </div>
{:else if entries === null}
  <div class="cards skeleton" aria-busy="true" aria-label="A carregar tempos">
    <div></div>
    <div></div>
  </div>
{:else}
  <p class="updated" class:stale={failed}>
    {#if failed}
      <Icon name="alert" size={16} /> Sem ligação. Tempos de {ago(age)}, a contar desde então.
    {:else if !fresh}
      Tempos guardados de {ago(age)} · a atualizar…
    {:else}
      Atualizado {ago(clock.now - fetchedAt)}<span class="wide">&nbsp;· atualiza sozinho a cada 20 s</span>
    {/if}
  </p>

  {#if push.error}<p class="push-error" role="alert">{push.error}</p>{/if}

  {#if count === 0}
    <div class="message">
      <h2>Sem comboios previstos.</h2>
      <p>O metro circula entre as 6h30 e a 1h00. Fora desse horário não há tempos de espera.</p>
    </div>
  {:else}
    {#each board.groups as group (group.id)}
      {#if group.cards.length}
        <section class="group" data-line={group.id}>
          <div class="group__head">
            <h2 class="group__title"><span class="dot"></span>Linha {LINES[group.id].name}</h2>
            <a class="group__map" href="#/linha/{LINES[group.id].slug}">
              <Icon name="map" size={16} /> Ver na linha
            </a>
          </div>
          <div class="cards">
            {#each group.cards as card (card.key)}
              <DirectionCard {card} {fetchedAt} station={station.code} />
            {/each}
          </div>
        </section>
      {/if}
    {/each}
    {#if board.unknown.length}
      <section class="group">
        <h2 class="group__title">Outros destinos</h2>
        <div class="cards">
          {#each board.unknown as card (card.key)}
            <DirectionCard {card} {fetchedAt} station={station.code} />
          {/each}
        </div>
      </section>
    {/if}
  {/if}
{/if}

<style>
  .head {
    display: grid;
    grid-template-columns: 1fr auto;
    align-items: end;
    gap: 16px;
  }

  .back {
    display: none;
  }

  .lines {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 14px;
    font-weight: 500;
    color: var(--muted);
  }

  h1 {
    margin-top: 6px;
    font-size: clamp(32px, 4vw, 48px);
    font-weight: 700;
    line-height: 1.05;
    letter-spacing: -0.03em;
  }

  .actions {
    display: flex;
    gap: 8px;
  }

  .actions :global(.filled) {
    fill: currentColor;
  }

  .actions :global(.spin) {
    animation: spin 0.9s linear infinite;
  }

  .updated {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 12px 0 28px;
    font-size: 14px;
    color: var(--muted);
  }

  .updated.stale {
    color: var(--warn);
  }

  .push-error {
    margin: -12px 0 20px;
    font-size: 14px;
    font-weight: 500;
    color: var(--alert);
  }

  .group + .group {
    margin-top: 32px;
  }

  .group__head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 8px;
  }

  .group__title {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 17px;
    font-weight: 600;
  }

  .group__map {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-height: 44px;
    font-size: 15px;
    font-weight: 500;
    color: var(--ink-2);
    text-decoration: none;
  }

  @media (hover: hover) and (pointer: fine) {
    .group__map:hover {
      color: var(--ink);
      text-decoration: underline;
      text-underline-offset: 3px;
    }
  }

  .cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr));
    gap: 16px;
  }

  .skeleton {
    margin-top: 48px;
  }

  .skeleton div {
    height: 220px;
    border-radius: var(--r-lg);
    background: var(--sunken);
    opacity: 0;
    animation: appear 200ms 300ms var(--ease) forwards;
  }

  .message {
    display: grid;
    justify-items: start;
    gap: 8px;
    max-width: 560px;
    margin-top: 32px;
    padding: 24px;
    border: 1px solid var(--hair);
    border-radius: var(--r-lg);
    background: var(--surface);
  }

  .message h2 {
    font-size: 20px;
    font-weight: 600;
  }

  .message p {
    color: var(--ink-2);
  }

  .message .btn {
    margin-top: 8px;
  }

  @media (max-width: 819px) {
    .head {
      grid-template-columns: 1fr auto;
      gap: 12px;
    }

    h1 {
      font-size: 30px;
    }

    .updated {
      margin: 8px 0 20px;
    }

    .wide {
      display: none;
    }

    .group + .group {
      margin-top: 24px;
    }

    .cards {
      gap: 12px;
    }

    .back {
      display: inline-flex;
      grid-column: 1 / -1;
      align-items: center;
      gap: 4px;
      justify-self: start;
      min-height: 44px;
      margin-left: -6px;
      padding-right: 8px;
      font-weight: 500;
      text-decoration: none;
    }
  }

  @keyframes spin {
    to { rotate: 360deg; }
  }

  @keyframes appear {
    to { opacity: 1; }
  }
</style>
