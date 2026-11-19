<script>
  import { store } from '../lib/store.svelte.js';
  import LineDots from './LineDots.svelte';

  let { current = null } = $props();

  const favourites = $derived(store.favourites.map((code) => store.stations.get(code)).filter(Boolean));
</script>

{#if favourites.length}
  <section aria-labelledby="favourites-title">
    <h2 id="favourites-title" class="label">As tuas estações</h2>
    <ul>
      {#each favourites as station (station.code)}
        <li>
          <a href="#/{station.code}" aria-current={station.code === current ? 'page' : undefined}>
            <LineDots lines={station.lines} />
            {station.name}
          </a>
        </li>
      {/each}
    </ul>
  </section>
{/if}

<style>
  ul {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin: 8px 0 0;
    padding: 0;
    list-style: none;
  }

  a {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-height: 44px;
    padding: 0 14px;
    border: 1px solid var(--hair);
    border-radius: 999px;
    background: var(--surface);
    font-weight: 500;
    text-decoration: none;
    transition:
      border-color var(--t-out) var(--ease),
      transform var(--t-out) var(--ease);
  }

  a[aria-current='page'] {
    border-color: var(--ink);
  }

  @media (hover: hover) and (pointer: fine) {
    a:hover {
      border-color: var(--hair-strong);
      transform: translateY(-1px);
      transition-duration: var(--t-in);
    }
  }
</style>
