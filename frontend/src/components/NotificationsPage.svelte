<script>
  import { push } from '../lib/push.svelte.js';
  import { store } from '../lib/store.svelte.js';
  import Icon from './Icon.svelte';
  import LineDots from './LineDots.svelte';
  import Page from './Page.svelte';

  const DAYS = [
    { n: 1, short: 'Seg', long: 'segunda' },
    { n: 2, short: 'Ter', long: 'terça' },
    { n: 3, short: 'Qua', long: 'quarta' },
    { n: 4, short: 'Qui', long: 'quinta' },
    { n: 5, short: 'Sex', long: 'sexta' },
    { n: 6, short: 'Sáb', long: 'sábado' },
    { n: 7, short: 'Dom', long: 'domingo' },
  ];

  const favourites = $derived(store.favourites.map((c) => store.stations.get(c)).filter(Boolean));
  const following = $derived(
    push.settings.trains
      .map((t) => ({ ...t, station: store.stations.get(t.station) }))
      .filter((t) => t.station),
  );

  // The add-reminder form, open for one station at a time.
  let adding = $state(null);
  let time = $state('08:00');
  let days = $state([1, 2, 3, 4, 5]);
  let formError = $state('');

  function openForm(code) {
    adding = code;
    time = '08:00';
    days = [1, 2, 3, 4, 5];
    formError = '';
  }

  async function saveReminder() {
    if (!/^\d{2}:\d{2}$/.test(time)) {
      formError = 'Escolhe a hora do lembrete.';
      return;
    }
    if (days.length === 0) {
      formError = 'Escolhe pelo menos um dia.';
      return;
    }
    await push.addReminder(adding, time, [...days].sort());
    adding = null;
  }

  function describeDays(list) {
    const key = [...list].sort().join();
    if (key === '1,2,3,4,5') return 'dias úteis';
    if (key === '6,7') return 'fins de semana';
    if (list.length === 7) return 'todos os dias';
    return DAYS.filter((d) => list.includes(d.n)).map((d) => d.short.toLowerCase()).join(', ');
  }
</script>

