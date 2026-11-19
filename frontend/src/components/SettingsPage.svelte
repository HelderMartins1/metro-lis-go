<script>
  import { theme } from '../lib/theme.svelte.js';
  import Icon from './Icon.svelte';
  import Page from './Page.svelte';

  const THEMES = [
    { value: 'auto', label: 'Automático', hint: 'Segue o dispositivo', icon: 'contrast' },
    { value: 'light', label: 'Claro', icon: 'sun' },
    { value: 'dark', label: 'Escuro', icon: 'moon' },
  ];
</script>

<Page title="Definições">
  <fieldset class="themes">
    <legend>Aspeto</legend>
    {#each THEMES as option (option.value)}
      <label class="option">
        <input
          type="radio"
          name="theme"
          value={option.value}
          checked={theme.choice === option.value}
          onchange={() => theme.set(option.value)}
        />
        <Icon name={option.icon} />
        <span class="option__text">
          {option.label}
          {#if option.hint}<span class="option__hint">{option.hint}</span>{/if}
        </span>
        <span class="option__check" aria-hidden="true"><Icon name="check" size={18} /></span>
      </label>
    {/each}
  </fieldset>

  <h2>Mais</h2>
  <nav class="links" aria-label="Mais">
    <a href="#/notificacoes"><Icon name="bell" /> Notificações <Icon name="chevron" size={18} /></a>
    <a href="#/sobre"><Icon name="info" /> Sobre o Cais <Icon name="chevron" size={18} /></a>
    <a href="#/privacidade"><Icon name="lock" /> Política de privacidade <Icon name="chevron" size={18} /></a>
  </nav>
</Page>

<style>
  .themes {
    display: grid;
    gap: 8px;
    margin: 0;
    padding: 0;
    border: 0;
  }

  legend {
    margin-bottom: 8px;
    padding: 0;
    font-size: 19px;
    font-weight: 600;
  }

  .option {
    position: relative;
    display: flex;
    align-items: center;
    gap: 14px;
    min-height: 56px;
    padding: 8px 16px;
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface);
    cursor: pointer;
  }

  .option input {
    position: absolute;
    opacity: 0;
    pointer-events: none;
  }

  .option:has(input:checked) {
    border-color: var(--ink);
    box-shadow: inset 0 0 0 1px var(--ink);
  }

  .option:has(input:focus-visible) {
    outline: 2px solid var(--focus);
    outline-offset: 2px;
  }

  .option__text {
    display: grid;
    flex: 1;
    font-weight: 500;
    line-height: 1.25;
  }

  .option__hint {
    font-size: 13px;
    font-weight: 400;
    color: var(--muted);
  }

  .option__check {
    visibility: hidden;
  }

  .option:has(input:checked) .option__check {
    visibility: visible;
  }

  .links {
    display: grid;
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface);
  }

  .links a {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 14px;
    min-height: 56px;
    padding: 0 16px;
    font-weight: 500;
    text-decoration: none;
  }

  .links a + a {
    border-top: 1px solid var(--hair);
  }

  .links a :global(svg:last-child) {
    color: var(--muted);
  }
</style>
