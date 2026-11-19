<script>
  import { clock } from '../lib/clock.svelte.js';
  import { ARRIVING, clockTime, minutes } from '../lib/format.js';
  import { push } from '../lib/push.svelte.js';
  import ApproachTrack from './ApproachTrack.svelte';
  import Icon from './Icon.svelte';

  // The server warns 2 min ahead; a train closer than this is too late to follow.
  const FOLLOW_MIN = 150;

  // card: { key (destination id), dest, trains: [{ id, seconds }] };
  // fetchedAt: when those seconds were true; station: where we are.
  let { card, fetchedAt, station } = $props();

  const elapsed = $derived(Math.max(0, (clock.now - fetchedAt) / 1000));
  const trains = $derived(card.trains.map((t) => ({ id: t.id, left: t.seconds - elapsed })));
  const next = $derived(trains[0]);

  const canNotify = $derived((push.supported && push.serverReady !== false) || push.needsInstall);
  const alert = $derived(push.followed(station, card.key));
  // Still on this board: once it has passed, the alert has done its job.
  const followedTrain = $derived(alert && trains.find((t) => t.id === alert.train));
  const toFollow = $derived(trains.find((t) => t.left > FOLLOW_MIN));

  function toggleFollow() {
    if (push.needsInstall) {
      location.hash = '#/notificacoes';
    } else if (followedTrain) {
      push.unfollow(alert.id);
    } else if (toFollow) {
      push.follow(station, card.key, toFollow.id);
    }
  }
</script>

<article class="card">
  {#if canNotify && next}
    <button
      class="icon-btn bell"
      class:is-on={!!followedTrain}
      aria-pressed={!!followedTrain}
      disabled={push.busy || (!followedTrain && !toFollow && !push.needsInstall)}
      aria-label={followedTrain ? 'Deixar de seguir este comboio' : 'Avisar 2 min antes do próximo comboio'}
      title={followedTrain
        ? 'Deixar de seguir'
        : toFollow || push.needsInstall
          ? 'Avisar 2 min antes'
          : 'O próximo comboio já está a chegar'}
      onclick={toggleFollow}
    >
      <Icon name="bell" class={followedTrain ? 'filled' : ''} />
    </button>
  {/if}
  <p class="label">Sentido</p>
  <h3>{card.dest ?? 'Destino por identificar'}</h3>

  {#if !next}
    <p class="none">Sem previsão para este sentido.</p>
  {:else}
    <!-- Soonest on the left, then the next ones in order. -->
    <div class="times">
      <p class="next" class:now={next.left <= ARRIVING}>
        {#if next.left <= 0}
          <span class="now__text">No cais</span>
        {:else if next.left <= ARRIVING}
          <span class="now__text">A chegar</span>
        {:else}
          <span class="value">{clockTime(next.left)}</span>
          <span class="unit">min</span>
        {/if}
      </p>
      {#if trains.length > 1}
        <p class="later">
          <span class="sr-only">Seguintes:</span>
          {#each trains.slice(1) as train (train.id)}
            <span class="later__time">{minutes(train.left)}</span>
          {/each}
        </p>
      {/if}
    </div>
    <ApproachTrack {trains} />
    {#if followedTrain}
      <p class="following">
        <Icon name="bell" size={16} class="filled" />
        Avisamos 2 min antes do comboio que chega em {minutes(followedTrain.left)}.
      </p>
    {/if}
  {/if}
</article>

<style>
  .card {
    position: relative;
    display: flex;
    flex-direction: column;
    /* The track's trains are track-wide boxes moved by transform; keep them from widening the page. */
    overflow: hidden;
    padding: 20px 24px;
    border: 1px solid var(--hair);
    border-radius: var(--r-lg);
    background: var(--surface);
  }

  h3 {
    margin-top: 2px;
    font-size: 22px;
    font-weight: 600;
    letter-spacing: -0.01em;
    line-height: 1.2;
  }

  .times {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 4px 20px;
    margin-top: 12px;
  }

  .next {
    display: flex;
    align-items: baseline;
    gap: 6px;
    min-height: 56px;
  }

  .value {
    font-size: 52px;
    font-weight: 600;
    line-height: 1;
    letter-spacing: -0.03em;
    font-variant-numeric: tabular-nums;
  }

  .unit {
    font-weight: 500;
    color: var(--muted);
  }

  .now {
    align-items: center;
    gap: 12px;
  }

  .now::before {
    content: '';
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--c);
    animation: pulse 1.6s var(--ease) infinite;
  }

  .now__text {
    font-size: 32px;
    font-weight: 700;
    letter-spacing: -0.02em;
    line-height: 1;
  }

  .later {
    display: flex;
    gap: 16px;
  }

  .later__time {
    font-size: 20px;
    font-weight: 600;
    color: var(--ink-2);
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .later__time + .later__time {
    color: var(--muted);
  }

  .none {
    margin-top: 12px;
    color: var(--muted);
  }

  .bell {
    position: absolute;
    top: 12px;
    right: 12px;
    border-color: transparent;
    background: none;
  }

  .bell.is-on {
    color: var(--warn);
  }

  .card :global(.filled) {
    fill: currentColor;
  }

  .following {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 12px;
    font-size: 14px;
    color: var(--ink-2);
  }

  .following :global(svg) {
    flex: none;
    color: var(--warn);
  }

  @media (max-width: 819px) {
    .card {
      padding: 16px 18px;
    }

    h3 {
      font-size: 19px;
    }

    .times {
      margin-top: 8px;
    }

    .next {
      min-height: 48px;
    }

    .value {
      font-size: 44px;
    }

    .now__text {
      font-size: 28px;
    }

    .later__time {
      font-size: 18px;
    }
  }

  @keyframes pulse {
    0% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--c) 50%, transparent); }
    100% { box-shadow: 0 0 0 10px transparent; }
  }
</style>
