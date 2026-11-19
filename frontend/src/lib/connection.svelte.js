// Whether live data can reach us: the device's own network state, plus what
// the last request to our API said. api.js reports every answer here.

class Connection {
  deviceOnline = $state(navigator.onLine);
  apiFailing = $state(false);

  // 'online' | 'offline' (no network on the device) | 'unavailable' (network, but no data)
  state = $derived(!this.deviceOnline ? 'offline' : this.apiFailing ? 'unavailable' : 'online');

  constructor() {
    addEventListener('online', () => (this.deviceOnline = true));
    addEventListener('offline', () => (this.deviceOnline = false));
  }

  report(ok) {
    this.apiFailing = !ok;
  }
}

export const connection = new Connection();

// Runs fn whenever the connection comes back, so screens refresh on their own.
export function onReconnect(fn) {
  addEventListener('online', fn);
  return () => removeEventListener('online', fn);
}
