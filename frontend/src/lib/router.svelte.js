// Hash routes:
//   #/                 home
//   #/BC               a station, by its code
//   #/linha/azul       the trains on a line, drawn on its diagram
//   #/sobre, #/privacidade, #/definicoes, #/notificacoes
// Hashes without the leading slash (like the skip link's #detail) are not routes.

export const PAGES = ['sobre', 'privacidade', 'definicoes', 'notificacoes'];

const HOME = { page: null, code: null, line: null };

function parse(previous) {
  const hash = location.hash;
  if (hash && !hash.startsWith('#/')) return previous;
  const path = decodeURIComponent(hash.slice(2));
  if (!path) return HOME;

  const lower = path.toLowerCase();
  if (PAGES.includes(lower)) return { ...HOME, page: lower };
  if (lower.startsWith('linha/')) return { ...HOME, page: 'linha', line: lower.slice(6) };
  return { ...HOME, code: path.toUpperCase() };
}

class Router {
  #route = $state(parse(HOME));

  constructor() {
    addEventListener('hashchange', () => (this.#route = parse(this.#route)));
  }

  get code() {
    return this.#route.code;
  }

  get page() {
    return this.#route.page;
  }

  get line() {
    return this.#route.line;
  }

  replace(code) {
    history.replaceState(null, '', `#/${code}`);
    this.#route = parse(this.#route);
  }
}

export const router = new Router();
