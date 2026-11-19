<script>
  import { onMount } from 'svelte';
  import Icon from './Icon.svelte';
  import Logo from './Logo.svelte';

  const KEY = 'cais:install-dismissed';

  let show = $state(false);
  let ios = $state(false);
  let prompt = null;

  onMount(() => {
    const installed = matchMedia('(display-mode: standalone)').matches || navigator.standalone;
    let dismissed = false;
    try {
      dismissed = localStorage.getItem(KEY) === '1';
    } catch {}
    if (installed || dismissed) return;

    ios = /iphone|ipad|ipod/i.test(navigator.userAgent) && !/crios|fxios/i.test(navigator.userAgent);
    if (ios) {
      show = true;
      return;
    }
    const onPrompt = (event) => {
      event.preventDefault();
      prompt = event;
      show = true;
    };
    addEventListener('beforeinstallprompt', onPrompt);
    return () => removeEventListener('beforeinstallprompt', onPrompt);
  });

  async function install() {
    if (!prompt) return;
    prompt.prompt();
    const { outcome } = await prompt.userChoice;
    if (outcome === 'accepted') show = false;
    prompt = null;
  }

  function dismiss() {
    show = false;
    try {
      localStorage.setItem(KEY, '1');
    } catch {}
  }
</script>

{#if show}
  <aside class="banner" aria-label="Instalar a app">
    <Logo size={40} />
    <p>
      <strong>Instala o Cais</strong>
      {#if ios}
        <span>Toca em Partilhar e depois em “Adicionar ao ecrã principal”.</span>
      {:else}
        <span>Abre direto da tua página inicial, como uma app.</span>
      {/if}
    </p>
    {#if !ios}
      <button class="btn btn--primary" onclick={install}>Instalar app</button>
    {/if}
    <button class="icon-btn" aria-label="Fechar aviso de instalação" onclick={dismiss}>
      <Icon name="close" />
    </button>
  </aside>
{/if}

<style>
  /* A card at the top of the home screen: it never covers the times. */
  .banner {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 12px;
    padding: 12px 12px 12px 16px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-lg);
    background: var(--surface);
  }

  .banner .btn {
    grid-column: 1 / -1;
    grid-row: 2;
  }

  p {
    display: grid;
    font-size: 14px;
    line-height: 1.35;
  }

  span {
    color: var(--muted);
  }
</style>
