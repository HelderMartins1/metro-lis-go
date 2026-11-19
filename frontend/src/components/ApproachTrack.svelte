<script>
  import { fade } from 'svelte/transition';

  // trains: [{ id, left }] with left = seconds until arrival, soonest first.
  let { trains } = $props();

  const SPAN = 12 * 60;

  // Read left to right, like the times above: the platform on the left (0),
  // trains further away to the right (1).
  const position = (left) => Math.min(Math.max(left / SPAN, 0), 1);
</script>

<div class="track" aria-hidden="true">
  <span class="rail"></span>
  {#each trains as train, i (train.id)}
    <span
      class="train"
      class:first={i === 0}
      class:far={train.left > SPAN}
      style:--p={position(train.left)}
      out:fade={{ duration: 200 }}
    >
      <span class="car"></span>
    </span>
  {/each}
  <span class="stop"></span>
</div>
<div class="scale" aria-hidden="true">
  <span>agora</span>
  <span>6 min</span>
  <span>12 min</span>
</div>

<style>
  .track {
    position: relative;
    height: 22px;
    margin: 16px 9px 0;
  }

  .rail {
    position: absolute;
    left: 0;
    right: 0;
    top: 50%;
    height: 2px;
    margin-top: -1px;
    border-radius: 1px;
    background: var(--hair-strong);
  }

  /* As wide as the track, so translateX(100%) carries a train the whole way. */
  .train {
    position: absolute;
    inset: 0;
    transform: translateX(calc(var(--p) * 100%));
    transition: transform 1s linear;
  }

  .car {
    position: absolute;
    left: 0;
    top: 50%;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--c);
    opacity: 0.55;
    translate: -50% -50%;
  }

  .first .car {
    width: 14px;
    height: 14px;
    opacity: 1;
    box-shadow: 0 0 0 3px var(--surface);
  }

  .far .car {
    opacity: 0.3;
  }

  .stop {
    position: absolute;
    left: 0;
    top: 50%;
    width: 18px;
    height: 18px;
    border: 4px solid var(--ink);
    border-radius: 50%;
    background: var(--surface);
    translate: -50% -50%;
  }

  .scale {
    display: flex;
    justify-content: space-between;
    margin-top: 6px;
    font-size: 13px;
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }

  @media (prefers-reduced-motion: reduce) {
    .train {
      transition: none;
    }
  }
</style>
