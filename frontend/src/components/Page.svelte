<script>
  import Icon from './Icon.svelte';

  let { title, children } = $props();
</script>

<article class="page">
  <a class="back" href="#/"><Icon name="back" size={18} /> Voltar</a>
  <h1>{title}</h1>
  {@render children()}
</article>

<style>
  .page {
    max-width: 680px;
  }

  .back {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    min-height: 44px;
    margin-left: -6px;
    padding-right: 8px;
    font-weight: 500;
    text-decoration: none;
  }

  h1 {
    margin: 8px 0 24px;
    font-size: clamp(30px, 3.5vw, 40px);
    font-weight: 700;
    line-height: 1.1;
    letter-spacing: -0.03em;
  }

  .page :global(h2) {
    margin: 32px 0 8px;
    font-size: 19px;
    font-weight: 600;
  }

  .page :global(p),
  .page :global(li) {
    color: var(--ink-2);
    line-height: 1.65;
  }

  .page :global(p + p),
  .page :global(ul + p) {
    margin-top: 12px;
  }

  .page :global(ul) {
    margin: 8px 0 0;
    padding-left: 20px;
  }

  .page :global(li + li) {
    margin-top: 6px;
  }

  .page :global(.notice) {
    padding: 16px 20px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-md);
    background: var(--surface);
    color: var(--ink);
  }

  .page :global(.actions) {
    margin-top: 24px;
  }

  .page :global(.meta) {
    margin-top: 40px;
    font-size: 14px;
    color: var(--muted);
  }
</style>
