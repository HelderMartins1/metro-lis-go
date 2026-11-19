<script>
  import { LINES } from '../lib/lines.js';

  let { lines } = $props();
</script>

<span class="dots" role="img" aria-label={lines.map((id) => `Linha ${LINES[id].name}`).join(', ')}>
  {#each lines as id (id)}
    <span class="dot" data-line={id}></span>
  {/each}
</span>

<style>
  .dots {
    display: inline-flex;
    gap: 4px;
    flex: none;
  }
</style>
