// Push notifications: turning them on and off, and what this device asked for.
// The server (push/) decides when to send; this keeps it told.

import { request } from './api.js';
import { store } from './store.svelte.js';

const KEY = 'cais:push';
// A followed train is forgotten here after this, as on the server.
const TRAIN_LIFE = 60 * 60 * 1000;

const ios = /iphone|ipad|ipod/i.test(navigator.userAgent);
const installed = matchMedia('(display-mode: standalone)').matches || navigator.standalone === true;

function read() {
  try {
    const v = JSON.parse(localStorage.getItem(KEY));
    if (v && Array.isArray(v.disruptions)) return { disruptions: [], reminders: [], trains: [], ...v };
  } catch {}
  return { disruptions: [], reminders: [], trains: [] };
}

function base64ToBytes(base64) {
  const padded = (base64 + '='.repeat((4 - (base64.length % 4)) % 4)).replace(/-/g, '+').replace(/_/g, '/');
  return Uint8Array.from(atob(padded), (c) => c.charCodeAt(0));
}

class Push {
  // The browser can do Web Push at all.
  supported = 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window;
  // An iPhone only gets Web Push once the app is on the home screen (iOS 16.4+).
  needsInstall = ios && !installed;

  serverReady = $state(null); // null until asked; false when the server has no push
  permission = $state(typeof Notification === 'undefined' ? 'denied' : Notification.permission);
  enabled = $state(false);
  busy = $state(false);
  error = $state('');
  settings = $state(read());

  #publicKey = null;
  #subscription = null;

  async init() {
    if (!this.supported) return;
    try {
      this.#publicKey = (await request('push/key')).publicKey;
      this.serverReady = true;
    } catch {
      this.serverReady = false;
      return;
    }
    const reg = await navigator.serviceWorker.ready;
    this.#subscription = await reg.pushManager.getSubscription();
    this.enabled = !!this.#subscription && Notification.permission === 'granted';
    // Followed trains that have had their time are gone on the server too.
    this.#save({ trains: this.settings.trains.filter((t) => Date.now() - t.at < TRAIN_LIFE) });
    if (this.enabled) this.#sync().catch(() => {});
  }

  // Asks for permission (must run from a tap) and subscribes this device.
  async enable() {
    this.error = '';
    this.busy = true;
    try {
      this.permission = await Notification.requestPermission();
      if (this.permission !== 'granted') {
        this.error = 'Sem autorização para notificações. Podes dá-la nas definições do browser ou do telemóvel.';
        return false;
      }
      const reg = await navigator.serviceWorker.ready;
      this.#subscription =
        (await reg.pushManager.getSubscription()) ??
        (await reg.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: base64ToBytes(this.#publicKey) }));
      await this.#sync();
      this.enabled = true;
      return true;
    } catch {
      this.error = 'Não foi possível ligar as notificações. Tenta outra vez daqui a pouco.';
      return false;
    } finally {
      this.busy = false;
    }
  }

  async disable() {
    this.error = '';
    this.busy = true;
    try {
      const endpoint = this.#subscription?.endpoint;
      await this.#subscription?.unsubscribe();
      if (endpoint) await request('push/subscription/delete', { method: 'POST', body: { endpoint } }).catch(() => {});
      this.#subscription = null;
      this.enabled = false;
      this.#save({ trains: [] });
    } finally {
      this.busy = false;
    }
  }

  // Disruption alerts only make sense for favourites; the list follows them.
  wantsDisruptions(code) {
    return this.settings.disruptions.includes(code);
  }

  async setDisruptions(code, on) {
    const list = this.settings.disruptions.filter((c) => c !== code);
    this.#save({ disruptions: on ? [...list, code] : list });
    await this.#syncOrReport();
  }

  remindersFor(code) {
    return this.settings.reminders.filter((r) => r.station === code);
  }

  async addReminder(station, time, days) {
    const id = crypto.randomUUID?.() ?? String(Date.now());
    this.#save({ reminders: [...this.settings.reminders, { id, station, time, days }] });
    await this.#syncOrReport();
  }

  async removeReminder(id) {
    this.#save({ reminders: this.settings.reminders.filter((r) => r.id !== id) });
    await this.#syncOrReport();
  }

  // The followed train on this platform, if any.
  followed(station, destination) {
    return this.settings.trains.find((t) => t.station === station && t.destination === destination) ?? null;
  }

  async follow(station, destination, train) {
    this.error = '';
    if (!this.enabled && !(await this.enable())) return;
    try {
      const alert = await request('push/trains', {
        method: 'POST',
        body: { endpoint: this.#subscription.endpoint, station, destination, train },
      });
      const others = this.settings.trains.filter((t) => !(t.station === station && t.destination === destination));
      this.#save({ trains: [...others, { id: alert.id, station, destination, train, at: Date.now() }] });
    } catch {
      this.error = 'Não foi possível seguir este comboio. Tenta outra vez.';
    }
  }

  async unfollow(id) {
    this.#save({ trains: this.settings.trains.filter((t) => t.id !== id) });
    if (this.#subscription) {
      await request('push/trains/delete', { method: 'POST', body: { endpoint: this.#subscription.endpoint, id } }).catch(
        () => {},
      );
    }
  }

  // Call when the favourites change: alerts for a station that stopped being one go too.
  resync() {
    return this.#syncOrReport();
  }

  async #syncOrReport() {
    if (!this.enabled) return;
    try {
      await this.#sync();
      this.error = '';
    } catch {
      this.error = 'Não foi possível guardar no servidor. Tenta outra vez quando houver ligação.';
    }
  }

  // The server keeps a copy of what this device wants; this sends it the whole list.
  async #sync() {
    const favourites = new Set(store.favourites);
    await request('push/subscription', {
      method: 'PUT',
      body: {
        subscription: this.#subscription.toJSON(),
        disruptions: this.settings.disruptions.filter((c) => favourites.has(c)),
        reminders: this.settings.reminders.filter((r) => favourites.has(r.station)),
      },
    });
  }

  #save(patch) {
    this.settings = { ...this.settings, ...patch };
    try {
      localStorage.setItem(KEY, JSON.stringify(this.settings));
    } catch {}
  }
}

export const push = new Push();
