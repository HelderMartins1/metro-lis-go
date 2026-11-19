<script>
  import { store } from '../lib/store.svelte.js';
  import { LINES } from '../lib/lines.js';
  import Icon from './Icon.svelte';
  import LineDots from './LineDots.svelte';

  let { current = null } = $props();

  const line = $derived(store.lines.find((l) => l.id === store.line));
  let list = $state();

  // Bring the open station into view when it is further down the line.
  $effect(() => {
    if (!current || !line || !list) return;
    list.querySelector('[aria-current="page"]')?.scrollIntoView({ block: 'nearest' });
  });
</script>

{#if line}
  <section data-line={line.id} aria-labelledby="stations-title">
    <div class="list-head">
      <h2 id="stations-title" class="label">
        Linha {LINES[line.id].name} · {line.stations.length} estações
      </h2>
      <a class="btn map-link" href="#/linha/{LINES[line.id].slug}">
        <Icon name="map" size={18} /> Ver comboios
      </a>
    </div>
    <ol bind:this={list}>
      {#each line.stations as s, i (s.code)}
        {@const others = store.stations.get(s.code).lines.filter((id) => id !== line.id)}
        <li class:interchange={others.length > 0} class:terminus={i === 0 || i === line.stations.length - 1}>
          <a href="#/{s.code}" aria-current={s.code === current ? 'page' : undefined}>
            <span class="node" aria-hidden="true"></span>
            <span class="name">{s.name}</span>
            {#if others.length}<LineDots lines={others} />{/if}
            {#if store.isFavourite(s.code)}
              <span class="fav" title="Favorita"><Icon name="star" size={16} /></span>
            {/if}
          </a>
        </li>
      {/each}
    </ol>
  </section>
{/if}

<style>
  .list-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .map-link {
    padding: 0 14px;
    font-size: 15px;
  }

  ol {
    margin: 8px 0 0;
    padding: 0;
    list-style: none;
  }

  li {
    position: relative;
  }

  /* The line itself: one continuous bar through every node. */
  li::before {
    content: '';
    position: absolute;
    z-index: 1;
    left: 21px;
    top: 0;
    bottom: 0;
    width: 6px;
    background: var(--c);
    pointer-events: none;
  }

  li:first-child::before {
    top: 50%;
  }

  li:last-child::before {
    bottom: 50%;
  }

  a {
    position: relative;
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 48px;
    padding: 0 12px 0 50px;
    border-radius: var(--r-md);
    text-decoration: none;
    transition: background-color var(--t-out) var(--ease);
  }

  .node {
    position: absolute;
    z-index: 2;
    left: 16px;
    top: 50%;
    width: 16px;
    height: 16px;
    margin-top: -8px;
    border: 4px solid var(--c);
    border-radius: 50%;
    background: var(--bg);
  }

  .interchange .node {
    left: 14px;
    width: 20px;
    height: 20px;
    margin-top: -10px;
    border-color: var(--ink);
    background: var(--surface);
  }

  .name {
    flex: 1;
    min-width: 0;
  }

  .terminus .name {
    font-weight: 600;
  }

  .fav {
    color: var(--yellow);
  }

  .fav :global(svg) {
    fill: currentColor;
  }

  a[aria-current='page'] {
    background: var(--surface);
    box-shadow: inset 0 0 0 1px var(--hair-strong);
  }

  a[aria-current='page'] .name {
    font-weight: 600;
  }

  a[aria-current='page'] .node {
    background: var(--c);
  }

  @media (hover: hover) and (pointer: fine) {
    a:hover {
      background: var(--surface);
      transition-duration: var(--t-in);
    }
  }
</style>
