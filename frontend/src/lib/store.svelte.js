import { get } from './api.js';
// The same file the Go server embeds: lines and stations need no connection.
import network from '../../../models/network.json';

const PREFIX = 'cais:';

function load(key, fallback) {
  try {
    const value = localStorage.getItem(PREFIX + key);
    return value == null ? fallback : JSON.parse(value);
  } catch {
    return fallback;
  }
}

function save(key, value) {
  try {
    localStorage.setItem(PREFIX + key, JSON.stringify(value));
  } catch {
    // Private mode or storage blocked: the app works, it just forgets.
  }
}

// code -> { code, name, lines: [line ids] }; interchanges list every line they serve.
function indexStations(lines) {
  const stations = new Map();
  for (const line of lines) {
    for (const s of line.stations) {
      const station = stations.get(s.code) ?? { code: s.code, name: s.name, lines: [] };
      station.lines.push(line.id);
      stations.set(s.code, station);
    }
  }
  return stations;
}

// The last times seen per station, kept for when the connection drops.
const BOARDS_KEY = 'boards:v1';
const BOARDS_KEPT = 12;

class Store {
  lines = network;
  status = $state.raw(null);
  statusFailed = $state(false);
  // destination id -> { name, lines: [line ids] }
  destinations = $state.raw(null);
  line = $state(network.some((l) => l.id === load('line', null)) ? load('line', null) : network[0].id);
  favourites = $state.raw(load('favourites', []));
  last = load('last', null);

  stations = indexStations(network);

  async loadStatus() {
    try {
      this.status = await get('lines');
      this.statusFailed = false;
    } catch {
      this.statusFailed = true;
    }
  }

  // Uses the saved list straight away, however old, and refreshes it once a day.
  async loadDestinations() {
    const cached = load('destinations:v1', null);
    if (cached) this.destinations = cached.data;
    if (cached && Date.now() - cached.at < 24 * 3600e3) return;
    try {
      const byLine = await get('destinations');
      const data = {};
      for (const [line, list] of Object.entries(byLine)) {
        for (const d of list ?? []) {
          (data[d.id_destino] ??= { name: d.nome_destino, lines: [] }).lines.push(line);
        }
      }
      this.destinations = data;
      save('destinations:v1', { at: Date.now(), data });
    } catch {
      setTimeout(() => this.loadDestinations(), 15000);
    }
  }

  // { entries, fetchedAt } from the last answer for this station, or null.
  lastBoard(code) {
    return load(BOARDS_KEY, {})[code] ?? null;
  }

  saveBoard(code, entries, fetchedAt) {
    const boards = load(BOARDS_KEY, {});
    boards[code] = { entries, fetchedAt };
    const newest = Object.entries(boards)
      .sort(([, a], [, b]) => b.fetchedAt - a.fetchedAt)
      .slice(0, BOARDS_KEPT);
    save(BOARDS_KEY, Object.fromEntries(newest));
  }

  setLine(id) {
    this.line = id;
    save('line', id);
  }

  isFavourite(code) {
    return this.favourites.includes(code);
  }

  toggleFavourite(code) {
    this.favourites = this.isFavourite(code)
      ? this.favourites.filter((c) => c !== code)
      : [...this.favourites, code];
    save('favourites', this.favourites);
  }

  remember(code) {
    this.last = code;
    save('last', code);
  }
}

export const store = new Store();