<Page title="Notificações">
  {#if !push.supported && push.needsInstall}
    <p class="notice">
      No iPhone, as notificações só funcionam com o Cais no ecrã principal (iOS 16.4 ou mais recente). Toca em
      Partilhar e depois em “Adicionar ao ecrã principal”, e abre o Cais a partir daí.
    </p>
  {:else if !push.supported}
    <p class="notice">
      Este browser não recebe notificações. No telemóvel, usa o Cais instalado no ecrã principal; no computador, o
      Chrome, o Edge, o Firefox ou o Safari.
    </p>
  {:else if push.serverReady === false}
    <p class="notice">As notificações não estão disponíveis neste momento. Tenta mais tarde.</p>
  {:else if push.serverReady === null}
    <p class="muted">A verificar…</p>
  {:else}
    <section class="master">
      <div>
        <h2>{push.enabled ? 'Notificações ligadas' : 'Notificações desligadas'}</h2>
        <p>
          {push.enabled
            ? 'Recebes avisos das tuas estações e dos comboios que segues, mesmo com a app fechada.'
            : 'Liga para receberes avisos de perturbações, lembretes com os próximos comboios e o aviso de que o teu comboio está a chegar.'}
        </p>
      </div>
      {#if push.enabled}
        <button class="btn" disabled={push.busy} onclick={() => push.disable()}>Desligar notificações</button>
      {:else}
        <button class="btn btn--primary" disabled={push.busy} onclick={() => push.enable()}>Ligar notificações</button>
      {/if}
    </section>
    {#if push.permission === 'denied'}
      <p class="notice">
        Bloqueaste as notificações do Cais. Para as receberes, autoriza-as nas definições do browser ou do telemóvel.
      </p>
    {/if}
  {/if}

  {#if push.error}<p class="error" role="alert">{push.error}</p>{/if}

  {#if push.enabled}
    <h2>As tuas estações</h2>
    {#if favourites.length === 0}
      <p>
        Ainda não tens estações guardadas. Abre uma estação e toca na estrela; depois escolhe aqui os avisos que queres
        dela.
      </p>
    {:else}
      <ul class="stations">
        {#each favourites as station (station.code)}
          {@const reminders = push.remindersFor(station.code)}
          <li class="station">
            <div class="station__head">
              <a href="#/{station.code}"><LineDots lines={station.lines} /> {station.name}</a>
            </div>

            <label class="switch">
              <span>
                Perturbações nas linhas
                <span class="hint">Quando uma linha desta estação fica condicionada ou volta ao normal.</span>
              </span>
              <input
                type="checkbox"
                role="switch"
                checked={push.wantsDisruptions(station.code)}
                onchange={(e) => push.setDisruptions(station.code, e.currentTarget.checked)}
              />
            </label>

            <div class="reminders">
              <p class="reminders__title">Lembretes com os próximos comboios</p>
              {#each reminders as r (r.id)}
                <div class="reminder">
                  <span><strong>{r.time}</strong> · {describeDays(r.days)}</span>
                  <button
                    class="icon-btn"
                    aria-label="Apagar lembrete das {r.time}"
                    title="Apagar lembrete"
                    onclick={() => push.removeReminder(r.id)}
                  >
                    <Icon name="trash" size={18} />
                  </button>
                </div>
              {/each}

              {#if adding === station.code}
                <form
                  class="form"
                  onsubmit={(e) => {
                    e.preventDefault();
                    saveReminder();
                  }}
                >
                  <label class="field">
                    <span>Hora</span>
                    <input type="time" bind:value={time} required />
                  </label>
                  <fieldset class="days">
                    <legend>Dias</legend>
                    {#each DAYS as d (d.n)}
                      <label class="day">
                        <input type="checkbox" value={d.n} bind:group={days} />
                        <span aria-hidden="true">{d.short}</span>
                        <span class="sr-only">{d.long}</span>
                      </label>
                    {/each}
                  </fieldset>
                  {#if formError}<p class="error">{formError}</p>{/if}
                  <div class="form__actions">
                    <button class="btn btn--primary" type="submit">Guardar lembrete</button>
                    <button class="btn" type="button" onclick={() => (adding = null)}>Cancelar</button>
                  </div>
                </form>
              {:else}
                <button class="btn add" onclick={() => openForm(station.code)}>
                  <Icon name="plus" size={18} /> Adicionar lembrete
                </button>
              {/if}
            </div>
          </li>
        {/each}
      </ul>
    {/if}

    <h2>Comboios que estás a seguir</h2>
    {#if following.length === 0}
      <p>
        Nenhum. Numa estação, toca no sino de um sentido e avisamos 2 minutos antes de o próximo comboio chegar.
      </p>
    {:else}
      <ul class="following">
        {#each following as t (t.id)}
          <li>
            <a href="#/{t.station.code}">Comboio {t.train} em {t.station.name}</a>
            <button class="btn" onclick={() => push.unfollow(t.id)}>Deixar de seguir</button>
          </li>
        {/each}
      </ul>
    {/if}
  {/if}
</Page>

<style>
  .muted {
    color: var(--muted);
  }

  .master {
    display: grid;
    gap: 16px;
    justify-items: start;
    padding: 20px;
    border: 1px solid var(--hair);
    border-radius: var(--r-lg);
    background: var(--surface);
  }

  .master h2 {
    margin: 0 0 4px;
  }

  .error {
    margin-top: 12px;
    color: var(--alert) !important;
    font-weight: 500;
  }

  .stations,
  .following {
    display: grid;
    gap: 12px;
    margin: 8px 0 0;
    padding: 0;
    list-style: none;
  }

  .station {
    display: grid;
    gap: 4px;
    padding: 16px 18px;
    border: 1px solid var(--hair);
    border-radius: var(--r-lg);
    background: var(--surface);
  }

  .station + .station {
    margin-top: 0;
  }

  .station__head a {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    min-height: 44px;
    font-size: 18px;
    font-weight: 600;
    color: var(--ink);
    text-decoration: none;
  }

  .switch {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 8px 0 12px;
    border-bottom: 1px solid var(--hair);
    color: var(--ink);
    font-weight: 500;
    cursor: pointer;
  }

  .hint {
    display: block;
    font-size: 13px;
    font-weight: 400;
    color: var(--muted);
  }

  /* A native checkbox drawn as a switch: still a checkbox to the keyboard and screen readers. */
  .switch input {
    flex: none;
    appearance: none;
    position: relative;
    width: 48px;
    height: 28px;
    margin: 0;
    border-radius: 14px;
    background: var(--hair-strong);
    cursor: pointer;
    transition: background-color var(--t-in) var(--ease);
  }

  .switch input::after {
    content: '';
    position: absolute;
    top: 3px;
    left: 3px;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 3px rgb(0 0 0 / 0.25);
    transition: transform 180ms var(--ease);
  }

  .switch input:checked {
    background: var(--ok);
  }

  .switch input:checked::after {
    transform: translateX(20px);
  }

  .reminders {
    display: grid;
    gap: 6px;
    padding-top: 10px;
  }

  .reminders__title {
    font-weight: 500;
    color: var(--ink) !important;
  }

  .reminder {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    font-variant-numeric: tabular-nums;
  }

  .add {
    justify-self: start;
    margin-top: 4px;
  }

  .form {
    display: grid;
    gap: 14px;
    margin-top: 8px;
    padding: 14px;
    border-radius: var(--r-md);
    background: var(--sunken);
  }

  .field {
    display: grid;
    gap: 6px;
    font-size: 14px;
    font-weight: 500;
  }

  .field input {
    height: 48px;
    padding: 0 12px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-sm);
    background: var(--surface);
    color: var(--ink);
    font: inherit;
    font-size: 16px;
  }

  .days {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin: 0;
    padding: 0;
    border: 0;
  }

  .days legend {
    width: 100%;
    margin-bottom: 6px;
    padding: 0;
    font-size: 14px;
    font-weight: 500;
  }

  .day {
    position: relative;
  }

  .day input {
    position: absolute;
    opacity: 0;
  }

  .day span[aria-hidden] {
    display: grid;
    place-items: center;
    min-width: 44px;
    height: 44px;
    padding: 0 8px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-sm);
    background: var(--surface);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
  }

  .day input:checked + span {
    border-color: var(--ink);
    background: var(--ink);
    color: var(--bg);
  }

  .day input:focus-visible + span {
    outline: 2px solid var(--focus);
    outline-offset: 2px;
  }

  .form__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .following li {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 8px 12px;
    padding: 12px 16px;
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface);
  }
</style>
