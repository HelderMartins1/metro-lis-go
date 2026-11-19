import { connection } from './connection.svelte.js';

const BASE = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '');

export async function request(path, { method = 'GET', body, timeout = 10000 } = {}) {
  const ctrl = new AbortController();
  const timer = setTimeout(() => ctrl.abort(), timeout);
  try {
    const res = await fetch(`${BASE}/api/${path}`, {
      method,
      signal: ctrl.signal,
      headers: { Accept: 'application/json', ...(body ? { 'Content-Type': 'application/json' } : {}) },
      body: body ? JSON.stringify(body) : undefined,
    });
    const data = await res.json().catch(() => null);
    // A 4xx is our request's fault, not the service's: the connection is fine.
    connection.report(res.status < 500);
    if (!res.ok) throw new Error(data?.error || `HTTP ${res.status}`);
    return data;
  } catch (err) {
    if (err.name === 'AbortError' || err instanceof TypeError) connection.report(false);
    throw err;
  } finally {
    clearTimeout(timer);
  }
}

export const get = (path, options) => request(path, options);
