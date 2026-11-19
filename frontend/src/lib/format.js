// Arrival times come from the API in seconds.

export const ARRIVING = 30;

export function clockTime(seconds) {
  const s = Math.max(0, Math.round(seconds));
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
}

export function minutes(seconds) {
  if (seconds <= ARRIVING) return 'a chegar';
  return `${Math.max(1, Math.round(seconds / 60))} min`;
}

export function ago(ms) {
  const s = Math.floor(ms / 1000);
  if (s < 5) return 'agora';
  if (s < 60) return `há ${s} s`;
  return `há ${Math.floor(s / 60)} min`;
}

export function normalize(text) {
  return text.normalize('NFD').replace(/\p{Diacritic}/gu, '').toLowerCase().trim();
}
