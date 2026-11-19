<script>
  import { store } from '../lib/store.svelte.js';
  import { normalize } from '../lib/format.js';
  import Icon from './Icon.svelte';
  import LineDots from './LineDots.svelte';

  let query = $state('');
  let open = $state(false);
  let active = $state(0);
  let input;

  const results = $derived.by(() => {
    const q = normalize(query);
    if (!q) return [];
    return [...store.stations.values()]
      .filter((s) => normalize(s.name).includes(q) || s.code.toLowerCase() === q)
      .sort(
        (a, b) =>
          normalize(b.name).startsWith(q) - normalize(a.name).startsWith(q) ||
          a.name.localeCompare(b.name, 'pt'),
      )
      .slice(0, 8);
  });

  function choose(station) {
    location.hash = `#/${station.code}`;
    query = '';
    open = false;
    input.blur();
  }

  function onkeydown(event) {
    if (event.key === 'ArrowDown') {
      event.preventDefault();
      active = Math.min(active + 1, results.length - 1);
    } else if (event.key === 'ArrowUp') {
      event.preventDefault();
      active = Math.max(active - 1, 0);
    } else if (event.key === 'Enter' && results[active]) {
      event.preventDefault();
      choose(results[active]);
    } else if (event.key === 'Escape') {
      query = '';
      open = false;
    }
  }

  // "/" jumps to the search field from anywhere.
  $effect(() => {
    const focusSearch = (event) => {
      const typing = /^(input|textarea|select)$/i.test(document.activeElement?.tagName ?? '');
      if (event.key === '/' && !typing) {
        event.preventDefault();
        input.focus();
      }
    };
    addEventListener('keydown', focusSearch);
    return () => removeEventListener('keydown', focusSearch);
  });
</script>

<div class="search" role="search">
  <label class="sr-only" for="search">Pesquisar estação</label>
  <Icon name="search" size={18} class="search__icon" />
  <input
    bind:this={input}
    bind:value={query}
    id="search"
    type="search"
    placeholder="Pesquisar estação"
    autocomplete="off"
    spellcheck="false"
    role="combobox"
    aria-controls="search-results"
    aria-expanded={open && query !== ''}
    aria-activedescendant={open && results[active] ? `option-${results[active].code}` : undefined}
    oninput={() => {
      open = true;
      active = 0;
    }}
    onfocus={() => (open = true)}
    onblur={() => (open = false)}
    {onkeydown}
  />
  <kbd aria-hidden="true">/</kbd>

  {#if open && query}
    <ul id="search-results" role="listbox" aria-label="Estações encontradas">
      {#each results as station, i (station.code)}
        <li
          id="option-{station.code}"
          role="option"
          aria-selected={i === active}
          onpointerdown={(event) => {
            event.preventDefault();
            choose(station);
          }}
          onpointerenter={() => (active = i)}
        >
          <span>{station.name}</span>
          <LineDots lines={station.lines} />
        </li>
      {:else}
        <li class="none">Nenhuma estação com “{query}”.</li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .search {
    position: relative;
    width: 100%;
  }

  .search :global(.search__icon) {
    position: absolute;
    left: 14px;
    top: 50%;
    translate: 0 -50%;
    color: var(--muted);
    pointer-events: none;
  }

  input {
    width: 100%;
    height: 44px;
    padding: 0 44px 0 42px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-md);
    background: var(--surface);
    color: var(--ink);
    font: inherit;
    transition:
      border-color 140ms var(--ease),
      box-shadow 140ms var(--ease);
  }

  input::placeholder {
    color: var(--muted);
  }

  input:focus {
    outline: none;
    border-color: var(--focus);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--focus) 25%, transparent);
  }

  input::-webkit-search-cancel-button {
    display: none;
  }

  kbd {
    position: absolute;
    right: 10px;
    top: 50%;
    translate: 0 -50%;
    min-width: 24px;
    padding: 1px 6px;
    border: 1px solid var(--hair-strong);
    border-radius: 6px;
    color: var(--muted);
    font: 500 13px/1.4 var(--font);
    text-align: center;
  }

  input:focus ~ kbd {
    display: none;
  }

  /* The shortcut only means something with a keyboard. */
  @media (hover: none) {
    kbd {
      display: none;
    }

    input {
      padding-right: 16px;
    }
  }

  ul {
    position: absolute;
    top: calc(100% + 6px);
    left: 0;
    right: 0;
    z-index: 20;
    margin: 0;
    padding: 6px;
    list-style: none;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-md);
    background: var(--surface);
    box-shadow: 0 12px 32px rgb(0 0 0 / 0.12);
  }

  li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-height: 44px;
    padding: 0 12px;
    border-radius: var(--r-sm);
    cursor: pointer;
  }

  li[aria-selected='true'] {
    background: var(--sunken);
  }

  .none {
    color: var(--muted);
    cursor: default;
  }
</style>
