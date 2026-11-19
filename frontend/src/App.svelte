<script>
  import { onMount } from 'svelte';
  import { store } from './lib/store.svelte.js';
  import { router } from './lib/router.svelte.js';
  import { lineFromSlug } from './lib/lines.js';
  import { onReconnect } from './lib/connection.svelte.js';
  import { push } from './lib/push.svelte.js';
  import TopBar from './components/TopBar.svelte';
  import OfflineBanner from './components/OfflineBanner.svelte';
  import Favourites from './components/Favourites.svelte';
  import LinePicker from './components/LinePicker.svelte';
  import StationList from './components/StationList.svelte';
  import StationView from './components/StationView.svelte';
  import LineMap from './components/LineMap.svelte';
  import EmptyState from './components/EmptyState.svelte';
  import AboutPage from './components/AboutPage.svelte';
  import PrivacyPage from './components/PrivacyPage.svelte';
  import SettingsPage from './components/SettingsPage.svelte';
  import NotificationsPage from './components/NotificationsPage.svelte';
  import InstallBanner from './components/InstallBanner.svelte';

  const station = $derived(router.code ? store.stations.get(router.code) : null);
  const mapLine = $derived(router.page === 'linha' ? lineFromSlug(router.line) : null);
  // On a phone the rail and the detail are two screens; on wider screens both show.
  const view = $derived(station || router.page ? 'detail' : 'home');

  // Keep the station list on a line that serves the open station, or on the line being drawn.
  $effect(() => {
    if (station && !station.lines.includes(store.line)) store.setLine(station.lines[0]);
    if (mapLine && store.line !== mapLine) store.setLine(mapLine);
  });

  $effect(() => {
    if (station) store.remember(station.code);
    document.title = station ? `${station.name} · Cais` : 'Cais';
  });

  $effect(() => {
    view;
    if (matchMedia('(max-width: 819px)').matches) scrollTo(0, 0);
  });

  // A station that stops being a favourite takes its alerts with it.
  let knownFavourites = store.favourites;
  $effect(() => {
    if (store.favourites !== knownFavourites) {
      knownFavourites = store.favourites;
      push.resync();
    }
  });

  onMount(() => {
    push.init();
    store.loadStatus();
    store.loadDestinations();
    const timer = setInterval(() => store.loadStatus(), 60000);
    const stop = onReconnect(() => store.loadStatus());
    return () => {
      clearInterval(timer);
      stop();
    };
  });
</script>

<a class="skip-link" href="#detail">Saltar para os tempos</a>

<div class="app" data-view={view}>
  <TopBar />
  <OfflineBanner />

  <main class="shell">
    <aside class="rail" aria-label="Linhas e estações">
      <InstallBanner />
      <Favourites current={station?.code} />
      <LinePicker />
      <StationList current={station?.code} />
      <footer class="note">
        <p>App não oficial, sem ligação ao Metropolitano de Lisboa.</p>
        <p><a href="#/sobre">Sobre</a> · <a href="#/privacidade">Privacidade</a></p>
      </footer>
    </aside>

    <section id="detail" class="detail" tabindex="-1" aria-label="Tempos de espera">
      {#if router.page === 'sobre'}
        <AboutPage />
      {:else if router.page === 'privacidade'}
        <PrivacyPage />
      {:else if router.page === 'definicoes'}
        <SettingsPage />
      {:else if router.page === 'notificacoes'}
        <NotificationsPage />
      {:else if mapLine}
        {#key mapLine}
          <LineMap lineId={mapLine} />
        {/key}
      {:else if station}
        {#key station.code}
          <StationView {station} />
        {/key}
      {:else if router.code}
        <div class="failed">
          <h1>Esta estação não existe.</h1>
          <p>O endereço pode estar incompleto. Escolhe uma estação da lista.</p>
        </div>
      {:else}
        <EmptyState />
      {/if}
    </section>
  </main>
</div>


<style>
  .app {
    display: grid;
    grid-template-rows: auto auto 1fr;
    min-height: 100dvh;
  }

  .shell {
    display: grid;
    grid-template-columns: 380px 1fr;
    min-height: 0;
  }

  .rail {
    display: grid;
    align-content: start;
    gap: 28px;
    padding: 24px 20px 32px 24px;
    border-right: 1px solid var(--hair);
  }

  .detail {
    min-width: 0;
    padding: 32px 40px 48px;
  }

  .detail:focus {
    outline: none;
  }

  .note {
    display: grid;
    gap: 4px;
    font-size: 13px;
    color: var(--muted);
  }

  .note a {
    display: inline-block;
    padding: 4px 0;
    color: var(--ink-2);
  }

  .failed {
    display: grid;
    justify-items: start;
    gap: 12px;
    max-width: 560px;
    padding-top: 8vh;
  }

  .failed h1 {
    font-size: clamp(26px, 3vw, 34px);
    font-weight: 700;
    line-height: 1.15;
    letter-spacing: -0.02em;
  }

  .failed p {
    color: var(--ink-2);
  }

  /* Wide screens: the app fills the window and each column scrolls on its own. */
  @media (min-width: 820px) {
    .app {
      height: 100dvh;
    }

    .rail,
    .detail {
      overflow-y: auto;
      overscroll-behavior: contain;
    }
  }

  @media (min-width: 820px) and (max-width: 1099px) {
    .shell {
      grid-template-columns: 320px 1fr;
    }

    .detail {
      padding: 28px 28px 40px;
    }
  }

  /* Phones: one screen at a time. */
  @media (max-width: 819px) {
    .shell {
      grid-template-columns: 1fr;
    }

    .rail {
      padding: 20px 16px 32px;
      border-right: 0;
    }

    .detail {
      padding: 8px 16px 40px;
      padding-bottom: max(40px, env(safe-area-inset-bottom));
    }

    .app[data-view='home'] .detail,
    .app[data-view='detail'] .rail {
      display: none;
    }

    /* On a station or page the times matter more than the search. */
    .app[data-view='detail'] :global(.topbar__search) {
      display: none;
    }
  }
</style>
