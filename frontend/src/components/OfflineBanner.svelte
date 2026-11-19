<script>
  import { connection } from '../lib/connection.svelte.js';
  import Icon from './Icon.svelte';
</script>

<!-- Always in the page, so screen readers announce the message when it appears. -->
<div class="banner" role="status">
  {#if connection.state === 'offline'}
    <p data-state="offline">
      <Icon name="offline" size={18} />
      <span><strong>Estás offline.</strong> Mostramos os últimos dados guardados; atualizam quando a rede voltar.</span>
    </p>
  {:else if connection.state === 'unavailable'}
    <p data-state="unavailable">
      <Icon name="alert" size={18} />
      <span><strong>Sem dados do Metro.</strong> O serviço não está a responder; tentamos outra vez sozinhos.</span>
    </p>
  {/if}
</div>

<style>
  p {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 10px 24px;
    border-bottom: 1px solid var(--hair);
    font-size: 14px;
    line-height: 1.4;
  }

  p :global(svg) {
    flex: none;
    margin-top: 1px;
  }

  [data-state='offline'] {
    background: var(--sunken);
    color: var(--ink);
  }

  [data-state='unavailable'] {
    background: color-mix(in srgb, var(--warn) 12%, var(--bg));
    color: var(--ink);
  }

  [data-state='unavailable'] :global(svg) {
    color: var(--warn);
  }

  @media (max-width: 819px) {
    p {
      padding: 10px 16px;
    }
  }
</style>
