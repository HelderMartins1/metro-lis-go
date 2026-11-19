<script>
  import { clock } from '../lib/clock.svelte.js';
  import Icon from './Icon.svelte';
  import Logo from './Logo.svelte';
  import Search from './Search.svelte';

  const format = new Intl.DateTimeFormat('pt-PT', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    timeZone: 'Europe/Lisbon',
  });
  const time = $derived(format.format(clock.now));
</script>

<header class="topbar">
  <a class="brand" href="#/" aria-label="Cais, início">
    <Logo size={34} />
    <span class="brand__text">
      <span class="brand__name">Cais</span>
      <span class="brand__sub">Metro de Lisboa · não oficial</span>
    </span>
  </a>
  <div class="topbar__search"><Search /></div>
  <div class="topbar__end">
    <time class="clock" aria-label="Hora em Lisboa">{time}</time>
    <a class="icon-btn" href="#/notificacoes" aria-label="Notificações" title="Notificações">
      <Icon name="bell" />
    </a>
    <a class="icon-btn" href="#/definicoes" aria-label="Definições, tema e privacidade" title="Definições">
      <Icon name="settings" />
    </a>
  </div>
</header>

<style>
  .topbar {
    display: grid;
    grid-template-columns: 1fr minmax(240px, 440px) 1fr;
    align-items: center;
    gap: 16px;
    padding: 12px 24px;
    padding-top: max(12px, env(safe-area-inset-top));
    border-bottom: 1px solid var(--hair);
    background: var(--bg);
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 12px;
    justify-self: start;
    min-height: 44px;
    text-decoration: none;
  }

  .brand__text {
    display: flex;
    flex-direction: column;
    line-height: 1.15;
  }

  .brand__name {
    font-size: 20px;
    font-weight: 700;
    letter-spacing: -0.02em;
  }

  .brand__sub {
    font-size: 13px;
    color: var(--muted);
  }

  .topbar__end {
    display: flex;
    align-items: center;
    gap: 8px;
    justify-self: end;
  }

  .clock {
    margin-right: 4px;
    font-size: 20px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    letter-spacing: -0.01em;
  }

  @media (max-width: 819px) {
    .topbar {
      grid-template-columns: 1fr auto;
      padding: 12px 16px;
      padding-top: max(12px, env(safe-area-inset-top));
    }

    .topbar__search {
      grid-column: 1 / -1;
      grid-row: 2;
    }

    /* The phone already shows the time in its status bar. */
    .clock {
      display: none;
    }
  }
</style>
