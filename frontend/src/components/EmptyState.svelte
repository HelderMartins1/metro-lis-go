<script>
  import { store } from '../lib/store.svelte.js';
</script>

<div class="empty">
  <h1>Escolhe uma estação.</h1>
  <p>
    Toca numa estação da lista ou pesquisa pelo nome. Carrega na estrela de uma estação para a teres
    sempre à mão.
  </p>
  {#if store.last && store.stations.get(store.last)}
    <a class="btn btn--primary" href="#/{store.last}">Abrir {store.stations.get(store.last).name}</a>
  {/if}
</div>

<style>
  .empty {
    display: grid;
    justify-items: start;
    gap: 12px;
    max-width: 520px;
    padding-top: 8vh;
  }

  h1 {
    font-size: clamp(30px, 3.5vw, 40px);
    font-weight: 700;
    line-height: 1.1;
    letter-spacing: -0.03em;
  }

  p {
    color: var(--ink-2);
  }

  .btn {
    margin-top: 8px;
  }
</style>
