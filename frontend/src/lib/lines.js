// Display names and the key each line has in /api/lines.
// slug is the line's name in its #/linha/… address.
export const LINES = {
  blue: { name: 'Azul', statusKey: 'azul', slug: 'azul' },
  yellow: { name: 'Amarela', statusKey: 'amarela', slug: 'amarela' },
  green: { name: 'Verde', statusKey: 'verde', slug: 'verde' },
  red: { name: 'Vermelha', statusKey: 'vermelha', slug: 'vermelha' },
};

export function lineFromSlug(slug) {
  return Object.keys(LINES).find((id) => LINES[id].slug === slug || id === slug) ?? null;
}

export function lineStatus(status, failed, id) {
  if (!status) return { state: 'unknown', label: failed ? 'Sem dados' : 'A carregar' };
  const raw = String(status[LINES[id].statusKey] ?? '').trim();
  const value = raw.toLowerCase();
  if (value === 'normal' || value === 'ok') return { state: 'ok', label: 'Normal' };
  if (value === 'condicionada') return { state: 'warn', label: 'Condicionada' };
  if (!raw) return { state: 'unknown', label: 'Sem dados' };
  return { state: 'alert', label: raw.charAt(0).toUpperCase() + raw.slice(1) };
}
